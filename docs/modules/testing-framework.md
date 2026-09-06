# Automated Testing Framework & Continuous Sanity Architecture

## 1. Overview & Testing Philosophy

To ensure reliability, prevent regressions across modular code additions, and verify working sanity without burning through free-tier API quotas, the Distributed AI Gateway Mesh includes a standardized, multi-tiered **Automated Testing Framework**.

In production, `tokman` compiles into a single native binary using **Go + Wails**. Therefore, tests are structured in two complementary layers:
1. **Go Native Testing Engine (`go test`):** High-speed unit, algorithmic, and in-memory integration testing running in `< 1 second`.
2. **Unified CLI Test Orchestrator (`tests/run_tests.sh`):** Executes Go tests, mock upstream integration suites, and end-to-end binary sanity tests across all modules.

```text
                                [ TEST ORCHESTRATOR: ./tests/run_tests.sh ]
                                                   │
          ┌────────────────────────┬───────────────┴───────────────┬────────────────────────┐
          ▼                        ▼                               ▼                        ▼
 ┌──────────────────┐    ┌──────────────────┐            ┌──────────────────┐     ┌──────────────────┐
 │ TIER 1: UNIT     │    │ TIER 2: MOCK E2E │            │ TIER 3: LIVE     │     │ TIER 4: RECOVERY │
 │ (Offline / Fast) │    │ (Zero-Cred E2E)  │            │ (Live Binary)    │     │ (Chaos & 429s)   │
 ├──────────────────┤    ├──────────────────┤            ├──────────────────┤     ├──────────────────┤
 │ • Context Math   │    │ • Mock Groq/Gem. │            │ • Wails Binary   │     │ • 429 Cooldown   │
 │ • `<think>` Strip│    │ • Router Proxy   │            │ • Cache Hits     │     │ • Jitter Timing  │
 │ • Identity Hash  │    │ • SQLite Audits  │            │ • Real Provider  │     │ • Safe Mode RAM  │
 │ • Token Buckets  │    │ • Classifier DAG │            │ • Actual Latency │     │ • Circuit Break  │
 └──────────────────┘    └──────────────────┘            └──────────────────┘     └──────────────────┘
```

---

## 2. Framework Directory Structure (`tests/`)

```text
tests/
├── run_tests.sh               # Interactive & CI CLI test runner (--unit, --module, --live, --all)
│
├── mocks/                     # Zero-credential offline upstream mock servers
│   ├── mock_upstream.go       # Go mock HTTP server mimicking Groq, Cerebras, and Gemini schemas
│   └── fixtures/              # Pre-recorded JSON completion responses & manifests
│
├── unit/                      # Tier 1: Pure Go algorithm, filter, and limiter tests
│   ├── filter_test.go         # Stateful <think> SSE stream tests
│   ├── limiter_test.go        # P0–P3 token-bucket and jitter math
│   ├── prune_test.go          # Context window preservation tests
│   └── storage_test.go        # In-memory LRU cache and SQLite schema tests
│
├── integration/               # Tier 2: Service contract & proxy forwarding tests
│   ├── router_test.go         # Model pool resolution and fallback tests
│   └── audit_test.go          # Transaction logging and spend audit tests
│
├── e2e/                       # Tier 3: Modular End-to-End Sanity Suites
│   ├── module1_core_test.go   # Gateway readiness, completions, sub-1ms cache
│   ├── module2_shaper_test.go # Stream filtering and rate limit responses
│   └── module3_super_test.go  # Autonomous intent routing tests
│
└── resiliency/                # Tier 4: Edge cases, rate limits, and failure handling
    ├── rate_limit_test.go     # Saturation and burst handling
    └── failover_test.go       # Circuit breaker recovery and 429 rotation
```

---

## 3. Four-Tier Testing Pyramid

### Tier 1: Offline Unit Testing
- **Execution:** `go test ./backend/... -v`
- **Speed:** `< 500 ms`
- **Scope:** Pure Go functions, regexes, SSE stream token chunking, and memory algorithms. Zero network calls or external processes required.

### Tier 2: Zero-Credential Mock E2E Testing
- **Execution:** `./tests/run_tests.sh --mock`
- **Speed:** `< 3 seconds`
- **Scope:** Spawns a lightweight local HTTP mock server (`tests/mocks/mock_upstream.go`) mimicking Groq/Gemini response structures. Validates that the Gateway connects, parses headers, caches completions, and writes SQLite records without requiring real API keys.

### Tier 3: Live Module Sanity Verification
- **Execution:** `./tests/run_tests.sh --module <N>`
- **Speed:** `~ 5–15 seconds`
- **Scope:** Runs against real upstream providers (e.g. Groq free tier) to verify live network round-trips, latency tracking, and cache hits.

### Tier 4: Resiliency & Chaos Testing
- **Execution:** `./tests/run_tests.sh --resiliency`
- **Speed:** `~ 10–30 seconds`
- **Scope:** Injects simulated HTTP 429s, network timeouts, and rapid burst requests to verify that rate limiters return HTTP 429 and circuit breakers back off cleanly.

---

## 4. Enforcement Rule Integration

Per **Rule 2 (`AGENTS.md`)**:
- All pull requests and module sign-offs require all Tier 1 unit tests to pass cleanly.
- Module verification gates require running `./tests/run_tests.sh --module <N>` with 100% assertions passing before logging results in `docs/progress.md`.
