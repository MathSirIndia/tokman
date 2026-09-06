# Module 3: Autonomous Supervisor Engine (`pool/auto`) & 14-Pool Federation

## 1. Module Overview & Architectural Role

Module 3 delivers the cluster's intelligent routing core. It bridges user requests to the **Master 144-Model Capability Matrix** without requiring manual model selection.

When a client submits a request with `model: "pool/auto"`, the **Autonomous Supervisor**:
1. Executes a **Tier 1 Intent Classification** in `< 40ms` using high-speed edge endpoints (Groq / Cerebras Llama 3.3).
2. Maps the classified intent to one of the **12 Core Capability Pools**.
3. Isolates intermediate reasoning steps in an ephemeral scratchpad envelope.
4. Triggers secondary sub-task passes (such as `pool/security-tester` AST checks on generated code).
5. Streams real-time chat telemetry (in-place edits for Telegram/Discord, accordions for WebUI).

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

---

## 2. The Master 144-Model Capability Catalog (14 Pools)

The matrix maps to 62 distinct upstream provider endpoints across 26 foundational model architectures:

| Capability Pool | Purpose | Primary Tier (4 Models) | Backup Tier 1 (4 Models) | Backup Tier 2 (4 Models) |
| :--- | :--- | :--- | :--- | :--- |
| **1. pool/general** | Frontline Chat & Summary | groq/llama-3.3-70b, cerebras/llama-3.3-70b, sambanova/deepseek-v3, gemini/gemini-2.0-flash | github/gpt-4o-mini, sambanova/llama-3.3-70b, mistral/mistral-small, openrouter/gemini-flash | cloudflare/llama-3.3-70b, openrouter/free-models, huggingface/llama-3.3, github/phi-4 |
| **2. pool/deep-reasoning** | Math, Logic, Multi-turn CoT | sambanova/deepseek-r1, groq/deepseek-r1-70b, github/deepseek-r1, gemini/gemini-2.0-think | cloudflare/deepseek-r1, huggingface/deepseek-r1, sambanova/qwen-72b-inst, openrouter/deepseek-r1 | github/phi-4-reasoning, groq/llama-3.3-spec, mistral/mistral-large, openrouter/qwen-72b |
| **3. pool/agent-coding** | Refactoring, Debug, Syntax | sambanova/qwen-2.5-coder, mistral/codestral, github/gpt-4o-mini, gemini/gemini-2.0-flash | cloudflare/qwen-2.5-coder, huggingface/qwen-coder, sambanova/llama-3.3-70b, openrouter/qwen-coder | groq/llama-3.3-70b, cerebras/llama-3.3-70b, github/llama-3.3-70b, openrouter/deepseek-v3 |
| **4. pool/document-anal** | Long Context Analysis (1M) | gemini/gemini-2.0-flash, gemini/gemini-flash-lite, gemini/gemini-1.5-flash, openrouter/gemini-flash | groq/llama-3.3-70b, mistral/mistral-small, github/llama-3.3-70b, github/gpt-4o-mini | sambanova/llama-3.3-70b, sambanova/deepseek-v3, cerebras/llama-3.3-70b, openrouter/llama-70b |
| **5. pool/web-research** | SearXNG Grounded Search | cerebras/llama-3.3-70b, groq/llama-3.3-70b, gemini/gemini-2.0-flash, groq/llama-3.1-8b | github/phi-4, mistral/mistral-small, sambanova/llama-3.3-70b, github/gpt-4o-mini | cloudflare/llama-3.3, openrouter/free-models, huggingface/llama-3.3, openrouter/gemini-flash |
| **6. pool/presentation** | Marp & Reveal.js Slide Decks | groq/llama-3.3-70b, cerebras/llama-3.3-70b, github/gpt-4o-mini, gemini/gemini-2.0-flash | sambanova/llama-3.3-70b, mistral/mistral-small, github/llama-3.3-70b, sambanova/qwen-coder | openrouter/llama-70b, cloudflare/llama-3.3, huggingface/llama-3.3, openrouter/deepseek-v3 |
| **7. pool/image-gen** | Keyframe & Illustration | cf/flux-1-schnell, hf/flux-1-schnell, hf/sdxl-base-1.0, cf/sdxl-base-1.0 | hf/stable-diffusion-3.5, hf/sdxl-lightning, cf/sdxl-lightning, pollinations/flux-model | pollinations/turbo-gen, hf/stable-diffusion-1.5, cf/stable-diffusion-xl, pollinations/midjourney |
| **8. pool/architect** | Video Conductor Manifest Planner | gemini/gemini-2.0-flash, gemini/gemini-2.0-think, groq/llama-3.2-90b-vis, sambanova/deepseek-r1 | github/gpt-4o-mini, mistral/pixtral-12b, cf/llama-3.2-11b-vision, openrouter/gemini-flash | huggingface/llama-11b-v, gemini/gemini-1.5-flash, sambanova/llama-3.3-70b, cerebras/llama-3.3-70b |
| **9. pool/security-tester** | Automated Code Critic Loop | sambanova/qwen-coder, gemini/gemini-2.0-flash, mistral/codestral, github/gpt-4o-mini | cf/qwen-2.5-coder, groq/deepseek-r1-70b, mistral/pixtral-12b, openrouter/qwen-coder | sambanova/deepseek-r1, github/llama-3.3-70b, huggingface/qwen-coder, openrouter/deepseek-r1 |
| **10. pool/stack-optimiz** | Zero-Touch Tuning Evaluator | gemini/gemini-2.0-flash, sambanova/deepseek-r1, groq/llama-3.3-70b, github/gpt-4o-mini | cerebras/llama-3.3-70b, mistral/mistral-small, cf/llama-3.3-70b, openrouter/gemini-flash | github/phi-4, sambanova/qwen-72b-inst, huggingface/llama-3.3, openrouter/deepseek-v3 |
| **11. pool/document-gen** | Typst / Pandoc PDF Engines | gemini/gemini-2.0-flash, cerebras/llama-3.3-70b, github/gpt-4o-mini, gemini/gemini-flash-lite | groq/llama-3.3-70b, sambanova/llama-3.3-70b, mistral/codestral, github/llama-3.3-70b | mistral/mistral-large, openrouter/free-models, cf/llama-3.3-70b, huggingface/llama-3.3 |
| **12. pool/audio-gen** | TTS Narration & Whisper STT | cf/openai-whisper, hf/kokoro-82m-tts, groq/whisper-large-v3, cf/melo-tts-speech | groq/whisper-large-v3, hf/xtts-v2-speech, cf/fastspeech2-tts, huggingface/speecht5 | hf/mms-tts-speech, pollinations/audio-synth, deepgram/aura-free, huggingface/piper-tts |
| **13. pool/auto** | Meta-Orchestrator Arbitrator | Autonomous Intent Classifier | Tool Handoff DAG | Ephemeral Memory Envelope |
| **14. pool/video-conductor** | Meta-Orchestrator Video | Architect Manifest Generator | Image & Audio Synthesizer | FFmpeg Compositor |

---

## 3. Inherent Gaps in Base Specification (`docs/report.md`) & Resolutions

| # | Inherent Gap in `report.md` | Architectural Impact | Module 3 Engineering Resolution |
| :-: | :--- | :--- | :--- |
| **G-3.1** | **Missing Supervisor Implementation** | `report.md` documented `pool/auto` conceptually, but `interceptor.py` did not implement classification logic, causing `model: pool/auto` to fail with 400 Bad Request in LiteLLM. | Implement `daemon/supervisor.py` directly inside the Interceptor runtime. When `model == "pool/auto"`, the supervisor executes intent classification and rewrites the target model before forwarding. |
| **G-3.2** | **Classification Latency Budget Overruns** | If intent classification takes 500ms–1s, the interactive UX is ruined. | Use a specialized 1-token output classifier prompt (`GENERAL|CODE|REASON|SEARCH|DOC`) executed on Cerebras or Groq LPU to consistently guarantee `< 40ms` latency. |
| **G-3.3** | **Infinite Loop in Critic Pass** | If `pool/security-tester` rejects code, an uncontrolled loop can exhaust the user's rate limits. | Clamp the verification loop to maximum 1 pass (`MAX_CRITIC_LOOPS=1`). If the second pass fails, return code with security warnings flagged in markdown. |

---

## 4. Component Deliverables & Configuration

### A. Intent Classification Prompt Engine (`daemon/supervisor.py`)
```python
CLASSIFIER_PROMPT = """Analyze the user query. Output EXACTLY ONE token representing the routing pool:
GENERAL (default chat, conversation, summary)
REASON (math, complex proofs, multi-step logic)
CODE (programming, refactoring, bug fixing, SQL)
SEARCH (real-time facts, current news, URLs)
DOC (analyzing attached documents, large inputs > 32k chars)

Query: {prompt}
Output:"""

POOL_MAP = {
    "GENERAL": "pool/general",
    "REASON": "pool/deep-reasoning",
    "CODE": "pool/agent-coding",
    "SEARCH": "pool/web-research",
    "DOC": "pool/document-anal"
}
```

### B. Live Chat Telemetry Dispatcher
Streams live progress updates during multi-step execution:
- **Telegram / Discord:** Edits in-flight message status (`🔍 Searching SearXNG...` → `💻 Debugging syntax...` → `Final Output`).
- **WhatsApp:** Prepends a standardized status tag to the final synthesized response.
- **Open WebUI:** Emits collapsible markdown accordions (`:::status Thinking...`).

---

## 5. Verification & Acceptance Runbook (`scripts/test_module3.sh`)

### Automated Verification Steps:
1. **Classifier Speed Test:** Sends 10 diverse prompts and asserts classification completes in `< 40ms` average.
2. **Coding Intent Routing:** Sends `"Write a binary search tree in Rust"` with `model: pool/auto`, verifies LiteLLM logs dispatch to `pool/agent-coding`.
3. **Reasoning Intent Routing:** Sends `"Prove that the square root of 2 is irrational"`, verifies dispatch to `pool/deep-reasoning`.
4. **General Intent Routing:** Sends `"Tell me a short joke"`, verifies dispatch to `pool/general`.
5. **Security Critic Pass Verification:** Injects known vulnerable code (e.g. SQL injection), verifies `pool/security-tester` executes a secondary AST review pass.

### Pass/Fail Criteria:
- **PASS:** Accuracy >= 90% across prompt classes; average classification latency < 50ms; critic pass executes without hanging.
- **FAIL:** Classification timeout > 100ms, incorrect pool assignment, or unhandled 400 Bad Request.
