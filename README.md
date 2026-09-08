# TokMan - Ultimate AI Orchestration

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Architecture](https://img.shields.io/badge/Architecture-Single%20Binary%20(%3C30MB)-success)](#architecture)
[![Memory Footprint](https://img.shields.io/badge/Resident%20RAM-%3C50MB-blue)](#performance)
[![License](https://img.shields.io/badge/License-Proprietary%20%2F%20Internal-red)](#)
[![Status](https://img.shields.io/badge/State-Modules%201%2C%202%20%26%203%20Complete%20(Gate%203%20Passed)-brightgreen)](#completed-features)

**TokMan** is a self-contained, enterprise-grade AI orchestration gateway, traffic shaper, stateful stream filter, and capability pool router. Engineered in pure native Go, TokMan compiles into a **single, standalone static binary** (<30MB) that boots in under 50 milliseconds and consumes under 50MB resident RAM.

TokMan operates in two distinct verticals out of the same binary:
1. **Desktop GUI Application** (Native Fyne v2 visual control plane with live telemetry, cache inspection, and pool management).
2. **Headless Server Daemon** (High-throughput CLI gateway for VPS, cloud instances, Docker containers, and edge appliances).

---

## Table of Contents

- [Architecture & Design Principles](#architecture--design-principles)
- [Completed Features (Done Thus Far)](#completed-features-done-thus-far)
- [Roadmap & Upcoming Pipeline (Modules 3–8)](#roadmap--upcoming-pipeline-modules-38)
- [Usage Instructions: Standalone Binary](#usage-instructions-standalone-binary)
  - [Configuration (`.env`)](#configuration-env)
  - [Vertical 1: Desktop GUI Mode](#vertical-1-desktop-gui-mode)
  - [Vertical 2: Headless CLI Daemon Mode](#vertical-2-headless-cli-daemon-mode)
  - [Production Systemd Service Setup](#production-systemd-service-setup)
- [API Reference & cURL Usage](#api-reference--curl-usage)
- [Building from Source (Developers Only)](#building-from-source-developers-only)
- [Automated Testing Suite](#automated-testing-suite)

---

## Architecture & Design Principles

```
+-------------------------------------------------------------------------+
|                    TokMan - Single Native Binary                        |
|                                                                         |
|  +--------------------+   +---------------------+   +----------------+  |
|  |  Fyne Desktop GUI  |   |  HTTP/SSE Gateway   |   |   Admin WebUI  |  |
|  |  (Visual Controls) |   |  (:8000 /v1/chat)   |   |   (/admin)     |  |
|  +---------+----------+   +----------+----------+   +--------+-------+  |
|            |                         |                       |          |
|  +---------v-------------------------v-----------------------v-------+  |
|  |                   Zero-Trust Security & Interceptor               |  |
|  |  - Constant-time auth verification   - 10MB payload clamp         |  |
|  |  - Restricted CORS origin matching   - Sensitive credential mask  |  |
|  +-----------------------------------+-------------------------------+  |
|                                      |                                  |
|  +-----------------------------------v-------------------------------+  |
|  |               Stateful Traffic Shaper & Context Guard             |  |
|  |  - P0-P3 Fair-Share Leaky Bucket     - Stateful <think> SSE Filter|  |
|  |  - Asymmetric Sliding-Window Pruning - Antiban Jitter Engine      |  |
|  +-----------------------------------+-------------------------------+  |
|                                      |                                  |
|  +-----------------------------------v-------------------------------+  |
|  |                     Local Persistence & Cache                     |  |
|  |  - Thread-Safe LRU Completion Cache (Sub-1ms instant hits)        |  |
|  |  - Embedded Pure-Go SQLite Database (Audit & telemetry store)     |  |
|  +-----------------------------------+-------------------------------+  |
|                                      |                                  |
|  +-----------------------------------v-------------------------------+  |
|  |             Canonical Registry (14 Capability Pools)              |  |
|  |    pool/auto | pool/coding | pool/research | pool/vision | ...     |  |
|  +-----------------------------------+-------------------------------+  |
+--------------------------------------|----------------------------------+
                                       | HTTPS Outbound (TLS)
                                       v
                     [ Upstream Providers (Groq / AI APIs) ]
```

- **Zero Runtime Dependencies:** No Python interpreters, Node.js runtimes, npm packages, or external database engines required at runtime.
- **Embedded Persistence:** Embedded pure Go SQLite database (`data/tokman.db`) handles structured audit logs, token counts, and cost telemetry.
- **In-Memory Sub-Millisecond Cache:** Thread-safe LRU completion cache yields zero-latency, zero-cost responses for repeated identical prompts.
- **Context Guard:** Strips and quarantines reasoning thinking blocks (`<think>...</think>`) in real time across server-sent event (SSE) streams.

---

## Completed Features (Done Thus Far)

### 1. Foundational Infrastructure & Testing Framework (Module 0)
- Native Go automated test suite covering unit, integration, and end-to-end proxy behavior.
- Zero-credential offline mock upstream server (`tests/mocks/mock_upstream.go`) simulating valid completions, SSE streaming chunks, and provider HTTP 429 backoff scenarios.
- Modular test orchestrators for Linux (`./tests/run_tests.sh`) and Windows PowerShell (`./tests/run_tests.ps1`).

### 2. Core Routing & State Persistence (Module 1 - Gate 1 Verified)
- High-throughput OpenAI-compatible HTTP Gateway listening on `:8000` (`/v1/chat/completions`, `/v1/models`).
- Embedded SQLite storage engine (`backend/storage/db.go`) recording full spend audits, provider response times, prompt tokens, and completion tokens.
- Thread-safe, in-memory 24-hour LRU completion cache (`backend/storage/cache.go`) caching deterministic queries with automatic TTL eviction.
- Upstream Groq Provider integration supporting `llama-3.3-70b-versatile` and `deepseek-r1-distill-llama-70b`.
- Operational endpoints: `/health/readiness`, `/health/liveness`, and embedded `/admin` web interface.

### 3. Traffic Shaper, Context Guard & Interceptor (Module 2 - Gate 2 Verified)
- **Multi-Channel Identity Normalizer:** Extracts and unifies user identifiers from API Keys, Bearer tokens, Telegram User IDs, Discord IDs, and remote client IPs.
- **Leaky-Bucket Rate Limiter with 4 Priority Classes:**
  - **P0 (Default / Free Tier):** 12 RPM, burst capacity 2.
  - **P1 (IDE / Cursor Interactive):** 60 RPM, burst capacity 10.
  - **P2 (Deep Research Batch):** 5 RPM, burst capacity 1.
  - **P3 (Media & Generative Pipeline):** 2 RPM, burst capacity 1.
  - Automatic HTTP 429 handling with RFC-compliant `Retry-After` headers and remaining quota metrics.
- **Context Guard (Stateful Streaming Filter):** Intercepts streaming Server-Sent Events (SSE) and quarantines raw `<think>...</think>` chain-of-thought tokens on the fly without breaking SSE JSON boundaries.
- **Dynamic Asymmetric Sliding-Window Pruner:** Protects against context-window exhaustion by automatically trimming historical non-system messages while preserving initial system prompts and the latest conversational turns.
- **Antiban Domain Jitter Engine:** Enforces a mandatory >=150ms inter-request delay floor with isolated pseudo-random jitter.

### 4. Security Hardening & Audit Remediation (100% Implemented)
- **CORS Protection:** Replaced wildcard headers with strict localhost/loopback origin matching (`localhost:*`, `127.0.0.1:*`, `[::1]:*`) and custom whitelisting via `CORS_ALLOWED_ORIGINS` with `Vary: Origin`.
- **Side-Channel Timing Protection:** Hardened master key authentication using `crypto/subtle.ConstantTimeCompare`.
- **Request Entity Safeguard:** Hard-clamped request payloads to 10MB via `http.MaxBytesReader`, immediately rejecting oversized payloads with HTTP 413.
- **Credential & Upstream Error Masking:** Sanitized error responses returned to clients while safely logging masked traces server-side.
- **Fair-Share Priority Isolation:** Maps rate-limiting priority strictly server-side from `X-Channel`, ignoring client-forged `X-Priority` headers unless `X-Tokman-Internal: true`.
- **Bounded Memory:** Automatic background eviction of stale client leaky buckets (`EvictStaleBuckets`).
- **Canonical Capability Registry:** Single source of truth (`backend/registry/registry.go`) defining all 14 canonical capability pools and metadata notes.

### 5. Autonomous Supervisor Engine & 14-Pool Federation (Module 3 - Gate 3 Verified)
- **Sub-40ms Hybrid Intent Classifier:** Two-tier classification combining Tier 0 lexical/regex heuristics (<1ms) with Tier 1 LLM 1-token arbitration prompt (<40ms deadline), dynamically routing prompts to `pool/agent-coding`, `pool/deep-reasoning`, `pool/web-research`, `pool/document-analysis`, or `pool/general`.
- **Master 144-Model Capability Catalog:** Comprehensive mapping of 14 capability pools with 3-tier intra-pool model fallback arrays ($\text{Primary Tier} \rightarrow \text{Backup Tier 1} \rightarrow \text{Backup Tier 2}$, $4 \times 4 \times 4$ models per pool).
- **Automated AST Security Critic Pass (`pool/security-tester`):** Inspects generated code for SQL injection, command execution, hardcoded keys, and unbounded reads with multi-turn refinement (`MAX_CRITIC_LOOPS=2`), early-exit on clean code, and security advisory tagging.
- **Ephemeral Scratchpad Envelope:** Isolates multi-step classification, routing decisions, and critic warnings from client chat history.
- **Live Orchestration Telemetry:** Emits real-time event updates and response headers (`X-Tokman-Pool`, `X-Tokman-Intent`, `X-Tokman-Classification-Ms`, `X-Tokman-Critic-Verdict`).

### 6. Dual-Vertical Native Desktop GUI (Fyne v2)
- Built-in multi-tab desktop dashboard:
  - **Endpoints Tab:** Active gateway status, listen port, and route inventory.
  - **Console Tab:** Interactive model inspector with detailed capability notes for all 14 canonical pools.
  - **Cache Tab:** Real-time LRU cache metrics, entry counts, and capacity gauges.
  - **Settings Tab:** Current runtime configurations and database paths.

---

## Roadmap & Upcoming Pipeline (Modules 4–8)

The remaining modules are staged in `docs/modules/` and will be implemented progressively according to the gated development roadmap:

| Module | Title | Primary Capabilities in the Pipeline |
| :--- | :--- | :--- |
| **Module 4** | **Local Search Grounding & Universal Chat Importer** | SearXNG native Go aggregator client with 1h memory cache, Universal Selective Chat Importer supporting 20+ archive formats (ChatGPT, Claude, TypingMind, LibreChat). |
| **Module 5** | **Tri-Tier Chat Bastions & Deterministic Admin ChatOps** | Outbound Telegram, Discord, and WhatsApp bot connectors, zero-LLM host telemetry (`/proc/meminfo`, Go runtime, RPM gauges), ephemeral key injection, PIN-protected Safe Mode (<50MB RAM clamp). |
| **Module 6** | **Deterministic Video Conductor & Multimedia Engines** | Single-worker FIFO channel queue, duration clamping (<=30s/<=60s), FFmpeg Ken Burns camera transformations (`zoompan`), Kokoro-82M TTS + Whisper alignment, headless Typst PDF & Marp slide rendering. |
| **Module 7** | **Stack Optimizer (`pool/stack-optimizer`) & Observability** | 7-day token burn telemetry aggregator, non-destructive SQLite proposal staging (48h TTL), zero-downtime hot-reloading upon operator approval. |
| **Module 8** | **Zero-Trust Perimeter, Mesh Ingress & Concentric Rings** | Tailscale overlay ingress (`tsnet`), multi-ring egress HTTP dialers (Rings 1–9), upstream 429 automated IP rotation, asymmetric egress routing. |

---

## Usage Instructions: Standalone Binary

> [!IMPORTANT]
> **No compiler or runtime required:** TokMan is distributed as a standalone pre-compiled executable (`tokman` on Linux/macOS, `tokman.exe` on Windows). You do not need to install Go, Python, or Node.js to run the binary.

### Configuration (`.env`)

TokMan automatically loads a `.env` configuration file placed in the same directory as the executable. Create a `.env` file before launching:

```bash
# TokMan Master Security Key (Used to authorize requests via Authorization or X-Master-Key)
LITELLM_MASTER_KEY=sk-master-internal-network-key

# Upstream LLM Provider API Keys
GROQ_API_KEY=gsk_your_groq_api_key_here

# Optional: Extra Allowed CORS Origins (comma-separated)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

---

### Vertical 1: Desktop GUI Mode

To launch the native graphical control panel with visual meters, cache inspectors, and capability pool browsers:

#### On Linux / macOS:
```bash
# Make sure binary has execution permissions
chmod +x ./tokman

# Launch GUI directly (or double-click the file in your desktop file manager)
./tokman
```

*(Optional)* Explicitly force GUI mode:
```bash
./tokman --gui
```

#### On Windows:
Double-click `tokman.exe` in Windows Explorer, or launch from PowerShell:
```powershell
.\tokman.exe
```

---

### Vertical 2: Headless CLI Daemon Mode

To run TokMan on a server, virtual machine, Docker container, or background terminal without initializing any desktop window:

#### Auto-Detection:
If launched on a Linux machine without a display server (i.e. neither `DISPLAY` nor `WAYLAND_DISPLAY` is set), TokMan **automatically detects headless mode** and boots straight into CLI daemon mode.

#### Explicit Headless Flag:
```bash
./tokman --headless
```

#### Custom Port and Database Storage:
```bash
./tokman --headless --port 9000 --db /var/lib/tokman/tokman.db
```

#### Available CLI Options:
| Flag | Default | Description |
| :--- | :--- | :--- |
| `--headless` | `false` | Run in headless server daemon mode (no GUI window initialized). |
| `--gui` | `false` | Force desktop GUI window mode. |
| `--port` | `8000` | HTTP gateway listening port. |
| `--db` | `data/tokman.db` | File path for embedded SQLite database persistence. |
| `--tab` | `0` | Default active tab in GUI mode (0: Endpoints, 1: Console, 2: Cache, 3: Settings). |

---

### Production Systemd Service Setup

To run TokMan as an unattended, auto-restarting background service on a Linux server:

1. Copy the standalone binary to `/usr/local/bin/tokman`:
   ```bash
   sudo cp tokman /usr/local/bin/tokman
   sudo chmod +x /usr/local/bin/tokman
   ```

2. Create directory for persistent data and configuration:
   ```bash
   sudo mkdir -p /etc/tokman /var/lib/tokman
   sudo cp .env /etc/tokman/.env
   ```

3. Create the systemd service file at `/etc/systemd/system/tokman.service`:
   ```ini
   [Unit]
   Description=TokMan - Ultimate AI Orchestration Gateway
   After=network.target

   [Service]
   Type=simple
   User=tokman
   Group=tokman
   WorkingDirectory=/var/lib/tokman
   EnvironmentFile=/etc/tokman/.env
   ExecStart=/usr/local/bin/tokman --headless --port 8000 --db /var/lib/tokman/tokman.db
   Restart=always
   RestartSec=5s
   LimitNOFILE=65535

   [Install]
   WantedBy=multi-user.target
   ```

4. Enable and start the service:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now tokman
   sudo systemctl status tokman
   ```

---

## API Reference & cURL Usage

Once the binary is running, it exposes OpenAI-compatible REST endpoints.

### 1. Health & Readiness
```bash
curl -i http://localhost:8000/health/readiness
```
```json
{
  "status": "ready",
  "service": "tokman",
  "cache_items": 0,
  "uptime": "1m22s"
}
```

### 2. Standard Chat Completion
Send a standard completion request using your master key:

```bash
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-master-internal-network-key" \
  -d '{
    "model": "llama-3.3-70b-versatile",
    "messages": [
      {"role": "system", "content": "You are a concise engineering assistant."},
      {"role": "user", "content": "Explain what an LRU cache is in one sentence."}
    ],
    "temperature": 0.2
  }'
```

### 3. Fair-Share Priority Channels
Use the `X-Channel` header to allocate traffic to priority classes:

- **IDE / Cursor Traffic (P1 - 60 RPM):**
  ```bash
  curl -X POST http://localhost:8000/v1/chat/completions \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer sk-master-internal-network-key" \
    -H "X-Channel: ide" \
    -d '{"model": "llama-3.3-70b-versatile", "messages": [{"role": "user", "content": "Refactor this function"}]}'
  ```

- **Batch Research Traffic (P2 - 5 RPM):**
  ```bash
  curl -X POST http://localhost:8000/v1/chat/completions \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer sk-master-internal-network-key" \
    -H "X-Channel: research" \
    -d '{"model": "llama-3.3-70b-versatile", "messages": [{"role": "user", "content": "Deep research query"}]}'
  ```

### 4. Cache Verification
Repeat an identical completion request. The second response will include the `X-Cache: HIT` header and return instantly in `<1ms` without consuming any upstream tokens:

```bash
curl -i -X POST http://localhost:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-master-internal-network-key" \
  -d '{"model": "llama-3.3-70b-versatile", "messages": [{"role": "user", "content": "Cache test query"}]}'
```

### 5. Admin Web Dashboard
Navigate to `http://localhost:8000/admin` in any web browser to view the built-in operational monitor and system statistics.

---

## Building from Source (Developers Only)

If you are developing TokMan and wish to compile the standalone binary from source code:

### Prerequisites:
- Go 1.24 or higher
- Cgo-compatible C compiler (`gcc` on Linux, MinGW-w64 on Windows)
- GUI libraries (Linux only): `libgtk-3-dev`, `libgl1-mesa-dev`, `xorg-dev`

### Compilation Scripts:
```bash
# On Linux / macOS:
./scripts/build.sh

# On Windows (PowerShell):
.\scripts\build.ps1
```

The compiled binary will be placed at `build/bin/tokman` (or `build/bin/tokman.exe`).

---

## Automated Testing Suite

TokMan enforces zero-regression testing. The automated test suite executes all Go unit, integration, and resiliency tests:

```bash
# Run the complete test suite
./tests/run_tests.sh --all

# Run specific module tests
./tests/run_tests.sh --module 1
./tests/run_tests.sh --module 2
```

On Windows:
```powershell
.\tests\run_tests.ps1 -All
```

---

## License

Proprietary / Internal Orchestration Architecture. All rights reserved.
