# Module 2: Native Go Stream Filter & Traffic Shaper (Production Architecture)

## 1. Module Overview & Architectural Role

Module 2 implements the **Traffic Shaper, Stateful Stream Filter, and Rate Limiter** directly inside the `tokman` Go backend runtime. It acts as an in-process proxy pipeline between incoming HTTP requests and upstream provider dispatchers.

This module is responsible for:
1. **Unified Identity Normalization:** Maps client UUIDs, Telegram IDs, and API tokens to canonical user quota records.
2. **P0–P3 Fair-Share Priority Queuing:** Applies token-bucket rate limits according to client profile classes (Interactive 12 RPM, IDE 60 RPM, Research 5 RPM).
3. **Context Guard (`<think>` SSE Quarantine):** A stateful stream parser that strips or quarantines DeepSeek Chain-of-Thought tokens (`<think>...</think>`) in real-time during Server-Sent Events (SSE) streaming without buffering the full response.
4. **Dynamic Asymmetric Sliding-Window Pruning:** Prevents `HTTP 400 Bad Request` context overflows when switching from 1M to 8k window models by pruning whole message turns while preserving system prompts.
5. **Anti-Ban Traffic Smoothing:** Enforces an inter-request delay floor (150ms–350ms) and randomized jitter per upstream provider domain.

```text
                                [ CLIENT INGRESS ]
                   (WebUI UUIDs, Telegram IDs, IDE Bearer Tokens)
                                        │
                                        ▼ (HTTP :8000)
┌───────────────────────────────────────────────────────────────────────────────────────┐
│                     TOKMAN GO TRAFFIC SHAPER & STREAM PIPELINE                        │
│  ├── 1. Unified Identity Resolution (Header & Token Normalizer)                       │
│  ├── 2. Leaky-Bucket Rate Limiter (P0–P3 Fair-Share Allocation & 429 Generator)       │
│  ├── 3. Dynamic Asymmetric Context Pruning (Preserves System Prompt + Recent Turns)   │
│  ├── 4. Inter-Request Jitter Engine (150–350ms Domain Floor in Go Goroutines)         │
│  └── 5. Context Guard: Stateful SSE Stream Scanner (<think> Token Sanitizer)          │
└───────────────────────────────────────┬───────────────────────────────────────────────┘
                                        │ Cleaned Stream / JSON Payloads
                                        ▼
                             [ Upstream Provider Dispatch ]
```

---

## 2. Resource Allocation Across Sizing Stages

Because the pipeline is built into Go's native HTTP middleware using goroutines, it consumes minimal memory and avoids Python/FastAPI container overhead:

| Metric | Stage 1: Ultra-Edge<br>(< 3GB RAM) | Stage 2: Standard Edge<br>(3–6GB RAM) | Stage 3: Power Node<br>(6–12GB RAM) | Stage 4: Workstation<br>(12–24GB RAM) | Stage 5: Uncapped<br>(> 24GB RAM) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Go Pipeline Memory** | ~5 MB | ~10 MB | ~20 MB | ~40 MB | Host Managed |
| **Processing Latency** | < 0.2 ms | < 0.1 ms | < 0.05 ms | < 0.02 ms | Bare-metal wire speed |

---

## 3. Inherent Gaps in Base Specification (`docs/report.md`) & Resolutions

| # | Inherent Gap in `report.md` | Architectural Impact | Wails / Go Engineering Resolution |
| :-: | :--- | :--- | :--- |
| **G-2.1** | **Streaming `<think>` Leaks** | `interceptor.py` in Section 19 only stripped `<think>` on non-streaming responses. During SSE streaming, internal thinking tokens streamed directly into user history unredacted. | Implement a stateful streaming scanner (`ThinkingTokenFilter` in `backend/filter/think_filter.go`) that filters SSE delta chunks on the fly and suppresses `<think>...</think>` tokens. |
| **G-2.2** | **Race Conditions on Delay Floor** | Simple `r.get()` followed by `asyncio.sleep()` causes race conditions when multiple concurrent requests hit the same provider domain. | Use Go channels and an atomic mutex-guarded timestamp map per provider domain to pace dispatches deterministically. |
| **G-2.3** | **Context Truncation Mid-Sentence** | Character-based pruning can cut JSON or messages in half, causing syntax errors in model payloads. | Prune strictly at message turn boundaries (`[]ChatMessage`), never truncating individual turn strings midway. |
| **G-2.4** | **Python Container Bloat** | Running FastAPI in Docker required a separate container consuming 160MB RAM. | **Eliminated**: Go HTTP middleware runs inside the single `tokman` binary, saving 160MB RAM. |

---

## 4. Component Deliverables & Configuration

### A. Stateful Streaming Token Filter (`backend/filter/think_filter.go`)
- Stateful token scanner maintaining an internal state (`StateNormal`, `StateInsideThink`).
- Inspects SSE `data: {"choices":[{"delta":{"content":"..."}}]}` stream chunks.
- If `<think>` is encountered, suppresses content streaming until `</think>` is reached.
- Passes through all non-reasoning tokens with zero perceptible streaming latency (< 1ms).

### B. Leaky-Bucket Rate Limiter (`backend/shaper/limiter.go`)
- Tracks requests per minute (RPM) and tokens per minute (TPM) per user identity.
- Enforces priority classes:
  - **P0 (Interactive Chat):** 12 RPM, priority scheduling.
  - **P1 (IDE / Coding Agents):** 60 RPM burstable.
  - **P2 (Deep Research / Autonomous Loops):** 5 RPM paced.
  - **P3 (Media & Bulk Jobs):** Serialized FIFO.
- If quota exceeded, returns `HTTP 429 Too Many Requests` with `Retry-After` header.

### C. Dynamic Context Pruner (`backend/filter/prune.go`)
- Receives message turn array `[]ChatMessage` and target model token limit.
- Always protects:
  1. Turn 0: System prompt (`role: "system"`).
  2. Final turn: Active user prompt (`role: "user"`).
- Dynamically prunes intermediate conversation turns from oldest to newest until the token budget is met.

### D. Inter-Request Jitter Engine (`backend/shaper/jitter.go`)
- Maintains last request timestamp for each upstream domain (`api.groq.com`, `api.cerebras.ai`, etc.).
- Enforces minimum 150ms delay floor + crypto-random 0–200ms jitter before releasing outbound HTTP requests.

---

## 5. Testing & Verification Runbook

### Automated Go Tests
```bash
go test ./backend/filter/... ./backend/shaper/... -v
```
Tests cover:
1. `TestThinkFilter_Streaming`: Feeds mocked SSE streams containing `<think>` tags and verifies only clean tokens are emitted.
2. `TestRateLimiter_P0_P3`: Verifies that 13th P0 request in 60s returns HTTP 429.
3. `TestContextPruner_TurnPreservation`: Verifies system prompt is never pruned while older turns are cleanly dropped.

### Manual Verification Gate 2
1. Stream a reasoning prompt via curl to `pool/deep-reasoning`:
   ```bash
   curl -N http://localhost:8000/v1/chat/completions \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer sk-master-internal-network-key" \
     -d '{"model": "pool/deep-reasoning", "stream": true, "messages": [{"role": "user", "content": "What is 7 * 8?"}]}'
   ```
2. Verify output contains `56` and contains **no raw `<think>` tags**.
3. Fire 15 rapid requests and verify HTTP 429 is returned on quota overflow with proper headers.
