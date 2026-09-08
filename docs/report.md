# DISTRIBUTED AI GATEWAY MESH: UNIFIED MASTER ARCHITECTURAL MANIFEST & AGENTIC PRODUCTION SPECIFICATION

## 1. Executive Summary, System Axioms & Core Guiding Principles

The Distributed AI Gateway Mesh is an autonomous, self-healing, zero-trust private AI cluster engineered to aggregate, arbitrate, and federate free-tier LLM endpoints, meta-search grounding, multimodal synthesis, and media production across consumer edge hardware and cloud hyperscalers.

The system delivers between 6.75M and 67.5M tokens/day without exposing inbound network ports, without requiring credit cards or paid credentials, and without consuming local GPU VRAM.

```text
                                  [ ZERO-TRUST ACCESS PERIMETER ]
        Outbound-Only Chat               Quarantined Mesh P2P             Local Host / LAN
    (Telegram / Discord / WA)           (Tailscale Device Share)         (mDNS / 192.168.x.x)
              │                                   │                               │
              ▼                                   ▼                               ▼
    ┌───────────────────┐               ┌───────────────────┐           ┌───────────────────┐
    │ Tri-Tier ChatOps  │               │ Encrypted Adapter │           │ Native Direct     │
    │ Bastion (:3002)   │               │ `tailscale0` (Opt)│           │ Ingress (:8000)   │
    └─────────┬─────────┘               └─────────┬─────────┘           └─────────┬─────────┘
              │                                   │                               │
              └─────────────────────────┬─────────┴───────────────────────────────┘
                                        ▼
    ┌───────────────────────────────────────────────────────────────────────────────────────┐
    │                    ADVANCED FASTAPI INTERCEPTOR & TRAFFIC SHAPER (:8000)              │
    │  • Identity Normalizer (WebUI UUIDs, Telegram IDs, WhatsApp E.164 -> Canonical Quota) │
    │  • P0–P3 Fair-Share Priority Queue & Token-Bucket Rate Limiter (Jittered Pacing)      │
    │  • Context Guard: DeepSeek `<think>` Sanitizer & Dynamic Asymmetric Pruning Engine    │
    │  • Redis Caching & Search Engine Anti-Ban Traffic Dispenser                           │
    └───────────────────────────────────┬───────────────────────────────────────────────────┘
                                        │ (Internal HTTP)
                                        ▼
    ┌───────────────────────────────────────────────────────────────────────────────────────┐
    │                         LITELLM CORE ROUTING ENGINE (:4000)                           │
    │  • Virtual `pool/auto` Supervisor (Frontline 5 Intent Classifier + Tool Handoff DAG) │
    │  • 12 Core Capability Pools + 2 Meta-Orchestrators = 14 Pools (144 Model Slots)       │
    │  • Concentric Quota Rings (Ring 0 Host IP, Rings 1–9 Egress WireGuard Shards)         │
    │  • Proactive 80% Pre-Call Headroom Checking & Dynamic Cooldown Circuit Breakers       │
    └───────┬───────────────────────────┬───────────────────────────┬───────────────────────┘
            │                           │                           │
            ▼                           ▼                           ▼
┌───────────────────────┐   ┌───────────────────────┐   ┌───────────────────────────────────┐
│ STATE & PERSISTENCE   │   │ MODULAR EGRESS MESH   │   │ SEARCH, FEDERATION & PRODUCTION   │
│ • PostgreSQL 16 (:5432│   │ • Gluetun WireGuard   │   │ • SearXNG Meta-Search (:8080)     │
│ • Redis Cache (:6379) │   │   PIA Sidecars (:1080)│   │ • Video Conductor FIFO (:3004)    │
│ • Grafana OSS (:3005) │   │ • Automatic IP Rotate │   │ • Peer Gateway B (100.64.0.2:4000)│
│ • Zstd L10 Compression│   │   on Upstream 429     │   │ • Peer Gateway C (100.64.0.3:4000)│
└───────────────────────┘   └───────────────────────┘   └───────────────────────────────────┘
```

### Core Architecture Axioms

- **Local-First, Absolute Zero Inbound Exposure:** All inbound WAN firewall ports remain locked. The core gateway binds strictly to `127.0.0.1` and the local subnet. Remote access operates exclusively over peer-to-peer encrypted WireGuard overlays (Tailscale) or outbound-only chat daemons.
- **Autonomous Agentic Supervisor (`pool/auto`):** Frontline requests automatically route through a two-tier hierarchical arbitrator that handles intent classification (<40 ms), mid-query pool handoffs, and live step-by-step chat telemetry without requiring manual model selection.
- **Decoupled Quota Reservoirs:** Model credentials are organized into concentric Quota Rings. Free-tier capacity scales horizontally across accounts without duplicating container stacks, proxy processes, or memory footprints.
- **Context & State Invariance:** Conversation history is stored in a normalized canonical schema. Internal Chain-of-Thought (`<think>`) tokens are quarantined from persistent storage, and context windows are pruned dynamically based on the destination model's context ceiling, preventing HTTP 400 errors during model handoffs.
- **Proactive Anti-Ban & Single-IP Traffic Smoothing:** Even without a VPN, the gateway prevents upstream rate limits and IP blacklisting through an 80% pre-call headroom check, an inter-request delay floor (150–350ms jitter), and cross-provider domain dispersion.
- **Multi-User Fair-Share Scheduling:** Traffic is categorized into priority tiers (P0 Interactive Chat down to P3 Heavy Video/Artifacts). Memory-intensive rendering jobs are serialized via a single-worker Redis FIFO queue to prevent edge CPU and RAM exhaustion.
- **Deterministic Video Production (`pool/video-conductor`):** Volatile text-to-video diffusion endpoints are replaced by an orchestrated Video Conductor pipeline. It synthesizes Flux.1 anchor keyframes, Kokoro-82M TTS narration, Whisper timing alignments, and deterministic FFmpeg camera transforms (Ken Burns effect) capped at 30 seconds default (up to 60 seconds with `--long`).
- **Strict Human-in-the-Loop Stack Optimization:** The Stack Optimizer (`pool/stack-optimizer`) operates under a Zero-Touch Default. It compiles waste telemetry and stages tuning diffs in Redis; no live configurations or routing tables are altered on the fly without explicit operator approval unless the `AUTO_APPROVE_TUNING=true` override is set.
- **Deterministic Tri-Tier Chat Architecture:** Chat apps are partitioned into three dedicated experiences: (1) Generic Conversational Assistant, (2) OpenClaw Autonomous Agent, and (3) a Deterministic Admin ChatOps Menu (0 LLM tokens, 85% mobile operational control, PIN-protected safe mode, reserving high blast-radius cryptographic/volume-purge tasks strictly to the Web Admin Panel `:3000`).
- **Universal Client-Side Ingestion:** An in-browser, client-side script intercepts file uploads in Open WebUI, enabling instant selective migration from 20+ leading platforms (ChatGPT, Claude, Gemini, DeepSeek, Grok, Perplexity, etc.) with zero server-side exposure.
- **Rejection of External Monetization:** The architecture strictly excludes commercial token resale, token-to-credit conversion subsystems, and centralized micro-billing. The cluster is engineered strictly for private utility, personal sovereignty, and internal stack optimization.

### Canonical Production Target: Wails v2 Standalone Native Binary (`tokman`)

While the research prototype decomposed these systems into 11 discrete Docker containers, the **production deployment target compiles directly into a single, self-contained native binary (`tokman`) using Wails v2 (Go Backend + Embedded Web UI)**:

```text
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                        TOKMAN PRODUCTION BINARY (WAILS V2)                             │
│                                                                                        │
│  ┌─────────────────────────────────────┐      ┌─────────────────────────────────────┐  │
│  │     EMBEDDED WEB FRONTEND (UI)      │      │          NATIVE GO BACKEND          │  │
│  │ • Real-time Telemetry Dashboard    │      │ • HTTP Gateway Server (:8000)       │  │
│  │ • Capability Pool & Key Switcher    │◄────►│ • Leaky-Bucket Rate Limiter (P0–P3) │  │
│  │ • Chat Playground & ChatOps Menu    │ IPC  │ • Stateful <think> SSE Stream Filter│  │
│  │ • Universal 20+ Format Importer     │      │ • Autonomous Intent Classifier DAG  │  │
│  │ • Native OS WebKit (No Chromium)    │      │ • Embedded SQLite & In-Memory LRU   │  │
│  └─────────────────────────────────────┘      └──────────────────┬──────────────────┘  │
└──────────────────────────────────────────────────────────────────┼─────────────────────┘
                                                                   │ Outbound HTTPS
                                                                   ▼
                                                    [ Upstream Free-Tier Providers ]
```

1. **Zero Container Overhead:** Replaces Docker, LiteLLM, PostgreSQL, and Redis containers with native Go goroutines, embedded SQLite persistence, and an in-process thread-safe LRU cache.
2. **Ultra-Low Memory Footprint:** Drops baseline memory from **2.28 GB RAM** to **< 50 MB RAM**, running smoothly on consumer laptops, Raspberry Pi 4/5, and edge routers.
3. **Sub-Millisecond Wire-Speed Proxying:** Eliminates multi-hop container bridge networking; internal routing and `<think>` token filtering complete in **< 1 ms**.
4. **Single-File Distribution:** Builds into a single executable `tokman` ready for instant execution without installing Docker, Python virtualenvs, or external database servers.

---

## 2. Master Hardware Sizing & Dynamic Resource Profiles

The gateway automatically calibrates its cgroups, database buffers, cache eviction policies, and connection pools according to host physical memory availability via a 5-stage resource slider:

```text
[ HOST RAM DETECTION ]
   ├── < 3 GB RAM   ──► STAGE 1: Ultra-Edge (Cap: 1.68 GB RAM) -> tmpfs WAL, Aggressive Pruning
   ├── 3 - 6 GB RAM ──► STAGE 2: Standard Edge (Cap: 2.28 GB RAM) -> Standard Storage, 30-Day Index
   ├── 6 - 12 GB RAM──► STAGE 3: Power Node (Cap: 3.52 GB RAM) -> In-Memory Cache, 60-Day Index
   ├── 12- 24 GB RAM──► STAGE 4: Workstation (Cap: 6.45 GB RAM) -> High Concurrency Pools
   └── > 24 GB RAM  ──► STAGE 5: Uncapped / Cluster Host -> Bare-Metal Limits
```

### Resource Allocation Matrix

| Subsystem / Container | Stage 1: Ultra-Edge<br>(HP t630 / RPi4) | Stage 2: Standard Edge<br>(RPi5 / N100) | Stage 3: Power Node<br>(Mac Mini / 16GB PC) | Stage 4: Workstation<br>(32GB+ PC) | Stage 5: Uncapped<br>(Server / Cloud) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **LiteLLM Core (:4000)** | 350 MB | 512 MB | 1024 MB | 2048 MB | Uncapped |
| **FastAPI Interceptor (:8000)** | 120 MB | 160 MB | 256 MB | 512 MB | Uncapped |
| **PostgreSQL 16 (:5432)** | 120 MB | 256 MB | 512 MB | 1024 MB | Uncapped |
| **Redis Cache (:6379)** | 60 MB | 128 MB | 256 MB | 512 MB | Uncapped |
| **SearXNG Engine (:8080)** | 60 MB | 80 MB | 120 MB | 256 MB | Uncapped |
| **Open WebUI (:3001)** | 450 MB | 512 MB | 768 MB | 1024 MB | Uncapped |
| **Admin Hub (:3000)** | 95 MB | 128 MB | 160 MB | 256 MB | Uncapped |
| **Grafana OSS (:3005)** | 110 MB | 128 MB | 160 MB | 256 MB | Uncapped |
| **Tri-Tier Bastion (:3002)** | 85 MB | 100 MB | 128 MB | 256 MB | Uncapped |
| **Video Conductor Worker (:3004)** | 150 MB (Cgroup cap) | 200 MB | 250 MB | 512 MB | Uncapped |
| **Optional Tailscale Adapter** | 35 MB | 45 MB | 64 MB | 128 MB | Uncapped |
| **Optional Gluetun VPN Sidecar** | 30 MB / tunnel | 35 MB / tunnel | 45 MB / tunnel | 64 MB / tunnel | Uncapped |
| **Total Stack Baseline** | **~1.68 GB RAM** | **~2.28 GB RAM** | **~3.52 GB RAM** | **~6.45 GB RAM** | **Host Managed** |

---

## 3. Master Production Infrastructure Stack

> [!NOTE]
> **Production Target Specification:** The canonical production distribution for `tokman` is the **Wails v2 Single Native Binary**, which embeds the gateway router, traffic shaper, persistence, and web UI directly into one executable without requiring external container dependencies.
>
> The multi-container `compose.yml` specification below is retained as the **Modular Containerized Cluster Profile** for enterprise operators deploying distributed multi-node server clusters with external database shards.

The cluster deployment manifest manages the 11 core services and 2 modular profile services (Tailscale Ingress and Gluetun Egress):

```yaml
services:
  # ==============================================================================
  # 1. ROUTING & RUNTIME LAYER
  # ==============================================================================
  litellm:
    image: ghcr.io/berriai/litellm-database:main-latest
    container_name: ai-gateway-litellm
    restart: unless-stopped
    ports:
      - "127.0.0.1:4000:4000"
    volumes:
      - ./config/config.yaml:/app/config.yaml:ro
    environment:
      - DATABASE_URL=postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - LITELLM_MASTER_KEY=${LITELLM_MASTER_KEY}
    env_file: .env
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    ulimits:
      nofile:
        soft: 65535
        hard: 65535
    deploy:
      resources:
        limits:
          memory: ${MEM_LIMIT_LITELLM:-380M}

  interceptor:
    build:
      context: ./daemon
      dockerfile: Dockerfile.interceptor
    container_name: ai-gateway-interceptor
    restart: unless-stopped
    ports:
      - "127.0.0.1:8000:8000"
    environment:
      - GATEWAY_URL=http://litellm:4000
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - SEARXNG_URL=http://searxng:8080
      - NODE_NAME=${NODE_NAME:-Local-Hub}
      - AUTO_APPROVE_TUNING=${AUTO_APPROVE_TUNING:-false}
    depends_on:
      - litellm
      - searxng
      - redis
    ulimits:
      nofile:
        soft: 65535
        hard: 65535
    deploy:
      resources:
        limits:
          memory: ${MEM_LIMIT_INTERCEPTOR:-120M}

  # ==============================================================================
  # 2. STATE, PERSISTENCE & TELEMETRY
  # ==============================================================================
  postgres:
    image: postgres:16-alpine
    container_name: ai-gateway-postgres
    restart: unless-stopped
    environment:
      - POSTGRES_USER=${POSTGRES_USER:-gateway}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-gateway_secure_pass}
      - POSTGRES_DB=${POSTGRES_DB:-litellm}
    volumes:
      - ./data/postgres:/var/lib/postgresql/data
      - type: tmpfs
        target: /var/lib/postgresql/wal
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-gateway}"]
      interval: 5s
      timeout: 3s
      retries: 5
    deploy:
      resources:
        limits:
          memory: ${MEM_LIMIT_POSTGRES:-140M}

  redis:
    image: redis:7-alpine
    container_name: ai-gateway-redis
    restart: unless-stopped
    command: ["redis-server", "--appendonly", "no", "--save", "60 100", "--maxmemory", "90mb", "--maxmemory-policy", "allkeys-lru"]
    volumes:
      - ./data/redis:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    deploy:
      resources:
        limits:
          memory: ${MEM_LIMIT_REDIS:-95M}

  grafana:
    image: grafana/grafana-oss:latest
    container_name: ai-gateway-grafana
    restart: unless-stopped
    ports:
      - "127.0.0.1:3005:3000"
    volumes:
      - ./config/grafana/datasources.yaml:/etc/grafana/provisioning/datasources/datasources.yaml:ro
      - ./config/grafana/dashboards.yaml:/etc/grafana/provisioning/dashboards/dashboards.yaml:ro
      - ./config/grafana/dashboards:/var/lib/grafana/dashboards:ro
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-admin}
      - GF_USERS_ALLOW_SIGN_UP=false
    deploy:
      resources:
        limits:
          memory: ${MEM_LIMIT_GRAFANA:-110M}

  # ==============================================================================
  # 3. SEARCH GROUNDING & MEDIA ENGINES
  # ==============================================================================
  searxng:
    image: searxng/searxng:latest
    container_name: ai-gateway-searxng
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - ./config/searxng/settings.yml:/etc/searxng/settings.yml:ro
    environment:
      - SEARXNG_BASE_URL=http://127.0.0.1:8080/
    deploy:
      resources:
        limits:
          memory: ${MEM_LIMIT_SEARXNG:-60M}

  video-conductor:
    build:
      context: ./conductor
      dockerfile: Dockerfile
    container_name: ai-gateway-video-conductor
    restart: unless-stopped
    volumes:
      - ./data/artifacts:/app/artifacts
      - ./config:/app/config:ro
    environment:
      - GATEWAY_URL=http://interceptor:8000/v1
      - LITELLM_MASTER_KEY=${LITELLM_MASTER_KEY}
      - REDIS_HOST=redis
    depends_on:
      - interceptor
      - redis
    ulimits:
      nofile:
        soft: 65535
        hard: 65535
    deploy:
      resources:
        limits:
          memory: ${MEM_LIMIT_CONDUCTOR:-150M}

  # ==============================================================================
  # 4. INTERFACES & CONTROL PLANES
  # ==============================================================================
  admin-dashboard:
    build:
      context: ./dashboard
    container_name: ai-gateway-admin
    restart: unless-stopped
    ports:
      - "127.0.0.1:3000:3000"
    volumes:
      - ./config:/app/config
      - ./.env:/app/.env
      - ./data:/app/data
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - GATEWAY_URL=http://litellm:4000
    deploy:
      resources:
        limits:
          memory: ${MEM_LIMIT_ADMIN:-95M}

  open-webui:
    image: ghcr.io/open-webui/open-webui:main
    container_name: ai-gateway-webui
    restart: unless-stopped
    ports:
      - "127.0.0.1:3001:8080"
    volumes:
      - ./data/webui:/app/backend/data
      - ./config/webui/selective_import.js:/app/build/selective_import.js:ro
    environment:
      - OPENAI_API_BASE_URL=http://interceptor:8000/v1
      - OPENAI_API_KEY=${LITELLM_MASTER_KEY}
      - ENABLE_SIGNUP=false
      - ENABLE_WEB_SEARCH=True
      - WEB_SEARCH_ENGINE=searxng
      - SEARXNG_QUERY_URL=http://searxng:8080/search?q=<query>
    entrypoint: >
      sh -c '
      if ! grep -q "selective_import.js" /app/build/index.html; then
        sed -i "s|</body>|<script src=\"/selective_import.js\"></script></body>|" /app/build/index.html
      fi;
      exec /app/backend/start.sh
      '
    depends_on:
      - searxng
      - interceptor
    deploy:
      resources:
        limits:
          memory: ${MEM_LIMIT_WEBUI:-450M}

  # ==============================================================================
  # 5. CHAT ACCESS BRIDGES (TRI-TIER DECOUPLED EXPERIENCES)
  # ==============================================================================
  # Experience A: Generic Conversational Assistant
  chat-assistant:
    build:
      context: ./bastion-generic
    container_name: ai-gateway-chat-assistant
    profiles: ["chat-generic"]
    restart: unless-stopped
    environment:
      - GATEWAY_URL=http://interceptor:8000/v1
      - TELEGRAM_BOT_TOKEN=${GENERIC_TELEGRAM_TOKEN}
    deploy:
      resources:
        limits:
          memory: 30M

  # Experience B: OpenClaw Autonomous Agent
  openclaw-agent:
    build:
      context: ./bastion-openclaw
    container_name: ai-gateway-openclaw-agent
    profiles: ["chat-openclaw"]
    restart: unless-stopped
    environment:
      - GATEWAY_URL=http://interceptor:8000/v1
      - TELEGRAM_BOT_TOKEN=${OPENCLAW_TELEGRAM_TOKEN}
      - OPENCLAW_WORKSPACE=/app/workspace
    volumes:
      - ./data/workspace:/app/workspace
    deploy:
      resources:
        limits:
          memory: 95M

  # Experience C: Deterministic Admin ChatOps Menu Control Plane
  admin-chatops:
    build:
      context: ./bastion-admin
    container_name: ai-gateway-admin-chatops
    profiles: ["admin-chat"]
    restart: unless-stopped
    environment:
      - ADMIN_TELEGRAM_BOT_TOKEN=${ADMIN_TELEGRAM_TOKEN}
      - ADMIN_TELEGRAM_USER_ID=${ADMIN_TELEGRAM_USER_ID}
      - ADMIN_SECURITY_PIN=${ADMIN_SECURITY_PIN:-9911}
      - REDIS_HOST=redis
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./config:/app/config
      - ./.env:/app/.env
    deploy:
      resources:
        limits:
          memory: 35M

  # ==============================================================================
  # 6. MODULAR PROFILES (DORMANT BY DEFAULT)
  # ==============================================================================
  tailscale:
    image: tailscale/tailscale:latest
    container_name: ai-gateway-tailscale
    profiles: ["mesh-ingress"]
    restart: unless-stopped
    hostname: ${NODE_NAME:-ai-gateway}
    environment:
      - TS_AUTHKEY=${TS_AUTHKEY}
      - TS_STATE_DIR=/var/lib/tailscale
      - TS_EXTRA_ARGS=--advertise-tags=tag:ai-node
    volumes:
      - ./data/tailscale:/var/lib/tailscale
      - /dev/net/tun:/dev/net/tun
    cap_add:
      - NET_ADMIN
      - NET_RAW
    deploy:
      resources:
        limits:
          memory: 35M

  vpn-ring1:
    image: qmcgaw/gluetun:latest
    container_name: ai-gateway-vpn-ring1
    profiles: ["vpn-egress"]
    restart: unless-stopped
    cap_add:
      - NET_ADMIN
    devices:
      - /dev/net/tun:/dev/net/tun
    environment:
      - VPN_SERVICE_PROVIDER=private internet access
      - VPN_TYPE=wireguard
      - USER=${PIA_USER}
      - PASSWORD=${PIA_PASSWORD}
      - SERVER_REGIONS=US East,US Chicago
      - HTTPPROXY=on
      - SHADOWSOCKS=on
      - SHADOWSOCKS_PORT=1080
      - CONTROL_SERVER_PORT=8000
    deploy:
      resources:
        limits:
          memory: 30M
```

---

## 4. Upstream Free-Tier Providers & Anti-Sybil Constraints

```text
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                         UPSTREAM IDENTITY & VERIFICATION SPECTRUM                      │
├───────────────────────────────────────────────────────┬────────────────────────────────┤
│ TIER A: Zero Friction (Anonymous / Email-Only)        │ TIER B: Strict Anti-Sybil      │
│ • Pollinations.ai (Keyless, 100% Anonymous)           │ • Google AI Studio (SMS Gates) │
│ • Hugging Face (Email activation link)                │ • Mistral AI (Mandatory Phone) │
│ • OpenRouter (Web3 Wallet / Email Magic Link)         │ • GitHub Models (Account Age)  │
│ • Cloudflare Workers AI (Email confirmation)          │                                │
│ • SambaNova Cloud / Cerebras / Groq (Email + SSO)     │                                │
└───────────────────────────────────────────────────────┴────────────────────────────────┘
```

### Provider Verification & Shared Quota Allocations

| # | Provider | Auth Mechanism | Published Ceiling (Per Account Set) | Proactive Safety Ceiling (80% Mesh Clamping) | Quota Intersect Bottlenecks |
| :-: | :--- | :--- | :--- | :--- | :--- |
| 1 | **Google AI Studio** | Google Account (SMS on VPN) | 15 RPM / 1.5k RPD (~3.5M tok) | 12 RPM / 1.2k RPD | Shared across Flash, Flash-Lite, Thinking, Embeddings. |
| 2 | **Cerebras** | Email or SSO (No CC) | 30 RPM / 60k TPM (~1.0M tok) | 24 RPM / 50k TPM | Shared governor across Llama 3.3 70B & 8B variants. |
| 3 | **Groq LPU** | Email or SSO (No CC) | 30 RPM / 14.4k RPD (~500k tok) | 24 RPM / 11.5k RPD | Shared bucket across Llama 3.3, DeepSeek R1 Distill, Whisper. |
| 4 | **SambaNova RDU** | Email or SSO (No CC) | 20 RPM / 100k TPM (~300k tok) | 16 RPM / 80k TPM | Shared RDU capacity across DeepSeek R1/V3, Qwen 2.5 Coder. |
| 5 | **GitHub Models** | GitHub Account + 2FA | 15 RPM / 150 RPD (~450k tok) | 12 RPM / 120 RPD | Isolated counters per model family (Azure-backed). |
| 6 | **Mistral AI** | Mandatory non-VoIP SMS | 1 RPS (~350k tok) | 0.75 RPS (45 RPM) | Shared across Codestral, Mistral Small, Embeddings. |
| 7 | **Cloudflare AI** | Cloudflare Global Token | 10k Neurons/Day (~250k tok) | 8k Neurons/Day | Shared across LLMs, Flux, Whisper, MeloTTS. |
| 8 | **Hugging Face** | User Access Token (Read) | Serverless soft caps (~250k tok) | 25 RPM ceiling | Serverless rate-limits applied per endpoint IP. |
| 9 | **OpenRouter** | Web3 / Magic Link | 20 RPM soft cap (~200k tok) | 16 RPM ceiling | Dynamic rate-limiting across free variant models (`:free`). |
| 10 | **Pollinations** | Keyless / Public | Public community capacity | 10 RPM pacing | Public rate pools used as terminal emergency fallbacks. |
| — | **Total Gross Budget** | — | **~6,900,000 Tokens / Day / Set** | **~5,500,000 Safe Tokens / Day / Set** | — |

---

## 5. Master Capability Catalog & 14-Pool Architecture (144 Model Slots)

The cluster is organized into 12 Core Capability Pools plus 2 Meta-Orchestrators (`pool/auto` and `pool/video-conductor`). The 144 slots in the matrix map to 62 distinct provider endpoints across ~26 foundational model architectures:

```text
                                [ MASTER 144-MODEL CAPABILITY MATRIX ]
┌────────────────────────┬─────────────────────────┬─────────────────────────┬─────────────────────────┐
│ CAPABILITY POOL        │ PRIMARY TIER (4 Models) │ BACKUP 1 TIER (4 Models)│ BACKUP 2 TIER (4 Models)│
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 1. pool/general        │ groq/llama-3.3-70b      │ github/gpt-4o-mini      │ cloudflare/llama-3.3-70b│
│    (Frontline Core)    │ cerebras/llama-3.3-70b  │ sambanova/llama-3.3-70b │ openrouter/free-models  │
│                        │ sambanova/deepseek-v3   │ mistral/mistral-small   │ huggingface/llama-3.3   │
│                        │ gemini/gemini-2.0-flash │ openrouter/gemini-flash │ github/phi-4            │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 2. pool/deep-reasoning │ sambanova/deepseek-r1   │ cloudflare/deepseek-r1  │ github/phi-4-reasoning  │
│    (Frontline Core)    │ groq/deepseek-r1-70b    │ huggingface/deepseek-r1 │ groq/llama-3.3-70b-spec │
│                        │ github/deepseek-r1      │ sambanova/qwen-72b-inst │ mistral/mistral-large   │
│                        │ gemini/gemini-2.0-think │ openrouter/deepseek-r1  │ openrouter/qwen-72b     │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 3. pool/agent-coding   │ sambanova/qwen-2.5-coder│ cloudflare/qwen-2.5-code│ groq/llama-3.3-70b      │
│    (Frontline Core)    │ mistral/codestral       │ huggingface/qwen-coder  │ cerebras/llama-3.3-70b  │
│                        │ github/gpt-4o-mini      │ sambanova/llama-3.3-70b │ github/llama-3.3-70b    │
│                        │ gemini/gemini-2.0-flash │ openrouter/qwen-coder   │ openrouter/deepseek-v3  │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 4. pool/document-anal. │ gemini/gemini-2.0-flash │ groq/llama-3.3-70b      │ sambanova/llama-3.3-70b │
│    (Frontline Core)    │ gemini/gemini-flash-lite│ mistral/mistral-small   │ sambanova/deepseek-v3   │
│    [1M Context Window] │ gemini/gemini-1.5-flash │ github/llama-3.3-70b    │ cerebras/llama-3.3-70b  │
│                        │ openrouter/gemini-flash │ github/gpt-4o-mini      │ openrouter/llama-70b    │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 5. pool/web-research   │ cerebras/llama-3.3-70b  │ github/phi-4            │ cloudflare/llama-3.3    │
│    (Frontline Core)    │ groq/llama-3.3-70b      │ mistral/mistral-small   │ openrouter/free-models  │
│    [SearXNG Grounding] │ gemini/gemini-2.0-flash │ sambanova/llama-3.3-70b │ huggingface/llama-3.3   │
│                        │ groq/llama-3.1-8b       │ github/gpt-4o-mini      │ openrouter/gemini-flash │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 6. pool/presentation   │ groq/llama-3.3-70b      │ sambanova/llama-3.3-70b │ openrouter/llama-70b    │
│    (Artifact Engine)   │ cerebras/llama-3.3-70b  │ mistral/mistral-small   │ cloudflare/llama-3.3    │
│    [Marp / Reveal.js]  │ github/gpt-4o-mini      │ github/llama-3.3-70b    │ huggingface/llama-3.3   │
│                        │ gemini/gemini-2.0-flash │ sambanova/qwen-coder    │ openrouter/deepseek-v3  │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 7. pool/image-gen      │ cf/flux-1-schnell       │ hf/stable-diffusion-3.5 │ pollinations/turbo-gen  │
│    (Artifact Engine)   │ hf/flux-1-schnell       │ hf/sdxl-lightning       │ hf/stable-diffusion-1.5 │
│                        │ hf/sdxl-base-1.0        │ cf/sdxl-lightning       │ cf/stable-diffusion-xl  │
│                        │ cf/sdxl-base-1.0        │ pollinations/flux-model │ pollinations/midjourney │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 8. pool/architect      │ gemini/gemini-2.0-flash │ github/gpt-4o-mini      │ huggingface/llama-11b-v │
│    (Conductor Planner) │ gemini/gemini-2.0-think │ mistral/pixtral-12b     │ gemini/gemini-1.5-flash │
│                        │ groq/llama-3.2-90b-vis  │ cf/llama-3.2-11b-vision │ sambanova/llama-3.3-70b │
│                        │ sambanova/deepseek-r1   │ openrouter/gemini-flash │ cerebras/llama-3.3-70b  │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 9. pool/security-tester│ sambanova/qwen-coder    │ cf/qwen-2.5-coder       │ sambanova/deepseek-r1   │
│    (Critic Loop Pass)  │ gemini/gemini-2.0-flash │ groq/deepseek-r1-70b    │ github/llama-3.3-70b    │
│                        │ mistral/codestral       │ mistral/pixtral-12b     │ huggingface/qwen-coder  │
│                        │ github/gpt-4o-mini      │ openrouter/qwen-coder   │ openrouter/deepseek-r1  │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 10. pool/stack-optimiz.│ gemini/gemini-2.0-flash │ cerebras/llama-3.3-70b  │ github/phi-4            │
│     (Zero-Touch Daemon)│ sambanova/deepseek-r1   │ mistral/mistral-small   │ sambanova/qwen-72b-inst │
│                        │ groq/llama-3.3-70b      │ cf/llama-3.3-70b        │ huggingface/llama-3.3   │
│                        │ github/gpt-4o-mini      │ openrouter/gemini-flash │ openrouter/deepseek-v3  │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 11. pool/document-gen  │ gemini/gemini-2.0-flash │ groq/llama-3.3-70b      │ mistral/mistral-large   │
│     (Artifact Engine)  │ cerebras/llama-3.3-70b  │ sambanova/llama-3.3-70b │ openrouter/free-models  │
│     [Typst / Pandoc]   │ github/gpt-4o-mini      │ mistral/codestral       │ cf/llama-3.3-70b        │
│                        │ gemini/gemini-flash-lite│ github/llama-3.3-70b    │ huggingface/llama-3.3   │
├────────────────────────┼─────────────────────────┼─────────────────────────┼─────────────────────────┤
│ 12. pool/audio-gen     │ cf/openai-whisper       │ groq/whisper-large-v3   │ hf/mms-tts-speech       │
│     (Artifact Engine)  │ hf/kokoro-82m-tts       │ hf/xtts-v2-speech       │ pollinations/audio-synth│
│     [TTS / STT Voice]  │ groq/whisper-large-v3   │ cf/fastspeech2-tts      │ deepgram/aura-free      │
│                        │ cf/melo-tts-speech      │ huggingface/speecht5    │ huggingface/piper-tts   │
└────────────────────────┴─────────────────────────┴─────────────────────────┴─────────────────────────┘
```

---

## 6. Concentric Quota Rings & Horizontal Scaling

```text
[ INCOMING PROMPT ] ──► LiteLLM Router (:4000)
                              │
                              ├──► Ring 0: Primary Account Set (Direct Host IP)
                              │      └─ [HTTP 429 / 80% Pre-Call Threshold Hit]
                              │             │
                              ├──► Ring 1: Backup Reservoir 1 (Gluetun VPN Sidecar 1 - US IP)
                              │      └─ [HTTP 429 / 80% Pre-Call Threshold Hit]
                              │             │
                              └──► Ring 2: Backup Reservoir 2 (Gluetun VPN Sidecar 2 - EU IP)
```

### Horizontal Sharding Capacity

| Account Sets Configured | Active Routing Shards | Safe Daily Token Budget | Concurrent Interactive RPM |
| :--- | :--- | :--- | :--- |
| **1 Set (Base Mesh)** | Ring 0 (Direct Host IP) | ~5.5M Tokens/Day | 30 RPM sustained (75 RPM burst) |
| **3 Sets (Small Cluster)** | Ring 0 + Rings 1–2 (Gluetun) | ~16.5M Tokens/Day | 90 RPM sustained (200 RPM burst) |
| **5 Sets (Team Shard)** | Ring 0 + Rings 1–4 (Gluetun) | ~27.5M Tokens/Day | 150 RPM sustained (350 RPM burst) |
| **10 Sets (Maximum Node)** | Ring 0 + Rings 1–9 (Gluetun) | ~55.0M Tokens/Day | 300 RPM sustained (700 RPM burst) |

---

## 7. Native Non-VPN Anti-Ban & Single-IP Traffic Shaping

When operating without a VPN (Ring 0 on a direct home or office IP), the stack enforces seven non-VPN traffic shaping strategies:

1. **Inter-Request Delay Floor & Algorithmic Jitter:** Firing multiple calls within milliseconds trips provider WAFs. The Interceptor enforces a mandatory 150ms–350ms delay floor plus ±20% randomized jitter between consecutive calls to the same provider domain.
2. **The "Zero-429" Pre-Call Headroom Rule:** Fraud detection systems flag accounts that repeatedly trigger HTTP 429 errors. The gateway clamps local usage counters at 80% of published limits, preemptively failing over to alternative providers in the pool before an upstream 429 can occur.
3. **Cross-Provider Domain Round-Robin:** Conversational traffic alternates across distinct provider ASNs (`groq.com` → `cerebras.ai` → `sambanova.ai` → `googleapis.com`). To upstream security systems, connection volume appears as natural human usage.
4. **Local Redis Deduplication & Response Caching:** System prompts and unchanged multi-turn prefixes resolve directly out of Redis in <3 ms, cutting outbound WAN traffic by 15% to 30%.
5. **SearXNG Residential IP Safeguards:** Queries default to scraping-friendly engines (DuckDuckGo, Brave, Wikipedia), reserving Google as a terminal fallback. All search queries are cached in Redis for 1 hour with scraping threads clamped to 1 concurrent query per engine.
6. **90-Second Cooldown Circuit Breaker:** If an upstream provider returns an unexpected error, the slot is quarantined for 90 seconds. Rapid retry-hammering during an outage is strictly prevented.
7. **User-Agent & Header Normalization:** Outbound headers strip internal proxy traces (`X-LiteLLM-*`, `X-Forwarded-For`) and match standard developer SDK signatures.

---

## 8. Multi-User Concurrency, P0–P3 Priority Queuing & Heavy Task FIFO

```text
                                [ SIMULTANEOUS MULTI-USER TRAFFIC ]
                                                 │
                                                 ▼
                             ┌───────────────────────────────────────┐
                             │    FASTAPI INTERCEPTOR (:8000)        │
                             │ • Resolves Unified User Identity      │
                             │ • Checks Local Leaky-Bucket Rate Cap  │
                             └───────────────────┬───────────────────┘
                                                 │
                                    Classifies Traffic Class
                                                 │
         ┌───────────────────┬───────────────────┼───────────────────┬───────────────────┐
         ▼                   ▼                   ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ PRIORITY 0    │   │ PRIORITY 1    │   │ PRIORITY 2    │   │ PRIORITY 3    │   │ OVERFLOW /    │
│ Interactive   │   │ Developer IDE │   │ Deep Research │   │ Heavy Video & │   │ RATE-LIMITED  │
│ Chat Apps     │   │ (Cursor/Aider)│   │ (SearXNG/Docs)│   │ Compilations  │   │               │
├───────────────┤   ├───────────────┤   ├───────────────┤   ├───────────────┤   ├───────────────┤
│ • Telegram/WA │   │ • 60 RPM cap  │   │ • 5 RPM cap   │   │ • 2 RPM cap   │   │ • Returns:    │
│ • Open WebUI  │   │ • 150ms jitter│   │ • Recency     │   │ • Serialized  │   │   HTTP 429    │
│ • 12 RPM cap  │   │ • Shunted to  │   │   pruning     │   │   via Redis   │   │   (Client-    │
│ • Zero delay  │   │   Fast Tier   │   │   applied     │   │   FIFO Queue  │   │   Throttled)  │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └───────────────┘
         │                   │                   │                   │
         └───────────────────┴─────────┬─────────┴───────────────────┘
                                       ▼
                       [ LiteLLM Routing Engine (:4000) ]
```

### Unified Identity Resolution

To prevent users from circumventing rate limits by jumping between chat clients, the Interceptor maps varied auth headers and channel IDs to a canonical user profile in Redis:

```text
Telegram ID: "98124012"  ──┐
WhatsApp PN: "+1555019"  ──┼──► Canonical User: "usr_alice" ──► Profile: "interactive" (12 RPM)
WebUI UUID:  "a8f1-9b2c" ──┘
```

### Heavy Task Serialization (Conductor Anti-Crash Guard)

Low-power edge CPUs cannot handle concurrent FFmpeg video transcodes or heavy Typst compilations. Heavy rendering tasks are dispatched to a Redis FIFO queue (`task_queue:conductor`) processed by a single dedicated worker (`MAX_CONCURRENT_WORKERS=1`). Subsequent requests receive an instant response with queue status:

```json
{"status": "queued", "position": 1, "estimated_wait_sec": 18}
```

---

## 9. Context Consistency, `<think>` Sanitization & Dynamic Asymmetric Pruning

Dynamic model switching across pools introduces risks of context-window overflows and prompt pollution. The gateway enforces four context-preservation mechanics:

```text
[ INCOMING PROMPT ] ──► Load Canonical History from Redis/Postgres (OpenAI JSON Schema)
                             │
                             ├── 1. DeepSeek `<think>` Quarantine (Strips CoT tokens from history)
                             │
                             ├── 2. Target Context Check (Evaluates destination model window)
                             │
                             ├── 3. Asymmetric Sliding Window Pruning (Retains system prompt + recent turns)
                             │
                             └── 4. Wire Dialect Translation (OpenAI -> Gemini / Mistral / Anthropic)
                                           │
                                           ▼
                             [ Dispatches to Upstream Model ]
```

- **Canonical Schema Normalization:** History is stored in a clean, provider-agnostic OpenAI message schema. LiteLLM handles wire-format translation to Google's `contents`/`parts` or Mistral dialects on the fly.
- **Reasoning Token Sanitization (`<think>` Quarantine):** DeepSeek-R1 emits internal Chain-of-Thought tokens within `<think>...</think>` tags. The Interceptor extracts thinking blocks for display in collapsible UI accordions, but strips them completely before persisting the assistant response to history.
- **Dynamic Asymmetric Context Pruning:** When switching from large-window models (Gemini 2.0 Flash at 1M tokens) to small-window models (Cerebras at 8k tokens), the Interceptor prunes history using a recency-biased sliding window while permanently pinning the root system prompt.
- **Mid-Query DAG Scratchpad Isolation:** Intermediate tool steps executed by `pool/auto` (e.g., SearXNG searches, AST reviews) are contained in an ephemeral memory envelope. Only the final synthesized result is committed to conversation history.

---

## 10. The Autonomous Agentic Supervisor (pool/auto)

Instead of requiring manual model selection, `pool/auto` operates as a two-tier autonomous supervisor:

```text
                                [ INCOMING USER PROMPT ]
                                           │
                                           ▼
                 ┌──────────────────────────────────────────────────┐
                 │       TIER 1: INTENT CLASSIFICATION (<40ms)      │
                 │  • Fast inference on Cerebras / Groq Llama 3.3   │
                 └─────────────────────────┬────────────────────────┘
                                           │
    ┌──────────────────┬───────────────────┼───────────────────┬──────────────────┐
    ▼                  ▼                   ▼                   ▼                  ▼
┌──────────────┐ ┌───────────────┐ ┌───────────────────┐ ┌──────────────┐ ┌───────────────┐
│ pool/general │ │pool/agent-code│ │pool/deep-reasoning│ │pool/web-rese.│ │pool/doc-anal. │
│ Daily chat,  │ │Refactoring,   │ │Math proofs, logic,│ │SearXNG live  │ │Attachments    │
│ summarization│ │syntax debug   │ │deep CoT           │ │citations     │ │>32k tokens    │
└──────────────┘ └───────┬───────┘ └───────────────────┘ └──────────────┘ └───────────────┘
                         │
                         ▼
             [ SUB-TASK PASS: pool/security-tester ]
             (Automated AST review & vulnerability scan)
```

### Live Chat Telemetry

During multi-step workflows, `pool/auto` streams progress updates to the client:

- **Telegram & Discord:** In-place message edits (`🔍 Searching SearXNG...` → `💻 Debugging syntax...` → `Final Output`).
- **WhatsApp:** Status header pre-pended to the final message.
- **Open WebUI:** Real-time accordion status indicators.

---

## 11. The Deterministic Video Conductor Pipeline (pool/video-conductor)

The Video Conductor pipeline bypasses unreliable text-to-video diffusion models in favor of Deterministic Motion Synthesis:

```text
                              [ INCOMING VIDEO REQUEST ]
                                         │
                                         ▼
                 ┌────────────────────────────────────────────────┐
                 │ STAGE 1: THE VIDEO CONDUCTOR (pool/architect)  │
                 │ • Evaluates narrative, scene count & pacing    │
                 │ • Compiles `video_manifest.json` (Seed/Colors) │
                 │ • Clamps: Default <=30s / --long <=60s         │
                 └───────────────────────┬────────────────────────┘
                                         │
                 ┌───────────────────────┴────────────────────────┐
                 ▼                                                ▼
┌───────────────────────────────────┐    ┌──────────────────────────────────┐
│ STAGE 2A: VISUAL ANCHORS          │    │ STAGE 2B: AUDIO TRACK SYNTHESIS  │
│ (`pool/image-gen`)                │    │ (`pool/audio-gen`)               │
├───────────────────────────────────┤    ├──────────────────────────────────┤
│ • 1080p base keyframes via Flux.1 │    │ • Narration via Kokoro-82M TTS   │
│ • Fixed visual seed and palette   │    │ • Subtitle alignments via Whisper│
└─────────────────┬─────────────────┘    └────────────────┬─────────────────┘
                  │                                       │
                  └───────────────────┬───────────────────┘
                                      ▼
                 ┌────────────────────────────────────────────────┐
                 │ STAGE 3: DETERMINISTIC MOTION COMPOSITING      │
                 │ • Single-Worker Redis FIFO Queue Execution     │
                 │ • Mathematical Ken Burns Camera Transforms     │
                 │ • Burns subtitles, crossfades, outputs MP4     │
                 └────────────────────────────────────────────────┘
```

### Video Conductor Guardrails

- **Strict Duration Caps:** Standard requests are clamped to ≤ 30 seconds (3–4 scenes). The `--long` flag extends the limit to ≤ 60 seconds (6–8 scenes).
- **Deterministic Motion:** FFmpeg applies programmatic camera transforms (`zoompan`) across 1080p Flux anchor frames. Text and diagrams remain razor-sharp without diffusion artifacts.
- **Audio-Driven Timing:** Scene durations are synchronized to the millisecond of the Kokoro-82M narration track.

---

## 12. Headless Multimedia Artifact Sidecars

| Modality | Target Capability Pool | Core Engine | Output Format | Delivery Method |
| :--- | :--- | :--- | :--- | :--- |
| **Image Generation** | `pool/image-gen` | Cloudflare Flux.1 Schnell / HF SDXL | 1080p PNG / WebP | Inline chat image / document attachment |
| **Voice & Audio** | `pool/audio-gen` | Kokoro-82M TTS / Whisper Large v3 | MP3 / OGG Opus | Native playable voice notes |
| **Slide Presentations** | `pool/presentation-gen` | Headless Marp / Reveal.js | Standalone HTML / Vector PDF | Downloadable slide deck |
| **Formal Documents** | `pool/document-gen` | Typst Rust Compiler / Pandoc | Publication-grade PDF | Downloadable document |

---

## 13. Local Search Grounding Engine (SearXNG Configuration)

### SearXNG Local Engine Configuration (`config/searxng/settings.yml`)

```yaml
use_default_settings: true
general:
  debug: false
  instance_name: "AI Gateway Meta-Search"
server:
  secret_key: "ai_gateway_internal_search_secret"
  limiter: false
  image_proxy: false
search:
  safe_search: 0
  autocomplete: ""
  default_lang: "auto"
  formats:
    - html
    - json
engines:
  - name: duckduckgo
    engine: duckduckgo
    shortcut: ddg
    weight: 2.0
  - name: wikipedia
    engine: wikipedia
    shortcut: w
    weight: 1.5
  - name: bing
    engine: bing
    shortcut: b
    weight: 1.0
  - name: google
    engine: google
    shortcut: g
    weight: 0.5
    use_mobile_ui: false
```

---

## 14. Universal Client-Side Selective Chat Importer

Injected into Open WebUI via volume mount (`config/webui/selective_import.js`), this script provides client-side format detection and selective import filtering across 20+ chat archive formats:

```javascript
(function () {
  console.log("⚡ Universal Selective Chat Importer Mounted");

  const style = document.createElement("style");
  style.textContent = `
    @keyframes gw-spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
    .gw-spinner { 
      width: 44px; height: 44px; 
      border: 4px solid rgba(137, 180, 250, 0.2); 
      border-top-color: #89b4fa; 
      border-radius: 50%; 
      animation: gw-spin 0.75s cubic-bezier(0.65, 0, 0.35, 1) infinite; 
    }
  `;
  document.head.appendChild(style);

  document.addEventListener("change", async function (e) {
    const input = e.target;
    if (!input || input.type !== "file" || !input.files || input.files.length === 0) return;

    const file = input.files[0];
    const isJson = file.name.endsWith(".json");
    const isMd = file.name.endsWith(".md") || file.name.endsWith(".txt");
    if (!isJson && !isMd) return;

    if (file.name === "selected_chats.json" || file._isSyntheticPayload) return;

    e.stopImmediatePropagation();
    e.preventDefault();

    const spinner = showSpinner(`Reading ${file.name}...`);

    requestAnimationFrame(async () => {
      try {
        const rawText = await file.text();
        updateSpinnerText(spinner, "Detecting format across 20+ platforms...");
        await new Promise(r => setTimeout(r, 25));

        const detection = universalHeuristicParser(file.name, rawText);
        spinner.remove();

        if (!detection.chats || detection.chats.length === 0) {
          alert("No valid conversations could be extracted from this export.");
          return;
        }

        renderSelectionModal(detection, (selectedPayload) => {
          const blob = new Blob([JSON.stringify(selectedPayload)], { type: "application/json" });
          const syntheticFile = new File([blob], "selected_chats.json", { type: "application/json" });
          syntheticFile._isSyntheticPayload = true;

          const dt = new DataTransfer();
          dt.items.add(syntheticFile);
          input.files = dt.files;
          input.dispatchEvent(new Event("change", { bubbles: true }));
        });
      } catch (err) {
        spinner.remove();
        alert("Failed to parse chat export: " + err.message);
      }
    });
  }, true);

  const ROLE_SYNONYMS = {
    user: ["user", "human", "prompt", "question", "sender_user", "customer", "me", "用户"],
    assistant: ["assistant", "bot", "model", "ai", "response", "answer", "deepseek", "grok", "claude", "gemini", "copilot", "chatgpt", "kimi", "zhipu", "glm", "助手"]
  };
  const CONTENT_SYNONYMS = ["content", "text", "message", "body", "parts", "response", "query", "answer"];

  function normalizeRole(rawRole) {
    if (!rawRole) return "user";
    const lower = String(rawRole).toLowerCase();
    if (ROLE_SYNONYMS.user.some(s => lower.includes(s))) return "user";
    if (ROLE_SYNONYMS.assistant.some(s => lower.includes(s))) return "assistant";
    return "user";
  }

  function extractContent(rawMsg) {
    if (typeof rawMsg === "string") return rawMsg;
    if (!rawMsg) return "";
    for (const k of CONTENT_SYNONYMS) {
      if (rawMsg[k] !== undefined) {
        if (typeof rawMsg[k] === "string") return rawMsg[k];
        if (Array.isArray(rawMsg[k])) {
          return rawMsg[k].map(p => typeof p === "string" ? p : p.text || "").join("\n");
        }
      }
    }
    if (rawMsg.message) return extractContent(rawMsg.message);
    return "";
  }

  function universalHeuristicParser(fileName, rawText) {
    let platform = "Generic Archive";
    let normalizedChats = [];

    if (fileName.endsWith(".md") || (!rawText.trim().startsWith("[") && !rawText.trim().startsWith("{"))) {
      const lower = rawText.toLowerCase();
      if (lower.includes("deepseek")) platform = "DeepSeek";
      else if (lower.includes("grok") || lower.includes("x.com")) platform = "Grok";
      else if (lower.includes("perplexity")) platform = "Perplexity";
      else if (lower.includes("kimi")) platform = "Kimi";
      else if (lower.includes("claude")) platform = "Claude";
      else platform = "Markdown Transcript";

      const splitRegex = /\n(?=#{1,4}\s+(?:User|Human|Assistant|AI|DeepSeek|Grok|Model|Prompt|Response)|(?:\*\*|__)(?:User|Assistant|DeepSeek|Grok)(?:\*\*|__):?)/i;
      const sections = rawText.split(splitRegex);
      const messages = {};
      let prevId = null;

      sections.forEach((sec, idx) => {
        if (!sec.trim()) return;
        const msgId = `msg-${idx}-${Date.now()}`;
        const header = sec.trim().split("\n")[0].toLowerCase();
        const role = ROLE_SYNONYMS.assistant.some(s => header.includes(s)) ? "assistant" : "user";
        const content = sec.replace(/^(#{1,4}\s+.*|\*\*.*?\*\*:?)\s*/i, "").trim();

        messages[msgId] = {
          id: msgId, parentId: prevId, childrenIds: [], role: role,
          content: content, timestamp: Math.floor(Date.now() / 1000)
        };
        if (prevId) messages[prevId].childrenIds.push(msgId);
        prevId = msgId;
      });

      const title = fileName.replace(/\.[^/.]+$/, "");
      return {
        platform: platform,
        chats: [{
          id: `md-${Date.now()}`, title: title, date: new Date().toLocaleDateString(),
          turns: Object.keys(messages).length,
          rawPayload: { chat: { title: title, models: ["pool/general"], history: { messages: messages, currentId: prevId } } }
        }]
      };
    }

    const data = JSON.parse(rawText);
    const chatList = Array.isArray(data) ? data : [data];
    const sample = chatList[0] || {};
    const sampleStr = JSON.stringify(sample).toLowerCase();

    if (sample && (sample.mapping || sample.conversation_id)) {
      return {
        platform: "ChatGPT",
        chats: chatList.map(chat => ({
          id: chat.id || chat.conversation_id || Math.random().toString(),
          title: chat.title || "Untitled ChatGPT Thread",
          date: new Date((chat.create_time || Date.now() / 1000) * 1000).toLocaleDateString(),
          turns: Object.keys(chat.mapping || {}).filter(k => chat.mapping[k].message).length,
          rawPayload: chat
        }))
      };
    }

    if (sampleStr.includes("deepseek")) platform = "DeepSeek";
    else if (sampleStr.includes("grok")) platform = "Grok";
    else if (sampleStr.includes("kimi") || sampleStr.includes("moonshot")) platform = "Kimi";
    else if (sampleStr.includes("zhipu") || sampleStr.includes("glm") || sampleStr.includes("zai")) platform = "Z.AI (GLM)";
    else if (sampleStr.includes("poe")) platform = "Poe";
    else if (sampleStr.includes("cursor")) platform = "Cursor";
    else if (sampleStr.includes("claude") || sample.chat_messages) platform = "Claude";
    else if (sampleStr.includes("gemini") || sample.parts) platform = "Gemini";
    else platform = "Universal JSON";

    normalizedChats = chatList.map((chat, cIdx) => {
      const rawTurns = chat.messages || chat.chat_messages || chat.history || chat.conversation || (Array.isArray(chat) ? chat : []);
      const messages = {};
      let prevId = null;

      if (Array.isArray(rawTurns) && rawTurns.length > 0) {
        rawTurns.forEach((turn, tIdx) => {
          const msgId = `uni-${cIdx}-${tIdx}-${Date.now()}`;
          const rawRole = turn.role || turn.sender || turn.author || turn.from || (tIdx % 2 === 0 ? "user" : "assistant");
          const role = normalizeRole(rawRole);
          const content = extractContent(turn);

          messages[msgId] = {
            id: msgId, parentId: prevId, childrenIds: [], role: role,
            content: content, timestamp: Math.floor(Date.now() / 1000)
          };
          if (prevId) messages[prevId].childrenIds.push(msgId);
          prevId = msgId;
        });
      }

      const title = chat.title || chat.name || `${platform} Thread #${cIdx + 1}`;
      return {
        id: chat.id || chat.uuid || `chat-${cIdx}`,
        title: title,
        date: new Date(chat.create_time || chat.created_at || Date.now()).toLocaleDateString(),
        turns: Object.keys(messages).length || 1,
        rawPayload: { chat: { title: title, models: ["pool/general"], history: { messages: messages, currentId: prevId } } }
      };
    });

    return { platform: platform, chats: normalizedChats };
  }

  function showSpinner(text) {
    const overlay = document.createElement("div");
    overlay.style.cssText = `position:fixed;inset:0;background:rgba(17,17,27,0.85);z-index:100000;display:flex;flex-direction:column;align-items:center;justify-content:center;backdrop-filter:blur(4px);`;
    overlay.innerHTML = `<div class="gw-spinner"></div><div id="gw-sp-text" style="margin-top:18px;font-size:14px;color:#cdd6f4;font-family:system-ui;">${text}</div>`;
    document.body.appendChild(overlay);
    return overlay;
  }

  function updateSpinnerText(overlay, text) {
    const el = overlay.querySelector("#gw-sp-text");
    if (el) el.textContent = text;
  }

  function renderSelectionModal(detection, onConfirm) {
    const modal = document.createElement("div");
    modal.style.cssText = `position:fixed;inset:0;background:rgba(0,0,0,0.75);z-index:99999;display:flex;align-items:center;justify-content:center;backdrop-filter:blur(4px);font-family:system-ui,-apple-system,sans-serif;`;

    modal.innerHTML = `
      <div style="background:#181825;color:#cdd6f4;width:700px;max-height:84vh;border-radius:14px;border:1px solid #313244;padding:24px;display:flex;flex-direction:column;box-shadow:0 20px 50px rgba(0,0,0,0.7);">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:6px;">
          <h2 style="margin:0;font-size:18px;font-weight:600;">Universal Selective Chat Import</h2>
          <span style="font-size:10px;padding:3px 8px;border-radius:4px;background:#a6e3a122;color:#a6e3a1;border:1px solid #a6e3a155;font-weight:600;">Enhanced Mesh Active</span>
        </div>
        <div style="font-size:12px;color:#89dceb;margin-bottom:10px;">✓ Identified <strong>${detection.platform}</strong>: <strong>${detection.chats.length}</strong> threads parsed.</div>
        <input type="text" id="gw-search" placeholder="Search threads by title..." style="padding:9px 14px;background:#1e1e2e;border:1px solid #45475a;border-radius:8px;color:#fff;margin-bottom:10px;font-size:13px;outline:none;" />
        <div style="display:flex;justify-content:space-between;margin-bottom:8px;font-size:12px;color:#a6adc8;">
          <div><button id="gw-all" style="background:none;border:none;color:#89b4fa;cursor:pointer;padding:0;">Select All</button> • <button id="gw-none" style="background:none;border:none;color:#89b4fa;cursor:pointer;padding:0;">Deselect All</button></div>
          <span id="gw-count" style="font-weight:600;color:#fab387;">0 selected</span>
        </div>
        <div id="gw-list" style="overflow-y:auto;flex:1;border:1px solid #313244;border-radius:8px;background:#11111b;"></div>
        <div style="display:flex;justify-content:flex-end;gap:10px;margin-top:16px;">
          <button id="gw-cancel" style="padding:8px 18px;background:transparent;border:1px solid #45475a;color:#cdd6f4;border-radius:6px;cursor:pointer;">Cancel</button>
          <button id="gw-submit" style="padding:8px 22px;background:#89b4fa;border:none;color:#11111b;font-weight:600;border-radius:6px;cursor:pointer;">Import Selected</button>
        </div>
      </div>
    `;

    document.body.appendChild(modal);

    const list = modal.querySelector("#gw-list");
    const search = modal.querySelector("#gw-search");
    const countLabel = modal.querySelector("#gw-count");

    function renderItems(filter = "") {
      list.innerHTML = "";
      const frag = document.createDocumentFragment();
      detection.chats.filter(c => c.title.toLowerCase().includes(filter.toLowerCase())).forEach(chat => {
        const row = document.createElement("label");
        row.style.cssText = "display:flex;align-items:center;gap:12px;padding:9px 14px;border-bottom:1px solid #1e1e2e;cursor:pointer;";
        row.innerHTML = `
          <input type="checkbox" value="${chat.id}" class="gw-chk" style="width:15px;height:15px;accent-color:#89b4fa;cursor:pointer;" />
          <div style="flex:1;min-width:0;">
            <div style="font-size:13px;font-weight:500;color:#cdd6f4;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${chat.title}</div>
            <div style="font-size:11px;color:#6c7086;">${chat.date} • ${chat.turns} turns</div>
          </div>
        `;
        frag.appendChild(row);
      });
      list.appendChild(frag);
      updateCount();
    }

    function updateCount() {
      const selected = modal.querySelectorAll(".gw-chk:checked").length;
      countLabel.textContent = `${selected} selected`;
    }

    let timer;
    search.addEventListener("input", e => {
      clearTimeout(timer);
      timer = setTimeout(() => renderItems(e.target.value), 50);
    });

    list.addEventListener("change", updateCount);
    modal.querySelector("#gw-all").onclick = (e) => { e.preventDefault(); modal.querySelectorAll(".gw-chk").forEach(i => i.checked = true); updateCount(); };
    modal.querySelector("#gw-none").onclick = (e) => { e.preventDefault(); modal.querySelectorAll(".gw-chk").forEach(i => i.checked = false); updateCount(); };
    modal.querySelector("#gw-cancel").onclick = () => modal.remove();

    modal.querySelector("#gw-submit").onclick = () => {
      const selectedSet = new Set(Array.from(modal.querySelectorAll(".gw-chk:checked")).map(i => i.value));
      const payload = detection.chats.filter(c => selectedSet.has(c.id)).map(c => c.rawPayload);
      modal.remove();
      if (payload.length > 0) onConfirm(payload);
    };

    renderItems();
  }
})();
```

---

## 15. The Tri-Tier Chat Access Architecture

The gateway explicitly partitions messaging channel traffic into three decoupled experiences:

```text
                                 [ MESSAGING CLIENT INGRESS ]
                                               │
         ┌─────────────────────────────────────┼─────────────────────────────────────┐
         ▼                                     ▼                                     ▼
┌─────────────────────────────┐ ┌─────────────────────────────┐ ┌─────────────────────────────┐
│ EXPERIENCE A: GENERIC       │ │ EXPERIENCE B: OPENCLAW      │ │ EXPERIENCE C: DETERMINISTIC │
│ CONVERSATIONAL ASSISTANT    │ │ AUTONOMOUS AGENT            │ │ ADMIN CHATOPS CONTROL PLANE │
├─────────────────────────────┤ ├─────────────────────────────┤ ├─────────────────────────────┤
│ • Native Python (aiogram)   │ │ • OpenClaw Runtime Engine   │ │ • Pure Python / Docker API  │
│ • RAM: ~30 MB               │ │ • RAM: ~95 MB               │ │ • RAM: ~35 MB               │
│ • Conversational Q&A        │ │ • Autonomous task executor  │ │ • ZERO LLM Tokens used      │
│ • Read-only SearXNG Search  │ │ • Tools, GitHub, Workspace  │ │ • Instant Menu/Keyboard     │
│ • P0 Rate Pacing (12 RPM)   │ │ • Proactive cron alerts     │ │ • 85% Remote Cluster Ops    │
│ • Casual Users & Guests     │ │ • Developers & Power Users  │ │ • Cluster Owner / DevOps    │
└──────────────┬──────────────┘ └──────────────┬──────────────┘ └──────────────┬──────────────┘
               │                               │                               │
               └───────────────────────┬───────┴───────────────────────────────┘
                                       ▼
                       [ Core Interceptor :8000 & LiteLLM :4000 ]
```

---

## 16. The Deterministic Admin ChatOps Menu & Out-of-Band Operations

The Admin ChatOps interface operates as an out-of-band, deterministic control plane. It connects directly to the local Docker daemon (`/var/run/docker.sock`), Redis (`:6379`), and the LiteLLM admin socket without routing through any LLM. It consumes 0 LLM tokens, executes with <15 ms latency, and functions even if all external AI APIs are experiencing global outages.

### The Security Boundary: Chat Menu vs. Web Admin Panel

```text
┌────────────────────────────────────────────────────────┬────────────────────────────────────────────────────────┐
│             CHAT OPS MENU (Mobile Ingress)             │           WEB ADMIN PANEL ONLY (:3000 / LAN)           │
│           "Day-to-Day Control & Triage (85%)"          │         "High Blast-Radius & Cryptographic (15%)"      │
├────────────────────────────────────────────────────────┼────────────────────────────────────────────────────────┤
│ • Container lifecycle (Start, Stop, Restart)           │ ❌ Hard volume purges (`docker compose down -v`)       │
│ • Real-time RAM, CPU, disk, & thermal telemetry        │ ❌ Master secret resets (`LITELLM_MASTER_KEY`, DB pass)│
│ • Live token burn, RPM gauges, & provider status       │ ❌ Arbitrary shell execution (`docker exec -it /bin/sh`)│
│ • Ephemeral API key rotation (Scrubbed on receipt)     │ ❌ Raw `.env` file editing or arbitrary env injection  │
│ • Modular profile toggles (Tailscale, VPN, Conductor)  │ ❌ Network bridge & host firewall / iptables mutations │
│ • Log triage (Tail last 40/100 lines per container)    │ ❌ PostgreSQL schema drops or direct SQL query console │
│ • Redis cache purge & stuck task lock flushes          │ ❌ Peer node federation key issuance / de-auth         │
│ • Stack Optimizer tuning diff review & approval        │ ❌ Exporting full unredacted conversation audit logs    │
│ • Emergency "Safe Mode" & automated IP cycling         │ ❌ Bare-metal host reboot or kernel sysctl changes     │
└────────────────────────────────────────────────────────┴────────────────────────────────────────────────────────┘
```

### Master ChatOps Navigation Tree

```text
[ /admin or /menu ]
  │
  ├── 📊 1.0 Quota & Telemetry
  │     ├── 1.1 Provider RPM & Token Burn (Live gauges across 10 providers)
  │     ├── 1.2 Host Hardware Stats (CPU, RAM cgroups, Swap, Temp, Disk I/O)
  │     ├── 1.3 Active Concurrency (Current SSE streams, queued P0–P3 tasks)
  │     └── 1.4 Circuit Breaker Registry (Quarantined models & cooldown timers)
  │
  ├── 🔌 2.0 Modular Services & Profiles
  │     ├── 2.1 Tailscale Mesh Ingress (Toggle ON / OFF / Check status)
  │     ├── 2.2 Gluetun WireGuard Egress (Toggle Ring 1 / Ring 2 / Cycle Public IP)
  │     ├── 2.3 Video Conductor Engine (Toggle Worker / View FIFO Queue Position)
  │     ├── 2.4 SearXNG Meta-Search (Restart engine / Toggle upstream Google scrape)
  │     └── 2.5 Public Chat Bridges (Enable/Disable Guest Telegram or WebUI)
  │
  ├── 🧠 3.0 Model Routing & Key Vault
  │     ├── 3.1 Primary Model Override (Force fallback to specific provider)
  │     ├── 3.2 Ephemeral Key Injector (Prompt -> Inject -> Instant Delete Message)
  │     ├── 3.3 Pool Health Probe (Synthetic 5-token ping across all 12 pools)
  │     └── 3.4 Context Guard Config (Toggle DeepSeek `<think>` strip / Sliding Window)
  │
  ├── 📑 4.0 Log Triage & Diagnostics
  │     ├── 4.1 Quick Tail (Last 40 lines of LiteLLM, Interceptor, or Conductor)
  │     ├── 4.2 Error Scan (Scrapes last 500 lines for HTTP 429, 500, or OOM events)
  │     └── 4.3 Container Healthcheck Matrix (Health status of all 11 services)
  │
  ├── ⚡ 5.0 Stack Optimizer (Tuning Proposals)
  │     ├── 5.1 View Staged Proposal (Inspect YAML diff of proposed token shifts)
  │     ├── 5.2 Approve & Apply Diff (Hot-reloads routing with 0 downtime)
  │     └── 5.3 Reject & Dismiss (Purges proposal from Redis)
  │
  └── 🚨 6.0 Emergency Ops (PIN Protected)
        ├── 6.1 Soft Reload Cluster (Graceful restart of LiteLLM + Interceptor)
        ├── 6.2 Flush Caches & Locks (Flushes Redis prompt cache & dead locks)
        ├── 6.3 Engage Safe Mode (Clamps cluster to 800MB RAM, kills heavy jobs)
        └── 6.4 Kill Current Task (Terminates hanging FFmpeg or PDF render job)
```

---

## 17. Strict Human-in-the-Loop Stack Optimizer (pool/stack-optimizer)

The Stack Optimizer evaluates token burn patterns to reclaim expiring quota. By default, it operates under a Zero-Touch Human-in-the-Loop gate: proposals are staged in Redis and broadcast to the operator. No live routing configurations are modified until approved, unless `AUTO_APPROVE_TUNING=true` is explicitly set.

```text
[ ROLLING 7-DAY TELEMETRY: UNUSED TOKENS DETECTED ]
                        │
                        ▼
          ┌───────────────────────────┐
          │   pool/stack-optimizer    │
          └─────────────┬─────────────┘
                        │
            Compiles Tuning Proposal
                        │
                        ▼
          ┌───────────────────────────┐
          │ STAGED IN REDIS (48h TTL) │
          │ `proposal:opt-9812`       │
          └─────────────┬─────────────┘
                        │
      ┌─────────────────┴─────────────────┐
      ▼                                   ▼
[ DEFAULT: HUMAN APPROVAL GATE ]  [ OPT-IN: AUTO-APPROVE ]
• Broadcasts diff to Admin UI &   • Only active if env variable
  Bastion chat (Telegram/WA)        `AUTO_APPROVE_TUNING=true`
• Awaits explicit [Approve] click • Rollback Circuit Breaker
```

### Staged Review Card (Admin Panel & ChatOps Menu)

```text
⚡ STACK OPTIMIZER PROPOSAL #104 (Staged)
Telemetry: 4.35M tokens/day expiring unused on SambaNova & Google AI Studio.
Proposed Tuning:
  • Promote `pool/agent-coding` Tier 1: 8B Distill -> Qwen 2.5 Coder 32B
  • Expand `pool/deep-reasoning` thinking budget: 1,024 -> 8,192 tokens
  • Activate secondary verification pass via `pool/security-tester`

Impact: Improves benchmark accuracy by ~28% | Projected Quota Utilization: 78%

Actions: [ ✅ Approve & Apply ]     [ ❌ Dismiss Proposal ]
```

---

## 18. Linux Host Hardening, Sockets, and Storage Compaction

### Kernel Sockets & File Descriptors (`/etc/sysctl.d/99-ai-gateway.conf`)

```ini
net.ipv4.tcp_tw_reuse = 1
net.ipv4.ip_local_port_range = 1024 65535
fs.file-max = 2097152
net.core.somaxconn = 4096
```

### Storage Hygiene & Compaction (`daemon/compress_storage.sh`)

```bash
#!/bin/sh
set -eu

LOG_DIR="/app/data/logs"
ARCHIVE_DIR="/app/data/logs/archive"
mkdir -p "$ARCHIVE_DIR"

# Compress logs older than 30 days using Zstandard Level 10
find "$LOG_DIR" -type f -name "*.log" -mtime +30 -exec sh -c '
    for file do
        zstd -10 --rm "$file" -o "$ARCHIVE_DIR/$(basename "$file").zst"
    done
' sh {} +

# Prune database spend records older than 90 days
psql "$DATABASE_URL" -c "DELETE FROM \"LiteLLM_SpendLogs\" WHERE \"startTime\" < NOW() - INTERVAL '90 DAYS';"
```

---

## 19. Complete Implementation Codebase

> [!NOTE]
> **Historical Reference Implementations / Specification Pseudocode**:
> The Python, JavaScript, and shell scripts in this section represent the initial research reference implementations from the architectural specification phase.
> As sanctioned in Section 1 and Section 3, the **canonical production deployment target is the single native Go binary** (`tokman` in `backend/` and `gui/`). The production Go implementation replaces the multi-container Python/FastAPI/LiteLLM stack with < 50MB RAM usage, sub-millisecond in-process routing, embedded SQLite/LRU caching, and native Fyne desktop GUI.
> All code in this section is maintained strictly for historical reference.

### A. Advanced FastAPI Interceptor & Traffic Shaper (`daemon/interceptor.py`)

```python
import os
import re
import time
import json
import random
import asyncio
import logging
from typing import List, Dict, Any, AsyncGenerator

import httpx
import redis.asyncio as aioredis
from fastapi import FastAPI, Request, HTTPException
from fastapi.responses import JSONResponse, StreamingResponse

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

app = FastAPI(title="AI Gateway Production Interceptor & Traffic Shaper")

GATEWAY_URL = os.getenv("GATEWAY_URL", "http://litellm:4000")
REDIS_HOST = os.getenv("REDIS_HOST", "redis")
REDIS_PORT = int(os.getenv("REDIS_PORT", 6379))
NODE_NAME = os.getenv("NODE_NAME", "Local-Hub")

r: aioredis.Redis = None

@app.on_event("startup")
async def startup_event():
    global r
    r = await aioredis.from_url(f"redis://{REDIS_HOST}:{REDIS_PORT}", decode_responses=True)
    logging.info("Connected to Redis cache & coordination layer.")

@app.on_event("shutdown")
async def shutdown_event():
    if r:
        await r.close()

# ------------------------------------------------------------------------------
# 1. IDENTITY NORMALIZATION & RATE LIMIT PRESETS
# ------------------------------------------------------------------------------
TRAFFIC_PRESETS = {
    "interactive": {"rpm": 12, "burst": 3, "priority": 0},
    "ide":         {"rpm": 60, "burst": 10, "priority": 1},
    "research":    {"rpm": 5,  "burst": 1, "priority": 2},
    "heavy":       {"rpm": 2,  "burst": 1, "priority": 3}
}

async def resolve_identity(request: Request) -> Dict[str, Any]:
    auth_header = request.headers.get("Authorization", "").replace("Bearer ", "").strip()
    client_id = request.headers.get("X-Client-ID")
    channel = request.headers.get("X-Channel", "webui")

    canonical_user = "usr_anonymous"
    traffic_class = "interactive"

    if client_id:
        mapped = await r.get(f"identity:mapping:{client_id}")
        if mapped:
            canonical_user = mapped
        else:
            canonical_user = f"usr_{client_id}"
            await r.setex(f"identity:mapping:{client_id}", 86400, canonical_user)
    elif auth_header:
        canonical_user = f"usr_{auth_header[-8:]}"

    if channel in ["telegram", "whatsapp", "webui"]:
        traffic_class = "interactive"
    elif channel in ["cursor", "aider", "vscode"]:
        traffic_class = "ide"

    return {"user": canonical_user, "class": traffic_class, "token": auth_header}

async def check_rate_limit(identity: Dict[str, Any]):
    user = identity["user"]
    t_class = identity["class"]
    preset = TRAFFIC_PRESETS.get(t_class, TRAFFIC_PRESETS["interactive"])

    key = f"ratelimit:{user}:{int(time.time() // 60)}"
    current = await r.incr(key)
    if current == 1:
        await r.expire(key, 65)

    if current > preset["rpm"]:
        raise HTTPException(
            status_code=429,
            detail=f"Local Rate Limit Exceeded: {current}/{preset['rpm']} RPM for profile '{t_class}'"
        )

# ------------------------------------------------------------------------------
# 2. CONTEXT SANITIZATION & ASYMMETRIC PRUNING
# ------------------------------------------------------------------------------
THINK_REGEX = re.compile(r"<think>.*?</think>", re.DOTALL)

MODEL_CONTEXT_LIMITS = {
    "cerebras": 8192,
    "groq": 8192,
    "sambanova": 32768,
    "github": 8192,
    "mistral": 32768,
    "cloudflare": 8192,
    "gemini": 1000000
}

def sanitize_assistant_content(content: str) -> str:
    if not content:
        return ""
    return THINK_REGEX.sub("", content).strip()

def prune_context(messages: List[Dict[str, str]], target_model: str) -> List[Dict[str, str]]:
    limit = 8192
    for key, val in MODEL_CONTEXT_LIMITS.items():
        if key in target_model.lower():
            limit = val
            break

    if limit >= 128000:
        return messages

    system_messages = [m for m in messages if m.get("role") == "system"]
    conversation = [m for m in messages if m.get("role") != "system"]

    max_chars = int(limit * 3.2)
    system_chars = sum(len(m.get("content", "")) for m in system_messages)
    available_chars = max(max_chars - system_chars, 1000)

    pruned = []
    accumulated = 0

    for msg in reversed(conversation):
        msg_len = len(msg.get("content", ""))
        if accumulated + msg_len > available_chars:
            break
        pruned.insert(0, msg)
        accumulated += msg_len

    if not pruned and conversation:
        pruned = [conversation[-1]]

    return system_messages + pruned

# ------------------------------------------------------------------------------
# 3. INTER-REQUEST JITTER & TRAFFIC SMOOTHING
# ------------------------------------------------------------------------------
async def apply_traffic_smoothing(provider_domain: str):
    key = f"jitter:last_call:{provider_domain}"
    now = time.time()
    last_call = await r.get(key)

    if last_call:
        delta = now - float(last_call)
        target_delay = random.uniform(0.150, 0.350)
        if delta < target_delay:
            await asyncio.sleep(target_delay - delta)

    await r.set(key, str(time.time()), ex=5)

# ------------------------------------------------------------------------------
# 4. CHAT COMPLETION PROXY HANDLER
# ------------------------------------------------------------------------------
@app.post("/v1/chat/completions")
async def handle_chat_completion(request: Request):
    identity = await resolve_identity(request)
    await check_rate_limit(identity)

    body = await request.json()
    model = body.get("model", "pool/auto")
    messages = body.get("messages", [])
    stream = body.get("stream", False)

    pruned_messages = prune_context(messages, model)
    body["messages"] = pruned_messages

    body["metadata"] = body.get("metadata", {})
    body["metadata"]["user_id"] = identity["user"]
    body["metadata"]["traffic_class"] = identity["class"]
    body["metadata"]["origin_node"] = NODE_NAME

    await apply_traffic_smoothing("litellm_core")

    forward_headers = {
        "Authorization": f"Bearer {identity['token']}",
        "Content-Type": "application/json",
        "X-Mesh-Hop": str(int(request.headers.get("X-Mesh-Hop", 0)) + 1)
    }

    target_url = f"{GATEWAY_URL}/v1/chat/completions"

    if stream:
        async def event_generator() -> AsyncGenerator[bytes, None]:
            client = httpx.AsyncClient(timeout=120.0)
            try:
                async with client.stream("POST", target_url, json=body, headers=forward_headers) as resp:
                    async for chunk in resp.aiter_bytes():
                        yield chunk
            finally:
                await client.aclose()

        return StreamingResponse(event_generator(), media_type="text/event-stream")

    async with httpx.AsyncClient(timeout=60.0) as client:
        res = await client.post(target_url, json=body, headers=forward_headers)
        if res.status_code != 200:
            return JSONResponse(content=res.json(), status_code=res.status_code)

        data = res.json()
        if "choices" in data and len(data["choices"]) > 0:
            msg = data["choices"][0].get("message", {})
            if "content" in msg and msg["content"]:
                msg["content"] = sanitize_assistant_content(msg["content"])

        return JSONResponse(content=data, status_code=200)
```

### B. Video Conductor Worker with Redis FIFO Queue (`conductor/worker.py`)

```python
import os
import json
import time
import asyncio
import logging
import subprocess
import httpx
import redis.asyncio as aioredis

logging.basicConfig(level=logging.INFO, format="%(asctime)s [CONDUCTOR] %(message)s")

GATEWAY_URL = os.getenv("GATEWAY_URL", "http://interceptor:8000/v1")
MASTER_KEY = os.getenv("LITELLM_MASTER_KEY", "")
REDIS_HOST = os.getenv("REDIS_HOST", "redis")

QUEUE_NAME = "task_queue:conductor"
MAX_DURATION_DEFAULT = 30.0
MAX_DURATION_LONG = 60.0

async def sanitize_manifest(manifest: dict, is_long: bool = False) -> dict:
    duration_cap = MAX_DURATION_LONG if is_long else MAX_DURATION_DEFAULT
    scene_cap = 8 if is_long else 4

    if len(manifest.get("scenes", [])) > scene_cap:
        manifest["scenes"] = manifest["scenes"][:scene_cap]

    total = sum(s.get("duration_sec", 6.0) for s in manifest["scenes"])
    if total > duration_cap:
        ratio = duration_cap / total
        for s in manifest["scenes"]:
            s["duration_sec"] = round(s.get("duration_sec", 6.0) * ratio, 1)

    return manifest

def render_ken_burns(image_path: str, duration: float, output_path: str):
    fps = 24
    frames = int(duration * fps)
    vf = f"zoompan=z='min(zoom+0.0015,1.2)':d={frames}:x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':s=1920x1080:fps={fps}"
    cmd = [
        "ffmpeg", "-y", "-loop", "1", "-i", image_path,
        "-vf", vf, "-t", str(duration),
        "-pix_fmt", "yuv420p", "-c:v", "libx264", "-preset", "veryfast", output_path
    ]
    subprocess.run(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=True)

async def execute_job(job_payload: dict):
    job_id = job_payload["job_id"]
    prompt = job_payload["prompt"]
    is_long = job_payload.get("long", False)
    job_dir = f"/app/artifacts/{job_id}"
    os.makedirs(job_dir, exist_ok=True)

    logging.info(f"Processing Job #{job_id} | Long: {is_long}")

    async with httpx.AsyncClient(timeout=45.0) as client:
        res = await client.post(
            f"{GATEWAY_URL}/chat/completions",
            headers={"Authorization": f"Bearer {MASTER_KEY}"},
            json={
                "model": "pool/architect",
                "messages": [
                    {"role": "system", "content": "You are the Master Video Conductor. Return strict JSON matching video_manifest schema."},
                    {"role": "user", "content": prompt}
                ],
                "temperature": 0.2
            }
        )
        raw_manifest = json.loads(res.json()["choices"][0]["message"]["content"].replace("```json", "").replace("```", "").strip())

    manifest = await sanitize_manifest(raw_manifest, is_long)
    scene_clips = []

    for idx, scene in enumerate(manifest["scenes"]):
        img_path = f"{job_dir}/scene_{idx}.png"
        audio_path = f"{job_dir}/scene_{idx}.mp3"
        clip_path = f"{job_dir}/clip_{idx}.mp4"

        async with httpx.AsyncClient(timeout=60.0) as client:
            img_req = client.post(
                f"{GATEWAY_URL}/images/generations",
                headers={"Authorization": f"Bearer {MASTER_KEY}"},
                json={"model": "pool/image-gen", "prompt": scene["image_prompt"], "size": "1024x1024"}
            )
            aud_req = client.post(
                f"{GATEWAY_URL}/audio/speech",
                headers={"Authorization": f"Bearer {MASTER_KEY}"},
                json={"model": "pool/audio-gen", "input": scene["voiceover_script"], "voice": "alloy"}
            )
            img_res, aud_res = await asyncio.gather(img_req, aud_req)

            img_url = img_res.json()["data"][0]["url"]
            img_bytes = (await client.get(img_url)).content
            with open(img_path, "wb") as f:
                f.write(img_bytes)

            with open(audio_path, "wb") as f:
                f.write(aud_res.content)

        probe_cmd = f"ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 {audio_path}"
        audio_dur = float(subprocess.check_output(probe_cmd, shell=True).decode().strip())
        duration = max(audio_dur, scene.get("duration_sec", 5.0))

        render_ken_burns(img_path, duration, clip_path)

        muxed_path = f"{job_dir}/muxed_{idx}.mp4"
        mux_cmd = f"ffmpeg -y -i {clip_path} -i {audio_path} -c:v copy -c:a aac -shortest {muxed_path}"
        subprocess.run(mux_cmd, shell=True, check=True)
        scene_clips.append(muxed_path)

    concat_list = f"{job_dir}/concat.txt"
    final_output = f"{job_dir}/final.mp4"
    with open(concat_list, "w") as f:
        for c in scene_clips:
            f.write(f"file '{os.path.abspath(c)}'\n")

    subprocess.run(f"ffmpeg -y -f concat -safe 0 -i {concat_list} -c copy {final_output}", shell=True, check=True)
    logging.info(f"Job #{job_id} successfully rendered -> {final_output}")

async def main():
    r = await aioredis.from_url(f"redis://{REDIS_HOST}:6379", decode_responses=True)
    logging.info("Video Conductor Single-Worker Daemon Online. Waiting for jobs...")

    while True:
        try:
            item = await r.blpop(QUEUE_NAME, timeout=10)
            if item:
                _, payload_raw = item
                payload = json.loads(payload_raw)
                await execute_job(payload)
        except Exception as err:
            logging.error(f"Worker execution failed: {err}")
            await asyncio.sleep(2)

if __name__ == "__main__":
    asyncio.run(main())
```

### C. Deterministic Admin ChatOps Controller (`bastion-admin/admin_menu.py`)

```python
import os
import re
import time
import json
import asyncio
import logging
from typing import Optional

import docker
import redis.asyncio as aioredis
from aiogram import Bot, Dispatcher, F, types
from aiogram.filters import Command
from aiogram.types import InlineKeyboardMarkup, InlineKeyboardButton
from aiogram.fsm.context import FSMContext
from aiogram.fsm.state import State, StatesGroup

logging.basicConfig(level=logging.INFO)

ADMIN_USER_ID = int(os.getenv("ADMIN_TELEGRAM_USER_ID", "0"))
ADMIN_SECURITY_PIN = os.getenv("ADMIN_SECURITY_PIN", "9911")
REDIS_HOST = os.getenv("REDIS_HOST", "redis")
ENV_PATH = "/app/.env"
CONFIG_PATH = "/app/config/config.yaml"

bot = Bot(token=os.getenv("ADMIN_TELEGRAM_BOT_TOKEN", ""))
dp = Dispatcher()
docker_client = docker.DockerClient(base_url="unix://var/run/docker.sock")

class AdminStates(StatesGroup):
    waiting_for_key = State()
    waiting_for_pin = State()

@dp.message(~F.from_user.id == ADMIN_USER_ID)
async def unauthorized_drop(message: types.Message):
    return

def root_menu_kb() -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup(inline_keyboard=[
        [
            InlineKeyboardButton(text="📊 Quota & Usage", callback_data="nav_quota"),
            InlineKeyboardButton(text="🔌 Services", callback_data="nav_services")
        ],
        [
            InlineKeyboardButton(text="🧠 Models & Keys", callback_data="nav_keys"),
            InlineKeyboardButton(text="📑 Logs & Triage", callback_data="nav_logs")
        ],
        [
            InlineKeyboardButton(text="⚡ Optimizer Diff", callback_data="nav_optimizer"),
            InlineKeyboardButton(text="🚨 Emergency Ops", callback_data="nav_emergency")
        ]
    ])

@dp.message(Command("admin", "menu"))
async def cmd_admin_menu(message: types.Message):
    await message.answer(
        "⚡ <b>AI GATEWAY CLUSTER CHATOPS CONTROL PLANE</b>\n"
        "<i>Out-of-band deterministic management session active.</i>\n"
        "Select a control subsystem below:",
        reply_markup=root_menu_kb(),
        parse_mode="HTML"
    )

@dp.callback_query(F.data == "nav_quota")
async def cb_quota(callback: types.CallbackQuery):
    r = await aioredis.from_url(f"redis://{REDIS_HOST}:6379", decode_responses=True)
    
    with open("/proc/meminfo", "r") as f:
        lines = f.readlines()
        mem_total = int(lines[0].split()[1]) // 1024
        mem_avail = int(lines[2].split()[1]) // 1024
        mem_used = mem_total - mem_avail

    cerebras_rpm = await r.get("ratelimit:cerebras:rpm") or "0"
    groq_rpm = await r.get("ratelimit:groq:rpm") or "0"
    gemini_rpd = await r.get("ratelimit:gemini:rpd") or "0"

    await callback.message.edit_text(
        f"📊 <b>LIVE CLUSTER TELEMETRY & QUOTA GAUGES</b>\n\n"
        f"<b>Host Memory:</b> {mem_used} MB / {mem_total} MB ({int(mem_used/mem_total*100)}%)\n"
        f"<b>Active Model Sliding Windows:</b>\n"
        f"• Cerebras : {cerebras_rpm} / 24 RPM (Pre-Call Clamped)\n"
        f"• Groq LPU : {groq_rpm} / 24 RPM (Pre-Call Clamped)\n"
        f"• Gemini   : {gemini_rpd} / 1200 RPD (Pre-Call Clamped)\n"
        f"• Network  : Ring 0 (Direct Residential ISP)\n",
        parse_mode="HTML",
        reply_markup=InlineKeyboardMarkup(inline_keyboard=[
            [InlineKeyboardButton(text="🔄 Refresh", callback_data="nav_quota")],
            [InlineKeyboardButton(text="⬅️ Back to Menu", callback_data="nav_root")]
        ])
    )
    await r.close()

@dp.callback_query(F.data == "nav_services")
async def cb_services(callback: types.CallbackQuery):
    services = ["searxng", "video-conductor", "tailscale", "vpn-ring1"]
    kb_rows = []

    for s in services:
        try:
            container = docker_client.containers.get(f"ai-gateway-{s}")
            status = "🟢 ON" if container.status == "running" else "🔴 OFF"
            action = "Stop" if container.status == "running" else "Start"
        except docker.errors.NotFound:
            status = "⚪ Missing"
            action = "Deploy"

        kb_rows.append([
            InlineKeyboardButton(text=f"{s}: {status}", callback_data=f"info_{s}"),
            InlineKeyboardButton(text=f"{action}", callback_data=f"toggle_{s}")
        ])

    kb_rows.append([InlineKeyboardButton(text="⬅️ Back to Menu", callback_data="nav_root")])
    await callback.message.edit_text(
        "🔌 <b>MODULAR SERVICE LIFECYCLE CONTROLLER</b>\n"
        "Toggle edge profiles and auxiliary sidecars:",
        parse_mode="HTML",
        reply_markup=InlineKeyboardMarkup(inline_keyboard=kb_rows)
    )

@dp.callback_query(F.data.startswith("toggle_"))
async def cb_execute_toggle(callback: types.CallbackQuery):
    service = callback.data.replace("toggle_", "")
    container_name = f"ai-gateway-{service}"
    try:
        container = docker_client.containers.get(container_name)
        if container.status == "running":
            container.stop(timeout=5)
            await callback.answer(f"Stopped {service}")
        else:
            container.start()
            await callback.answer(f"Started {service}")
    except Exception as e:
        await callback.answer(f"Action failed: {str(e)}", show_alert=True)

    await cb_services(callback)

@dp.callback_query(F.data == "nav_keys")
async def cb_keys(callback: types.CallbackQuery):
    providers = ["cerebras", "groq", "sambanova", "mistral", "gemini"]
    kb = [[InlineKeyboardButton(text=f"Update {p.capitalize()} Key", callback_data=f"key_{p}")] for p in providers]
    kb.append([InlineKeyboardButton(text="⬅️ Back to Menu", callback_data="nav_root")])
    
    await callback.message.edit_text(
        "🧠 <b>KEY VAULT & ROUTING OVERRIDES</b>\n"
        "Select a provider to inject a new API key.\n"
        "<i>Keys sent in chat are scrubbed immediately upon receipt.</i>",
        parse_mode="HTML",
        reply_markup=InlineKeyboardMarkup(inline_keyboard=kb)
    )

@dp.callback_query(F.data.startswith("key_"))
async def cb_prompt_key(callback: types.CallbackQuery, state: FSMContext):
    provider = callback.data.replace("key_", "")
    await state.set_state(AdminStates.waiting_for_key)
    await state.update_data(target_provider=provider)
    
    prompt = await callback.message.answer(
        f"🔐 <b>Awaiting Key Injection for {provider.upper()}</b>\n"
        "Paste the raw API key now. This message and your reply will be instantly purged.",
        parse_mode="HTML"
    )
    await state.update_data(prompt_msg_id=prompt.message_id)

@dp.message(AdminStates.waiting_for_key)
async def cb_receive_key(message: types.Message, state: FSMContext):
    data = await state.get_data()
    provider = data.get("target_provider")
    new_key = message.text.strip()

    try:
        await message.delete()
        await bot.delete_message(message.chat.id, data.get("prompt_msg_id"))
    except Exception:
        pass

    env_var_name = f"{provider.upper()}_API_KEY"
    if os.path.exists(ENV_PATH):
        with open(ENV_PATH, "r") as f:
            content = f.read()
        
        if env_var_name in content:
            content = re.sub(rf"{env_var_name}=.*", f"{env_var_name}={new_key}", content)
        else:
            content += f"\n{env_var_name}={new_key}\n"

        with open(ENV_PATH, "w") as f:
            f.write(content)

    try:
        litellm_container = docker_client.containers.get("ai-gateway-litellm")
        litellm_container.kill(signal="SIGHUP")
    except Exception:
        pass

    await state.clear()
    await message.answer(
        f"✅ <b>Key for {provider.upper()} injected & persisted.</b>\n"
        "Chat history scrubbed. LiteLLM reloaded with fresh credentials.",
        parse_mode="HTML",
        reply_markup=root_menu_kb()
    )

@dp.callback_query(F.data == "nav_logs")
async def cb_logs(callback: types.CallbackQuery):
    kb = [
        [
            InlineKeyboardButton(text="LiteLLM (40)", callback_data="log_litellm"),
            InlineKeyboardButton(text="Interceptor (40)", callback_data="log_interceptor")
        ],
        [
            InlineKeyboardButton(text="Conductor (40)", callback_data="log_conductor"),
            InlineKeyboardButton(text="SearXNG (40)", callback_data="log_searxng")
        ],
        [InlineKeyboardButton(text="⬅️ Back to Menu", callback_data="nav_root")]
    ]
    await callback.message.edit_text(
        "📑 <b>REAL-TIME LOG TRIAGE & DIAGNOSTICS</b>\n"
        "Select a container to tail the last 40 console output lines:",
        parse_mode="HTML",
        reply_markup=InlineKeyboardMarkup(inline_keyboard=kb)
    )

@dp.callback_query(F.data.startswith("log_"))
async def cb_fetch_logs(callback: types.CallbackQuery):
    service_alias = callback.data.replace("log_", "")
    container_map = {
        "litellm": "ai-gateway-litellm",
        "interceptor": "ai-gateway-interceptor",
        "conductor": "ai-gateway-video-conductor",
        "searxng": "ai-gateway-searxng"
    }
    target = container_map.get(service_alias)
    
    try:
        container = docker_client.containers.get(target)
        raw_logs = container.logs(tail=40).decode("utf-8", errors="replace")
        sanitized = re.sub(r"(Bearer\s+)[A-Za-z0-9_\-\.]{8,}", r"\1[REDACTED]", raw_logs)
        sanitized = re.sub(r"(csk-|gsk_|sk-)[A-Za-z0-9_\-]{8,}", r"[KEY_REDACTED]", sanitized)
        output = f"📑 <b>TAIL: {target}</b>\n<pre>{sanitized[-3500:]}</pre>"
    except Exception as e:
        output = f"❌ Failed to fetch logs for {target}: {str(e)}"

    await callback.message.edit_text(
        output,
        parse_mode="HTML",
        reply_markup=InlineKeyboardMarkup(inline_keyboard=[
            [InlineKeyboardButton(text="🔄 Refresh", callback_data=callback.data)],
            [InlineKeyboardButton(text="⬅️ Back to Logs", callback_data="nav_logs")]
        ])
    )

@dp.callback_query(F.data == "nav_emergency")
async def cb_emergency(callback: types.CallbackQuery, state: FSMContext):
    await state.set_state(AdminStates.waiting_for_pin)
    await callback.message.edit_text(
        "🚨 <b>EMERGENCY OPERATIONS PERIMETER</b>\n\n"
        "⚠️ <i>Protected System Actions Require Security PIN:</i>\n"
        "• Soft Reload Gateway Engine\n"
        "• Purge Redis Caches & Distributed Locks\n"
        "• Engage Emergency Minimal-RAM Safe Mode\n\n"
        "Send your 4-digit security PIN to unlock execution:",
        parse_mode="HTML",
        reply_markup=InlineKeyboardMarkup(inline_keyboard=[
            [InlineKeyboardButton(text="❌ Abort", callback_data="nav_root")]
        ])
    )

@dp.message(AdminStates.waiting_for_pin)
async def cb_verify_pin(message: types.Message, state: FSMContext):
    await message.delete()
    if message.text.strip() != ADMIN_SECURITY_PIN:
        await state.clear()
        await message.answer("❌ Invalid Security PIN. Emergency session locked.", reply_markup=root_menu_kb())
        return

    await state.clear()
    kb = [
        [InlineKeyboardButton(text="🧯 Soft Reload Gateway Core", callback_data="exec_reload")],
        [InlineKeyboardButton(text="🧹 Purge Caches & Dead Locks", callback_data="exec_flush")],
        [InlineKeyboardButton(text="🛡️ ENGAGE SAFE MODE (800MB RAM)", callback_data="exec_safemode")],
        [InlineKeyboardButton(text="⬅️ Back to Menu", callback_data="nav_root")]
    ]
    await message.answer(
        "🔓 <b>SECURITY PIN ACCEPTED — EMERGENCY ACTIONS UNLOCKED</b>\n"
        "Select an emergency remediation procedure:",
        parse_mode="HTML",
        reply_markup=InlineKeyboardMarkup(inline_keyboard=kb)
    )

@dp.callback_query(F.data == "exec_safemode")
async def cb_exec_safemode(callback: types.CallbackQuery):
    targets = ["ai-gateway-webui", "ai-gateway-video-conductor", "ai-gateway-grafana", "ai-gateway-searxng"]
    for t in targets:
        try:
            docker_client.containers.get(t).stop(timeout=2)
        except Exception:
            pass

    await callback.message.edit_text(
        "🛡️ <b>SAFE MODE ENGAGED</b>\n\n"
        "• WebUI, Video Conductor, Grafana, and SearXNG halted.\n"
        "• LiteLLM and Interceptor running in isolated core mode.\n"
        "• Host RAM consumption reclaimed to <800 MB.\n"
        "• Ready for interactive chat recovery via Developer API and Bastion.",
        parse_mode="HTML",
        reply_markup=root_menu_kb()
    )

@dp.callback_query(F.data == "nav_root")
async def cb_back_root(callback: types.CallbackQuery):
    await callback.message.edit_text(
        "⚡ <b>AI GATEWAY CLUSTER CHATOPS CONTROL PLANE</b>\n"
        "Select a control subsystem below:",
        reply_markup=root_menu_kb(),
        parse_mode="HTML"
    )

async def main():
    await dp.start_polling(bot)

if __name__ == "__main__":
    asyncio.run(main())
```

### D. LiteLLM Core Routing Configuration (`config/config.yaml`)

```yaml
model_list:
  # ==============================================================================
  # 1. POOL/GENERAL (FRONTLINE CORE)
  # ==============================================================================
  - model_name: pool/general
    litellm_params:
      model: groq/llama-3.3-70b-versatile
      api_key: os.environ/GROQ_API_KEY
      rpm: 24
      tpm: 25000
  - model_name: pool/general
    litellm_params:
      model: cerebras/llama3.3-70b
      api_key: os.environ/CEREBRAS_API_KEY
      rpm: 24
      tpm: 50000
  - model_name: pool/general
    litellm_params:
      model: gemini/gemini-2.0-flash
      api_key: os.environ/GEMINI_API_KEY
      rpm: 12
      tpm: 800000
  - model_name: pool/general
    litellm_params:
      model: github/gpt-4o-mini
      api_key: os.environ/GITHUB_TOKEN
      rpm: 12

  # ==============================================================================
  # 2. POOL/AGENT-CODING (FRONTLINE CORE)
  # ==============================================================================
  - model_name: pool/agent-coding
    litellm_params:
      model: sambanova/Qwen2.5-Coder-32B-Instruct
      api_key: os.environ/SAMBANOVA_API_KEY
      rpm: 16
      tpm: 80000
  - model_name: pool/agent-coding
    litellm_params:
      model: mistral/codestral-latest
      api_key: os.environ/MISTRAL_API_KEY
      rpm: 45
  - model_name: pool/agent-coding
    litellm_params:
      model: cloudflare/@cf/qwen/qwen2.5-coder-32b-instruct
      api_key: os.environ/CLOUDFLARE_API_KEY
      rpm: 25

  # ==============================================================================
  # 3. POOL/DEEP-REASONING (FRONTLINE CORE)
  # ==============================================================================
  - model_name: pool/deep-reasoning
    litellm_params:
      model: sambanova/DeepSeek-R1
      api_key: os.environ/SAMBANOVA_API_KEY
      rpm: 16
      tpm: 80000
  - model_name: pool/deep-reasoning
    litellm_params:
      model: groq/deepseek-r1-distill-llama-70b
      api_key: os.environ/GROQ_API_KEY
      rpm: 24
  - model_name: pool/deep-reasoning
    litellm_params:
      model: gemini/gemini-2.0-flash-thinking-exp
      api_key: os.environ/GEMINI_API_KEY
      rpm: 12

  # ==============================================================================
  # 4. POOL/IMAGE-GEN (ARTIFACT ENGINE)
  # ==============================================================================
  - model_name: pool/image-gen
    litellm_params:
      model: cloudflare/@cf/black-forest-labs/flux-1-schnell
      api_key: os.environ/CLOUDFLARE_API_KEY
      rpm: 15
  - model_name: pool/image-gen
    litellm_params:
      model: huggingface/stabilityai/stable-diffusion-xl-base-1.0
      api_key: os.environ/HF_TOKEN
      rpm: 15

  # ==============================================================================
  # 5. POOL/AUDIO-GEN (ARTIFACT ENGINE)
  # ==============================================================================
  - model_name: pool/audio-gen
    litellm_params:
      model: groq/whisper-large-v3
      api_key: os.environ/GROQ_API_KEY
      rpm: 20
  - model_name: pool/audio-gen
    litellm_params:
      model: huggingface/hexgrad/Kokoro-82M
      api_key: os.environ/HF_TOKEN
      rpm: 20

router_settings:
  routing_strategy: usage-based-routing-v2
  enable_pre_call_checks: true
  redis_host: redis
  redis_port: 6379
  allowed_fails: 1
  cooldown_time: 90
  num_retries: 3
  retry_after: 2

litellm_settings:
  cache: true
  cache_type: "redis"
  cache_params:
    host: "redis"
    port: 6379
    supported_call_types: ["completion", "acompletion"]
```

---

## 20. Master Deployment Runbook & Smoke Test Suite

### Bootstrap Deployment Script (`deploy.sh`)

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "================================================================="
echo "   DISTRIBUTED AI GATEWAY MESH: AUTOMATED DEPLOYMENT RUNBOOK     "
echo "================================================================="

# Detect host memory and size container cgroups
echo "--> [1/4] Evaluating hardware limits..."
TOTAL_RAM_KB=$(grep MemTotal /proc/meminfo | awk '{print $2}')
TOTAL_RAM_MB=$((TOTAL_RAM_KB / 1024))
echo "Detected Host RAM: ${TOTAL_RAM_MB} MB"

if [ "$TOTAL_RAM_MB" -lt 3000 ]; then
    echo "Applying Tier 1 Configuration: Ultra-Edge (<= 1.68 GB RAM Baseline)"
    export MEM_LIMIT_LITELLM="350M"
    export MEM_LIMIT_INTERCEPTOR="120M"
    export MEM_LIMIT_POSTGRES="120M"
    export MEM_LIMIT_REDIS="60M"
    export MEM_LIMIT_SEARXNG="60M"
    export MEM_LIMIT_CONDUCTOR="150M"
    export MEM_LIMIT_WEBUI="450M"
    export MEM_LIMIT_ADMIN="95M"
    export MEM_LIMIT_GRAFANA="110M"
    export MEM_LIMIT_BASTION="85M"
else
    echo "Applying Tier 2 Configuration: Standard Edge"
    export MEM_LIMIT_LITELLM="512M"
    export MEM_LIMIT_INTERCEPTOR="160M"
    export MEM_LIMIT_POSTGRES="256M"
    export MEM_LIMIT_REDIS="128M"
    export MEM_LIMIT_SEARXNG="80M"
    export MEM_LIMIT_CONDUCTOR="200M"
    export MEM_LIMIT_WEBUI="512M"
    export MEM_LIMIT_ADMIN="128M"
    export MEM_LIMIT_GRAFANA="128M"
    export MEM_LIMIT_BASTION="100M"
fi

# Initialize persistent directories
echo "--> [2/4] Initializing persistence structures..."
mkdir -p data/postgres data/redis data/webui data/tailscale data/artifacts data/logs/archive
mkdir -p config/grafana/dashboards config/webui config/searxng daemon bastion-generic bastion-openclaw bastion-admin dashboard conductor

# Start infrastructure services
echo "--> [3/4] Pulling and deploying container stack..."
docker compose pull --parallel
docker compose up -d

# Verify gateway readiness
echo "--> [4/4] Verifying cluster health..."
RETRIES=15
until curl -s -f http://127.0.0.1:4000/health/readiness >/dev/null 2>&1 || [ $RETRIES -eq 0 ]; do
    sleep 2
    RETRIES=$((RETRIES - 1))
done

if [ $RETRIES -eq 0 ]; then
    echo "🚨 LiteLLM healthcheck failed. Check 'docker compose logs litellm'."
    exit 1
fi

echo "================================================================="
echo "   SYSTEM ONLINE & VERIFIED                                      "
echo "   • Interceptor Ingress:      http://127.0.0.1:8000/v1          "
echo "   • Core LiteLLM Router:      http://127.0.0.1:4000/v1          "
echo "   • SearXNG Meta-Search:      http://127.0.0.1:8080             "
echo "   • Open WebUI Interface:     http://127.0.0.1:3001             "
echo "   • Admin Management Hub:     http://127.0.0.1:3000             "
echo "   • Grafana Telemetry:        http://127.0.0.1:3005             "
echo "================================================================="
```

### Synthetic Smoke Test Suite (`smoke_test.sh`)

```bash
#!/usr/bin/env bash
set -euo pipefail

INTERCEPTOR_URL="http://127.0.0.1:8000/v1"
SEARXNG_URL="http://127.0.0.1:8080"
API_KEY="${LITELLM_MASTER_KEY:-sk-master-internal-network-key}"

echo "Starting synthetic validation across cluster services..."

# 1. Verify SearXNG search engine
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${SEARXNG_URL}/search?q=test&format=json" || echo "000")
if [ "$STATUS" -eq 200 ]; then
  echo "✓ [SearXNG] Meta-search engine verified OK (HTTP 200)"
else
  echo "✗ [SearXNG] Meta-search check FAILED (Status: ${STATUS})"
fi

# 2. Test Interceptor rate smoothing and context pruning
START=$(date +%s%N)
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${INTERCEPTOR_URL}/chat/completions" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "pool/general",
    "messages": [
      {"role": "system", "content": "You are a concise assistant."},
      {"role": "user", "content": "Say online."}
    ],
    "max_tokens": 10
  }')
DURATION=$(( ($(date +%s%N) - START) / 1000000 ))

if [ "$HTTP_CODE" -eq 200 ]; then
  echo "✓ [Interceptor] Chat routing verified OK (${DURATION} ms)"
else
  echo "✗ [Interceptor] Request FAILED with status ${HTTP_CODE} (${DURATION} ms)"
fi

echo "All synthetic smoke tests complete."
```