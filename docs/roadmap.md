# TokMan - Ultimate AI Orchestration: Master Engineering Roadmap

## 1. Executive Summary & Vision

**TokMan - Ultimate AI Orchestration** is an autonomous, self-healing, zero-trust private AI cluster engineered to aggregate, arbitrate, and federate free-tier LLM endpoints, meta-search grounding, multimodal synthesis, and media production across consumer edge hardware and cloud hyperscalers.

This roadmap provides a complete, 1:1 mapping of all 20 architectural sections from [docs/report.md](file:///home/gamers/dev/tokman/docs/report.md) into 8 self-contained, testable engineering modules. Every capability pool, service, security rule, and hardware profile is accounted for.

---

## 2. Master Specification Mapping Table (from `docs/report.md`)

| Section in `docs/report.md` | Core Topic / Specification | Target Implementation Module | Detailed Spec File |
| :--- | :--- | :--- | :--- |
| **Section 1** | Executive Summary, Axioms & Zero-Trust Perimeter | Architecture-wide Foundation | [All Modules](file:///home/gamers/dev/tokman/docs/modules/) |
| **Section 2** | Master Hardware Sizing & Dynamic Resource Profiles | Resource cgroups & auto-tuning | [Module 1](file:///home/gamers/dev/tokman/docs/modules/module-1-core-routing-persistence.md) |
| **Section 3** | Master Production Infrastructure Stack (`compose.yml`) | Service container definitions & profiles | [All Modules](file:///home/gamers/dev/tokman/docs/modules/) |
| **Section 4** | Upstream Free-Tier Providers & Anti-Sybil Constraints | Quota ceilings & provider verification | [Module 1](file:///home/gamers/dev/tokman/docs/modules/module-1-core-routing-persistence.md), [Module 3](file:///home/gamers/dev/tokman/docs/modules/module-3-autonomous-supervisor-capability-pools.md) & [Onboarding Spec](file:///home/gamers/dev/tokman/docs/key-pool-model-onboarding.md) |
| **Section 5** | Master Capability Catalog & 14-Pool Architecture (144 Slots) | 12 pools + 2 meta-orchestrators | [Module 3](file:///home/gamers/dev/tokman/docs/modules/module-3-autonomous-supervisor-capability-pools.md) & [Onboarding Spec](file:///home/gamers/dev/tokman/docs/key-pool-model-onboarding.md) |
| **Section 6** | Concentric Quota Rings & Horizontal Scaling | Ring 0 Host IP, Rings 1–9 Gluetun VPNs | [Module 8](file:///home/gamers/dev/tokman/docs/modules/module-8-zero-trust-mesh-vpn-rings.md) |
| **Section 7** | Native Non-VPN Anti-Ban & Single-IP Traffic Shaping | Delay floors, jitter, 80% headroom checks | [Module 2](file:///home/gamers/dev/tokman/docs/modules/module-2-fastapi-interceptor-traffic-shaper.md) |
| **Section 8** | Multi-User Concurrency, P0–P3 Priority Queuing & FIFO | Identity normalization & leaky-bucket RPM | [Module 2](file:///home/gamers/dev/tokman/docs/modules/module-2-fastapi-interceptor-traffic-shaper.md) |
| **Section 9** | Context Consistency, `<think>` Quarantine & Pruning | Regex quarantine & sliding window pruning | [Module 2](file:///home/gamers/dev/tokman/docs/modules/module-2-fastapi-interceptor-traffic-shaper.md) |
| **Section 10** | Autonomous Agentic Supervisor (`pool/auto`) | Sub-40ms intent classifier & DAG handoffs | [Module 3](file:///home/gamers/dev/tokman/docs/modules/module-3-autonomous-supervisor-capability-pools.md) |
| **Section 11** | Deterministic Video Conductor (`pool/video-conductor`) | Flux.1, Kokoro-82M, Whisper, FFmpeg Ken Burns | [Module 6](file:///home/gamers/dev/tokman/docs/modules/module-6-media-conductor-artifact-engines.md) |
| **Section 12** | Headless Multimedia Artifact Sidecars | Typst documents, Marp/Reveal.js slides | [Module 6](file:///home/gamers/dev/tokman/docs/modules/module-6-media-conductor-artifact-engines.md) |
| **Section 13** | Local Search Grounding Engine (SearXNG Configuration) | Private meta-search & residential IP protection | [Module 4](file:///home/gamers/dev/tokman/docs/modules/module-4-search-grounding-open-webui.md) |
| **Section 14** | Universal Client-Side Selective Chat Importer | In-browser parser for 20+ chat archive formats | [Module 4](file:///home/gamers/dev/tokman/docs/modules/module-4-search-grounding-open-webui.md) |
| **Section 15** | Tri-Tier Chat Access Architecture | Decoupled assistant, agent & ChatOps | [Module 5](file:///home/gamers/dev/tokman/docs/modules/module-5-tri-tier-chat-bastions-chatops.md) |
| **Section 16** | Deterministic Admin ChatOps Menu & Out-of-Band Ops | 0-token mobile control plane, PIN safe mode | [Module 5](file:///home/gamers/dev/tokman/docs/modules/module-5-tri-tier-chat-bastions-chatops.md) |
| **Section 17** | Strict Human-in-the-Loop Stack Optimizer | 7-day token burn analyzer & staged diffs | [Module 7](file:///home/gamers/dev/tokman/docs/modules/module-7-stack-optimizer-observability.md) |
| **Section 18** | Linux Host Hardening, Sockets, and Compaction | Kernel sysctl tuning & Zstd L10 log compaction | [Module 1](file:///home/gamers/dev/tokman/docs/modules/module-1-core-routing-persistence.md) |
| **Section 19** | Complete Implementation Codebase | Python, JS, YAML, and Shell services | [All Modules](file:///home/gamers/dev/tokman/docs/modules/) |
| **Section 20** | Master Deployment Runbook & Smoke Test Suite | Automated deployment & verification suite | [All Modules](file:///home/gamers/dev/tokman/docs/modules/) |

---

## 3. Architectural Dependency & Execution Flow (Native Go Single-Binary Production)

```text
[Module 1: Core Go Gateway, Routing & Embedded Persistence]
       │ (Native Fyne v2 GUI, Go HTTP Gateway :8000, Embedded SQLite & LRU Cache)
       ▼
[Module 2: Native Go Stream Filter & Traffic Shaper]
       │ (Stateful <think> SSE Stream Sanitizer, P0–P3 Leaky Bucket, Jitter, Pruning)
       ▼
[Module 3: Autonomous Supervisor & 14-Pool Federation]
       │ (Sub-40ms Intent Classifier, 144 Model Slots, Tool DAG Handoff, Critic Loop)
       │
       ├──► [Module 4: Search Grounding & Universal WebUI]
       │           │ (Native Fyne GUI Dashboard & Stitch Web Admin, In-Browser Universal Importer, SearXNG)
       │           ▼
       │    [Module 5: Tri-Tier Chat Bastions & Admin ChatOps]
       │           │ (Outbound Telegram/WA/Discord Bastions, 0-Token Admin Menu, Safe Mode)
       │           ▼
       │    [Module 7: Stack Optimizer & Embedded Observability]
       │             (Embedded Spend Telemetry, SQLite Proposals, Real-time GUI Gauges)
       │
       └──► [Module 6: Video Conductor & Artifact Engines]
                   │ (FIFO Channel, Flux.1 + Kokoro + Whisper + FFmpeg Ken Burns, Typst, Marp)
                   ▼
            [Module 8: Zero-Trust Perimeter & Concentric Quota Rings]
                     (Tailscale Ingress, Multi-Ring Outbound Dialers, 429 IP Rotation)
```

---

## 4. Master Repository Directory Structure (Native Go Production Architecture)

```text
tokman/
├── .env                          # Local secrets & API keys (git-ignored)
├── .env.example                  # Template environment variables
├── .gitignore                    # Git ignore rules for data, keys, and caches
├── compose.yml                   # Enterprise Cluster Profile reference configuration
├── main.go                       # Standalone application entrypoint (CLI daemon or GUI launch)
│
├── backend/                      # Production Go Backend Core
│   ├── config/                   # Configuration & environment loaders (dotenv.go)
│   ├── gateway/                  # High-performance HTTP reverse proxy (:8000), /v1/chat/completions
│   ├── filter/                   # Stateful <think> SSE stream sanitizer, dynamic context pruner
│   ├── registry/                 # Centralized 14-capability pool & model catalog (Single Source of Truth)
│   ├── shaper/                   # P0-P3 leaky-bucket rate limiter, domain jitter engine
│   ├── supervisor/               # Sub-40ms intent classifier, DAG handoff, critic loop
│   ├── storage/                  # Embedded pure-Go SQLite persistence, LRU completion cache
│   ├── types/                    # Shared core domain types (ChatMessage, PoolInfo)
│   ├── bastion/                  # ChatOps mobile control plane, 0-token admin menu, Telegram bot
│   ├── conductor/                # FIFO media queue, FFmpeg Ken Burns runner, Typst/Marp pipeline
│   ├── optimizer/                # 7-day token burn analyzer, non-destructive proposal engine
│   └── search/                   # SearXNG client & web grounding provider
│
├── gui/                          # Pure Native Desktop GUI (Fyne v2, Stitch Theme & 4 Master Tabs)
│   ├── app.go                    # GUI runtime lifecycle, window orchestration & tab assembly
│   ├── theme.go                  # Authentic Stitch-compliant dark theme tokens & font typography
│   ├── tab_dashboard.go          # System overview, cluster health, and real-time activity
│   ├── tab_endpoints.go          # Master 14-pool service catalog & provider endpoint routing
│   ├── tab_console.go            # Interactive chat playground, model selector & latency inspector
│   ├── tab_cache.go              # In-memory LRU completion cache ledger & key inspector
│   └── tab_settings.go           # Local network binding, security keys & live .env reload
│
├── docs/                         # Project architecture, planning & tracking docs
│   ├── report.md                 # Architectural Master Specification (GFM)
│   ├── roadmap.md                # Engineering Roadmap & Acceptance Criteria
│   ├── progress.md               # Progress Tracker & Viability Gap Log
│   └── modules/                  # Deep-dive modular specifications
│       ├── module-1-core-routing-persistence.md
│       ├── module-2-fastapi-interceptor-traffic-shaper.md
│       ├── module-3-autonomous-supervisor-capability-pools.md
│       ├── module-4-search-grounding-open-webui.md
│       ├── module-5-tri-tier-chat-bastions-chatops.md
│       ├── module-6-media-conductor-artifact-engines.md
│       ├── module-7-stack-optimizer-observability.md
│       └── module-8-zero-trust-mesh-vpn-rings.md
│
├── scripts/                      # Build & verification automation
│   ├── build.sh                  # Runs wails build to produce standalone binary
│   └── test_module1.sh           # Test harness runner
│
├── tests/                        # Automated Testing Suite (Go tests + Mock Harness)
│   ├── run_tests.sh              # CLI test orchestrator (--unit, --mock, --live)
│   ├── mocks/                    # Zero-credential offline upstream mock servers (Go/Python)
│   ├── unit/                     # Tier 1: Pure Go router, filter, and limiter unit tests
│   ├── integration/              # Tier 2: Service contract & proxy forwarding tests
│   └── e2e/                      # Tier 3: Standalone binary end-to-end sanity tests
│
└── data/                         # Local persistence (git-ignored)
    ├── tokman.db                 # Embedded SQLite database (spend logs, audits)
    └── artifacts/                # Generated media (videos, images, audio, PDF)
```

---

## 5. Comprehensive Module Specifications Summary

### [Module 1: Core Routing, State Persistence & Host Hardening](file:///home/gamers/dev/tokman/docs/modules/module-1-core-routing-persistence.md)
- **Services:** Wails Runtime, Go HTTP Gateway (`:8000`), Embedded SQLite (`data/tokman.db`), In-Memory LRU Cache.
- **Deliverables:** `wails.json`, `main.go`, `app.go`, `backend/gateway/`, `backend/storage/`, `.env.example`.
- **Hardware Budget:** **< 50 MB RAM** total baseline (vs 896MB in Docker).
- **Manual Verification Acceptance:** Single binary build (`wails build`), `/health/readiness`, Groq prompt completion (`pool/general`), sub-1ms local cache hit, SQLite spend record verified.

### [Module 2: Native Go Stream Filter & Traffic Shaper](file:///home/gamers/dev/tokman/docs/modules/module-2-fastapi-interceptor-traffic-shaper.md)
- **Services:** Go Middleware Pipeline (`backend/interceptor/`, `backend/filter/`, `backend/shaper/`)
- **Deliverables:** Stateful SSE `<think>` token filter, P0–P3 leaky-bucket limiter, dynamic asymmetric sliding-window pruning, inter-request jitter.
- **Hardware Budget:** In-process Go runtime (< 10 MB additional).
- **Manual Verification Acceptance:** `<think>` tags completely stripped from downstream assistant history, rate limit 429 returns on quota exceed, context pruned within token limits.

### [Module 3: Autonomous Supervisor Engine (`pool/auto`) & 14-Pool Federation](file:///home/gamers/dev/tokman/docs/modules/module-3-autonomous-supervisor-capability-pools.md)
- **Services:** Go Supervisor Engine (`backend/supervisor/`)
- **Deliverables:** Sub-40ms zero-shot classifier, full 14-pool / 144-model slot Go registry, ephemeral tool scratchpad, `pool/security-tester` critic loop.
- **Hardware Budget:** In-process Go runtime (< 15 MB).
- **Manual Verification Acceptance:** Sub-40ms classification latency, automated routing of coding prompts to `pool/agent-coding`, reasoning proofs to `pool/deep-reasoning`, critic pass execution.

### [Module 4: Local Search Grounding & Universal WebUI Interface](file:///home/gamers/dev/tokman/docs/modules/module-4-search-grounding-open-webui.md)
- **Services:** Embedded Wails WebUI Frontend (`frontend/`), SearXNG Aggregator Client (`backend/search/`)
- **Deliverables:** Embedded single-page dashboard, Universal Client-Side Selective Chat Importer modal (`frontend/src/importer.js`), Web grounding citation pipeline.
- **Hardware Budget:** Handled by native OS WebKitGTK (~30–40 MB).
- **Manual Verification Acceptance:** Wails dashboard renders cleanly, search grounding returns citations, selective chat importer modal parses 20+ archive formats client-side.

### [Module 5: Tri-Tier Chat Bastions & Deterministic Admin ChatOps](file:///home/gamers/dev/tokman/docs/modules/module-5-tri-tier-chat-bastions-chatops.md)
- **Services:** Go ChatOps Controller (`backend/bastion/`)
- **Deliverables:** Deterministic Admin ChatOps menu, host hardware telemetry parser (`/proc/meminfo`), ephemeral key injector, PIN-protected Safe Mode.
- **Hardware Budget:** In-process Go runtime (< 10 MB).
- **Manual Verification Acceptance:** 0-token mobile control plane, live host hardware telemetry, secret key injection with auto-purging, PIN-protected Safe Mode clamping RAM footprint.

### [Module 6: Deterministic Video Conductor & Multimedia Artifact Engines](file:///home/gamers/dev/tokman/docs/modules/module-6-media-conductor-artifact-engines.md)
- **Services:** Go Media Conductor (`backend/conductor/`)
- **Deliverables:** Single-worker FIFO channel processor, FFmpeg Ken Burns compositor, Kokoro-82M TTS + Whisper alignment pipeline, headless Typst and Marp renderers.
- **Hardware Budget:** Bounded burst limit (< 512 MB).
- **Manual Verification Acceptance:** Manifest generated within duration caps (<=30s/<=60s), smooth 1080p MP4 render with synchronized audio.

### [Module 7: Stack Optimizer (`pool/stack-optimizer`) & Observability Control Plane](file:///home/gamers/dev/tokman/docs/modules/module-7-stack-optimizer-observability.md)
- **Services:** Go Background Optimizer Daemon (`backend/optimizer/`), Embedded Telemetry UI
- **Deliverables:** 7-day token burn telemetry aggregator, SQLite proposal staging (`proposals` table), operator approval hot-reload, embedded real-time gauges.
- **Hardware Budget:** In-process Go runtime (< 10 MB).
- **Manual Verification Acceptance:** 7-day token burn aggregated from SQLite, non-destructive proposal staged, operator approval hot-reloads routing tables with 0 downtime.

### [Module 8: Zero-Trust Perimeter, Mesh Ingress & Concentric Quota Rings](file:///home/gamers/dev/tokman/docs/modules/module-8-zero-trust-mesh-vpn-rings.md)
- **Services:** Tailscale Mesh Ingress (`tailscale0` / embedded `tsnet`), Multi-Ring Egress HTTP Dialers
- **Deliverables:** Concentric Quota Rings (Ring 0 host IP, Rings 1–9 VPN egress shards), upstream 429 automated IP rotation.
- **Hardware Budget:** < 30 MB.
- **Manual Verification Acceptance:** WAN firewall ports sealed, remote access functional via Tailscale mesh IP, Ring 1 egress IP confirmed distinct from host Ring 0, simulated 429 cycles outbound dialer.

---

## 6. Continuous Automated Testing Framework

To ensure system sanity and prevent regressions as new code is introduced, the repository adopts a continuous automated testing framework defined in [docs/modules/testing-framework.md](file:///home/gamers/dev/tokman/docs/modules/testing-framework.md).

### Multi-Tier Testing Architecture:
1. **Tier 1 (Unit):** Pure algorithmic tests running in `< 3 seconds` (sliding window context math, `<think>` regex sanitization, duration clamping).
2. **Tier 2 (Mock E2E):** Zero-credential integration tests running against synthetic local LLM servers (`tests/mocks/mock_upstream.py`), allowing offline CI/CD without burning free API quotas.
3. **Tier 3 (Live Module E2E):** Live smoke tests verifying real container contracts (`tests/e2e/test_module<N>_*.py`).
4. **Tier 4 (Resiliency & Chaos):** Simulated HTTP 429 backoff, circuit breaker trip/recovery, and emergency Safe Mode memory reclamation.

### Incremental Test Addition Rule:
For every new module or functional change, developers must:
- Add unit tests in `tests/unit/` for all business logic.
- Add an E2E test in `tests/e2e/` verifying the module's primary acceptance criteria.
- Execute `./tests/run_tests.sh --module <N>` before signing off on manual verification gates.

