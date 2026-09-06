# Key, Model & Capability Pool Onboarding Specification

This document defines the unified architectural lifecycle and operational standard operating procedures (SOP) for onboarding **Upstream API Keys**, **Upstream AI Models**, and **Logical Capability Pools** across the Distributed AI Gateway Mesh (`tokman`).

---

## 1. Architectural Distribution Matrix

The onboarding process spans several modular layers of the gateway:

| Onboarding Domain | Module & Layer | Mechanism / Interface | Storage / State |
| :--- | :--- | :--- | :--- |
| **Provider API Keys** | **Module 1 (Desktop UI & Bootstrap)**<br>**Module 5 (ChatOps Key Vault)**<br>**Module 8 (Multi-Ring Sybil Scaling)** | • `.env` file bootstrap<br>• Wails Desktop UI Settings Tab<br>• Telegram 0-Token Ephemeral Key Injector<br>• Multi-Account Sets (Rings 0–9) | In-Memory Vault + Embedded SQLite (`data/tokman.db`), masked on disk and in logs |
| **Upstream Models** | **Module 3 (Master 144-Slot Catalog)**<br>**Module 7 (Stack Optimizer Telemetry)** | • Dynamic Model Catalog (`backend/gateway/catalog.go`)<br>• Provider payload adapter registration<br>• Auto-Tuning Proposal Staging | SQLite `model_registry` table with JSON fallback |
| **Capability Pools** | **Module 3 (Autonomous Supervisor)**<br>**Module 7 (Hot-Reloading)** | • 14 Standard Capability Pools<br>• 3-Tier Multi-Model Fallback Arrays (4×4×4)<br>• Sub-40ms Intent Classifier (`pool/auto`) | SQLite `capability_pools` table + Supervisor DAG |

---

## 2. Pillar 1: Upstream Provider API Key Onboarding

Operators configure API keys for the 10 free-tier upstream providers using four progressive tiers:

```text
               ┌────────────────────────────────────────────────────────┐
               │         TIER 1: STATIC BOOTSTRAP (.env file)           │
               │  • Bootstraps node with initial credentials on launch  │
               └───────────────────────────┬────────────────────────────┘
                                           │
               ┌───────────────────────────▼────────────────────────────┐
               │       TIER 2: NATIVE WAILS DESKTOP UI (Settings)       │
               │  • Live key update without restarting the Gateway      │
               └───────────────────────────┬────────────────────────────┘
                                           │
               ┌───────────────────────────▼────────────────────────────┐
               │    TIER 3: CHATOPS EPHEMERAL INJECTOR (Telegram Bot)   │
               │  • Operator sends key in DM -> probed -> chat shredded │
               └───────────────────────────┬────────────────────────────┘
                                           │
               ┌───────────────────────────▼────────────────────────────┐
               │      TIER 4: MULTI-RING SYBIL SCALING (Account Sets)   │
               │  • Account Sets 1–10 bound to Concentric Egress Rings  │
               └────────────────────────────────────────────────────────┘
```

### Supported Upstream Providers
1. **Groq LPU**: `GROQ_API_KEY` (30 RPM, Llama 3.3 70B, DeepSeek R1 Distill)
2. **Cerebras**: `CEREBRAS_API_KEY` (30 RPM / 60k TPM, Llama 3.3 70B)
3. **Google AI Studio**: `GEMINI_API_KEY` (15 RPM / 1.5k RPD, Gemini 2.0 Flash, Thinking, Flash-Lite)
4. **SambaNova Systems**: `SAMBANOVA_API_KEY` (20 RPM / 100k TPM, DeepSeek R1, DeepSeek V3, Qwen 2.5 Coder)
5. **GitHub Models**: `GITHUB_TOKEN` (15 RPM / 150 RPD, GPT-4o-mini, Phi-4)
6. **Mistral AI**: `MISTRAL_API_KEY` (1 RPS / 45 RPM, Codestral, Mistral Small)
7. **Cloudflare AI**: `CLOUDFLARE_API_TOKEN` & `CLOUDFLARE_ACCOUNT_ID` (10k Neurons/Day, Flux.1 Schnell, Whisper, MeloTTS)
8. **Hugging Face**: `HF_TOKEN` (Serverless soft caps, SDXL, Kokoro-82M TTS, Qwen Coder)
9. **OpenRouter**: `OPENROUTER_API_KEY` (20 RPM soft cap across `:free` models)
10. **Pollinations.ai**: Keyless / Public fallback

### Step-by-Step Key Onboarding SOP
1. **Via Configuration (`.env`):**
   - Add or update the key variable in `/home/gamers/dev/tokman/.env`.
   - The Gateway loads all present provider keys on initialization.
2. **Via Desktop Interface:**
   - Navigate to the **Settings** tab in the desktop application.
   - Enter the key under the corresponding provider field and click **Save Key**.
   - The key is instantly loaded into the runtime memory vault and persisted to `.env` / SQLite without dropping active connections.
3. **Via ChatOps (Module 5):**
   - In the Telegram Bastion bot, trigger `/admin` → `🧠 Models & Keys` → `🔑 Inject Ephemeral Key`.
   - Select the target provider and paste the key.
   - The bot fires a 1-token verification probe (`/health/probe`). Upon HTTP 200, the key is installed into the in-memory vault, and Telegram immediately deletes both the user prompt and the bot message to prevent token leakage in chat history.

---

## 3. Pillar 2: Upstream Model Onboarding

In **Module 3**, the hardcoded routing switch from Module 1 is replaced by the **Dynamic Model Catalog**.

### Model Registration Schema
Every upstream model registered in `tokman` implements the canonical schema:

```json
{
  "model_id": "llama-3.3-70b-versatile",
  "provider": "groq",
  "display_name": "Llama 3.3 70B Versatile",
  "context_window": 131072,
  "max_output_tokens": 8192,
  "rpm_ceiling": 30,
  "tpm_ceiling": 100000,
  "supports_streaming": true,
  "supports_tools": true,
  "endpoint_url": "https://api.groq.com/openai/v1/chat/completions",
  "auth_header_format": "Bearer {key}",
  "adapter_type": "openai_compatible"
}
```

### Model Adapters Supported
- `openai_compatible`: Groq, Cerebras, SambaNova, Mistral, GitHub Models, OpenRouter.
- `gemini_native`: Google AI Studio REST v1beta format.
- `cloudflare_workers_ai`: Cloudflare `@cf/...` payload mapper.
- `huggingface_router`: Hugging Face Inference API format.

### Step-by-Step Model Onboarding SOP
1. **Define Model Metadata:** Add the model entry to `backend/gateway/catalog.go` (or SQLite `model_registry`).
2. **Assign Provider & Rate Profile:** Specify the associated provider key and safety ceilings (RPM/TPM clamping at 80%).
3. **Slot into Capability Pools:** Assign the model to one or more capability pools as either Primary, Backup Tier 1, or Backup Tier 2.
4. **Run Unit & Integration Test:**
   ```bash
   go test ./backend/gateway -run TestModelRegistration -v
   ```

---

## 4. Pillar 3: Capability Pool Onboarding

A **Capability Pool** is an abstracted endpoint (`pool/<name>`) exposed to AI clients (Open WebUI, IDE plugins, CLI agents) that automatically balances, shapes traffic for, and fails over between multiple models.

### Master 14-Pool Standard Catalog (Module 3)

| Pool Identifier | Logical Specialization | Primary Tier (4 Slots) | Backup Tier 1 (4 Slots) | Backup Tier 2 (4 Slots) |
| :--- | :--- | :--- | :--- | :--- |
| `pool/general` | Conversational Chat & Summary | Groq Llama 3.3, Cerebras Llama 3.3, SambaNova DeepSeek V3, Gemini 2.0 Flash | GitHub GPT-4o-mini, SambaNova Llama 3.3, Mistral Small, OpenRouter Gemini | Cloudflare Llama 3.3, OpenRouter Free, HF Llama 3.3, GitHub Phi-4 |
| `pool/deep-reasoning` | Math Proofs, Logic, Deep CoT | SambaNova DeepSeek R1, Groq DeepSeek R1 70B, GitHub DeepSeek R1, Gemini Think | Cloudflare DeepSeek R1, HF DeepSeek R1, SambaNova Qwen 72B, OpenRouter R1 | GitHub Phi-4 Reasoning, Groq Speculative, Mistral Large, OpenRouter Qwen |
| `pool/agent-coding` | Syntax, Refactoring, Debugging | SambaNova Qwen 2.5 Coder, Mistral Codestral, GitHub GPT-4o-mini, Gemini Flash | Cloudflare Qwen Coder, HF Qwen Coder, SambaNova Llama 3.3, OpenRouter Qwen | Groq Llama 3.3, Cerebras Llama 3.3, GitHub Llama 3.3, OpenRouter V3 |
| `pool/document-anal` | Long Context Documents (> 32k) | Gemini 2.0 Flash (1M), Gemini Flash-Lite, Gemini 1.5 Flash, OpenRouter Gemini | Groq Llama 3.3, Mistral Small, GitHub Llama 3.3, GitHub GPT-4o-mini | SambaNova Llama 3.3, SambaNova V3, Cerebras Llama 3.3, OpenRouter Llama |
| `pool/web-research` | Real-time Search Grounding | Cerebras Llama 3.3, Groq Llama 3.3, Gemini 2.0 Flash, Groq Llama 3.1 8B | GitHub Phi-4, Mistral Small, SambaNova Llama 3.3, GitHub GPT-4o-mini | Cloudflare Llama 3.3, OpenRouter Free, HF Llama 3.3, OpenRouter Gemini |
| `pool/presentation` | Marp & Reveal.js Deck Gen | Groq Llama 3.3, Cerebras Llama 3.3, GitHub GPT-4o-mini, Gemini 2.0 Flash | SambaNova Llama 3.3, Mistral Small, GitHub Llama 3.3, SambaNova Qwen | OpenRouter Llama, Cloudflare Llama, HF Llama, OpenRouter V3 |
| `pool/image-gen` | Keyframe Synthesis (Flux / SD) | Cloudflare Flux.1 Schnell, HF Flux.1 Schnell, HF SDXL 1.0, Cloudflare SDXL | HF SD 3.5, HF SDXL Lightning, Cloudflare SDXL Lightning, Pollinations Flux | Pollinations Turbo, HF SD 1.5, Cloudflare SDXL, Pollinations Midjourney |
| `pool/architect` | Media Manifest Planner | Gemini 2.0 Flash, Gemini 2.0 Think, Groq Llama 3.2 90B Vision, SambaNova R1 | GitHub GPT-4o-mini, Mistral Pixtral 12B, Cloudflare Llama 3.2 11B, OpenRouter | HF Llama 11B Vision, Gemini 1.5 Flash, SambaNova Llama, Cerebras Llama |
| `pool/security-tester` | AST Review & Vulnerability Critic | SambaNova Qwen Coder, Gemini 2.0 Flash, Mistral Codestral, GitHub GPT-4o-mini | Cloudflare Qwen Coder, Groq DeepSeek R1, Mistral Pixtral 12B, OpenRouter | SambaNova R1, GitHub Llama 3.3, HF Qwen Coder, OpenRouter R1 |
| `pool/stack-optimiz` | Zero-Touch Tuning Evaluator | Gemini 2.0 Flash, SambaNova DeepSeek R1, Groq Llama 3.3, GitHub GPT-4o-mini | Cerebras Llama 3.3, Mistral Small, Cloudflare Llama 3.3, OpenRouter Gemini | GitHub Phi-4, SambaNova Qwen 72B, HF Llama 3.3, OpenRouter V3 |
| `pool/document-gen` | Typst & Pandoc PDF Generators | Gemini 2.0 Flash, Cerebras Llama 3.3, GitHub GPT-4o-mini, Gemini Flash-Lite | Groq Llama 3.3, SambaNova Llama 3.3, Mistral Codestral, GitHub Llama 3.3 | Mistral Large, OpenRouter Free, Cloudflare Llama 3.3, HF Llama 3.3 |
| `pool/audio-gen` | TTS Narration & Whisper STT | Cloudflare Whisper, HF Kokoro-82M TTS, Groq Whisper Large v3, Cloudflare Melo | Groq Whisper v3, HF XTTS-v2, Cloudflare FastSpeech2, HF SpeechT5 | HF MMS TTS, Pollinations Audio, Deepgram Aura, HF Piper TTS |
| `pool/auto` | Autonomous Meta-Orchestrator | Sub-40ms Groq/Cerebras Intent Classifier → Dynamic Pool Dispatch DAG |
| `pool/video-conductor` | Deterministic Video Pipeline | Architect Manifest → Flux Keyframes → Kokoro TTS → FFmpeg Camera Transform |

### Step-by-Step Pool Onboarding SOP (Adding a Custom Pool)
1. **Declare the Pool Configuration:**
   ```json
   {
     "pool_id": "pool/custom-finance",
     "description": "Financial modeling, numeric validation, and tabular extraction",
     "primary_models": ["groq/llama-3.3-70b", "sambanova/deepseek-r1"],
     "backup_tier_1": ["gemini/gemini-2.0-flash", "github/gpt-4o-mini"],
     "backup_tier_2": ["mistral/mistral-small"],
     "classifier_tag": "FINANCE"
   }
   ```
2. **Wire Intent Tag into Supervisor:**
   Update the single-token intent classifier prompt in `backend/supervisor/classifier.go` so `FINANCE` maps to `pool/custom-finance`.
3. **Register in Router:**
   The router exposes `pool/custom-finance` in `GET /v1/models` and accepts incoming chat requests automatically.

---

## 5. Summary of Implementation by Module

- **Module 1 (Completed):** Bootstrapped `pool/general` and `pool/deep-reasoning` with Groq key support; Settings UI key persistence; SQLite transaction audit.
- **Module 2 (Queued Next):** Traffic shaper leaky-bucket limiter and stateful `<think>` stream filter.
- **Module 3 (Queued After Mod 2):** Full Dynamic Model Catalog, 14 Capability Pools with 144 model slots, and Sub-40ms Autonomous Supervisor (`pool/auto`).
- **Module 4 (Queued):** WebUI search grounding and pool selector dropdowns.
- **Module 5 (Queued):** Telegram ChatOps `🧠 Models & Keys` Ephemeral Key Injector.
- **Module 7 (Queued):** Stack Optimizer non-destructive routing proposal staging and zero-downtime hot reload.
- **Module 8 (Queued):** Concentric Quota Rings 0–9 for multi-account key scaling.
