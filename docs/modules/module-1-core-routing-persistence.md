# Module 1: Core Routing, State Persistence & Host Hardening (Wails Production Binary)

## 1. Module Overview & Architectural Role

Module 1 establishes the foundational data plane of the **Distributed AI Gateway Mesh** as a **single, self-contained native binary (`tokman`) built with Wails v2 (Go + Web UI)**.

In production, `tokman` combines the high-performance HTTP gateway router, in-memory LRU completion cache, embedded SQLite persistence, and modern web desktop interface into one single executable.

This module is responsible for:
1. Receiving OpenAI-compatible chat completion requests (`/v1/chat/completions`) and health probes (`/health/readiness`, `/health/liveness`) on `:8000`.
2. Routing requests with dynamic failover across upstream model capability pools (bootstrapped with Groq `llama-3.3-70b-versatile` and `deepseek-r1-distill-llama-70b`).
3. Caching prompt completions in-memory with sub-1ms response times.
4. Logging every transaction, token count, latency metric, and spend record into an embedded SQLite database (`data/tokman.db`).
5. Running within a compact **< 50 MB RAM** footprint on host edge devices without requiring Docker, PostgreSQL, or Redis containers.

```text
                                [ CLIENT / LOCAL INGRESS ]
                                            │
                                            ▼ (HTTP :8000)
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                        TOKMAN NATIVE GO GATEWAY CORE (:8000)                           │
│  • OpenAI-Compatible Proxy (/v1/chat/completions, /health/readiness)                   │
│  • Capability Pool Router (Usage-based routing & Upstream Failover)                    │
│  • In-Memory Thread-Safe LRU Completion Cache (< 1ms latency)                          │
│  • Embedded SQLite Spend & Transaction Logger (data/tokman.db)                         │
└───────────────────────────────────────────┬────────────────────────────────────────────┘
                                            │ Outbound HTTPS
                                            ▼
                             [ Upstream Provider Endpoints ]
                             • Groq LPU (Llama 3.3 70B, DeepSeek R1)
                             • Cerebras, Gemini, Mistral
```

---

## 2. Resource Allocation Across Sizing Stages

Because `tokman` compiles directly to native Go and uses native OS WebKitGTK, memory consumption is a fraction of containerized deployments:

| Component | Stage 1: Ultra-Edge<br>(< 3GB RAM) | Stage 2: Standard Edge<br>(3–6GB RAM) | Stage 3: Power Node<br>(6–12GB RAM) | Stage 4: Workstation<br>(12–24GB RAM) | Stage 5: Uncapped<br>(> 24GB RAM) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Go Gateway Core** | 15 MB | 25 MB | 40 MB | 80 MB | Uncapped |
| **In-Memory LRU Cache** | 10 MB | 20 MB | 50 MB | 100 MB | 512 MB |
| **Embedded SQLite** | 2 MB | 5 MB | 10 MB | 20 MB | 50 MB |
| **Wails WebUI (WebKit)**| 20 MB | 30 MB | 40 MB | 60 MB | 100 MB |
| **Total Stack Baseline**| **~47 MB** | **~80 MB** | **~140 MB** | **~260 MB** | **Host Managed** |

---

## 3. Inherent Gaps in Base Specification (`docs/report.md`) & Resolutions

| # | Inherent Gap in `report.md` | Architectural Impact | Wails Single-Binary Engineering Resolution |
| :-: | :--- | :--- | :--- |
| **G-1.1** | **LiteLLM Database Migration Lockups** | When LiteLLM starts simultaneously with Postgres, LiteLLM attempts Prisma migrations before Postgres is ready, resulting in crash loops. | **Eliminated**: Embedded SQLite database initialized inside the Go process with zero external connection locks. |
| **G-1.2** | **PostgreSQL RAM Bloat on Edge** | Default Postgres `shared_buffers` and WAL checkpointing will exhaust RAM on 2GB Raspberry Pi devices. | **Eliminated**: Pure Go embedded SQLite (`modernc.org/sqlite` or `mattn/go-sqlite3`) operates with minimal RAM (< 5 MB) and direct filesystem WAL. |
| **G-1.3** | **Redis Container Overhead** | Running external Redis container consumes 60MB–128MB RAM just for completion caching. | **Eliminated**: Built-in thread-safe Go LRU cache (`sync.RWMutex` + doubly linked list) with sub-1ms access latency. |
| **G-1.4** | **Credential Exposure in Logs** | LiteLLM debug logs print raw API keys and Bearer tokens if an upstream call fails. | Native Go HTTP client sanitizes headers before logging errors; sensitive keys masked to `***`. |

---

## 4. Component Deliverables & Configuration

### A. Environment Configuration (`.env.example` & `.env`)
```bash
# Cluster Master Secret
LITELLM_MASTER_KEY=sk-master-internal-network-key

# Embedded Gateway Port
GATEWAY_PORT=8000

# Upstream Provider Keys
GROQ_API_KEY=gsk_replace_with_real_key
CEREBRAS_API_KEY=csk_replace_with_real_key
GEMINI_API_KEY=AIzaSy_replace_with_real_key

# Node Identity
NODE_NAME=Local-Hub
```

### B. Wails Project Configuration (`wails.json`)
```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "tokman",
  "outputfilename": "tokman",
  "frontend:dir": "frontend",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "author": {
    "name": "Distributed AI Gateway Mesh"
  }
}
```

### C. Native Go Application Core (`main.go` & `app.go`)
- `main.go`: Initializes Wails desktop application window, sets up Go context, boots the internal HTTP Gateway router on `:8000`, and binds frontend IPC events.
- `app.go`: Implements Go backend methods callable from the Wails frontend:
  - `GetHealth()`: Returns gateway health, uptime, and memory usage.
  - `GetModels()`: Returns list of available capability pools and models.
  - `SendTestPrompt()`: Tests upstream model connectivity.

### D. Native Go Gateway Server (`backend/gateway/server.go`)
- HTTP Handler implementing:
  - `GET /health/readiness` & `GET /health/liveness` -> HTTP 200 OK
  - `POST /v1/chat/completions` -> OpenAI-compatible chat completion handler.
- Dispatcher:
  - Matches `model: "pool/general"` to Groq `llama-3.3-70b-versatile`.
  - Matches `model: "pool/deep-reasoning"` to Groq `deepseek-r1-distill-llama-70b`.
  - Checks in-memory LRU cache before making network calls.
  - Asynchronously logs completion metadata (model, tokens, latency) to SQLite.

### E. Embedded Persistence & Cache (`backend/storage/`)
- `backend/storage/db.go`: SQLite schema creating `spend_logs` table (`id`, `timestamp`, `model`, `prompt_tokens`, `completion_tokens`, `latency_ms`).
- `backend/storage/cache.go`: Thread-safe LRU cache with configurable size (default: 1,000 completions) and TTL.

---

## 5. Testing & Verification Runbook

### Automated Unit & Integration Tests
Execute Go test suite:
```bash
go test ./backend/gateway/... ./backend/storage/... -v
```

### Manual Verification Gate 1
1. Build the production binary:
   ```bash
   wails build
   ```
2. Run the application:
   ```bash
   ./build/bin/tokman
   ```
3. Verify readiness probe:
   ```bash
   curl http://localhost:8000/health/readiness
   # Expected: {"status": "healthy", "pools": ["pool/general", "pool/deep-reasoning"]}
   ```
4. Verify Groq completion via `pool/general`:
   ```bash
   curl http://localhost:8000/v1/chat/completions \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer sk-master-internal-network-key" \
     -d '{"model": "pool/general", "messages": [{"role": "user", "content": "Say online"}]}'
   ```
5. Verify cache hit: Repeat identical query and confirm response latency is < 5ms.
6. Verify SQLite log entry:
   ```bash
   sqlite3 data/tokman.db "SELECT model, prompt_tokens, completion_tokens FROM spend_logs ORDER BY id DESC LIMIT 1;"
   ```
