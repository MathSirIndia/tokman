# Module 4: Local Search Grounding (SearXNG) & Universal WebUI Interface

## 1. Module Overview & Architectural Role

Module 4 provides private internet search grounding and the primary interactive web interface for the cluster. It pairs a zero-tracking meta-search engine (**SearXNG**) with **Open WebUI**, augmented with an in-browser client-side chat migration utility (**Universal Selective Chat Importer**).

This module delivers:
1. **Private Meta-Search Grounding:** Queries aggregated across DuckDuckGo, Wikipedia, Brave, and terminal fallbacks without tracking or outbound leaks.
2. **Anti-Ban Residential Safeguards:** Redis 1-hour search caching and scraping concurrency clamped to 1 thread per engine.
3. **Open WebUI Interactive Gateway:** Configured on `:3001` with web search hooked directly into SearXNG and completion requests routed through the Interceptor `:8000`.
4. **Universal Client-Side Selective Importer:** Intercepts file uploads in the browser to parse, filter, and import conversation histories from 20+ external AI platforms without sending unredacted raw files to the server.

```text
                                [ CLIENT BROWSER ]
                                        │
                                        ▼
┌───────────────────────────────────────────────────────────────────────────────────────┐
│                           OPEN WEBUI INTERFACE (:3001)                                │
│  • Client-Side Selective Importer Injected (`selective_import.js`)                    │
│  • Connects to FastAPI Interceptor (`http://interceptor:8000/v1`)                     │
│  • Built-in Web Search routed to SearXNG                                              │
└───────────────────────────────────────┬───────────────────────────────────────────────┘
                                        │ (Search Queries)
                                        ▼
┌───────────────────────────────────────────────────────────────────────────────────────┐
│                         SEARXNG META-SEARCH ENGINE (:8080)                            │
│  • Anti-Ban Throttling: 1 Scraping Thread per Engine                                  │
│  • Residential IP Protection: DuckDuckGo (2.0) > Wikipedia (1.5) > Bing > Google (0.5)│
│  • Redis Query Caching: 1-hour TTL                                                    │
└───────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Resource Allocation Across Sizing Stages

| Container | Stage 1: Ultra-Edge<br>(< 3GB RAM) | Stage 2: Standard Edge<br>(3–6GB RAM) | Stage 3: Power Node<br>(6–12GB RAM) | Stage 4: Workstation<br>(12–24GB RAM) | Stage 5: Uncapped<br>(> 24GB RAM) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `ai-gateway-searxng` | 60 MB | 80 MB | 120 MB | 256 MB | Uncapped |
| `ai-gateway-webui` | 450 MB | 512 MB | 768 MB | 1024 MB | Uncapped |
| **Module 4 Subtotal** | **510 MB** | **592 MB** | **888 MB** | **1280 MB** | **Host Managed** |

---

## 3. Inherent Gaps in Base Specification (`docs/report.md`) & Resolutions

| # | Inherent Gap in `report.md` | Architectural Impact | Module 4 Engineering Resolution |
| :-: | :--- | :--- | :--- |
| **G-4.1** | **Open WebUI Local Embedding OOM** | Open WebUI defaults to loading local Hugging Face `sentence-transformers` for document RAG, requiring > 1GB RAM and crashing edge nodes. | Disable local models via `RAG_EMBEDDING_ENGINE=openai` pointing to `http://interceptor:8000/v1` (`pool/general`), freeing 600MB host RAM. |
| **G-4.2** | **SearXNG Secret Key Warning** | Hardcoded secret key in `settings.yml` causes startup warnings and session invalidation. | Parameterize `SEARXNG_SECRET_KEY` via `.env` and inject dynamically during container initialization. |
| **G-4.3** | **Client-Side JS Injection Race Condition** | Open WebUI builds occasionally overwrite `index.html` on restart, wiping the injected `<script src="/selective_import.js">`. | Use container entrypoint wrapper to idempotently re-insert the script tag before running `start.sh`. |

---

## 4. Component Deliverables & Configuration

### A. SearXNG Configuration (`config/searxng/settings.yml`)
```yaml
use_default_settings: true
general:
  debug: false
  instance_name: "AI Gateway Meta-Search"
server:
  secret_key: "${SEARXNG_SECRET_KEY:-ai_gateway_internal_search_secret}"
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

### B. Universal Selective Chat Importer (`config/webui/selective_import.js`)
- Injected into Open WebUI DOM.
- Heuristic parser identifies ChatGPT, Claude, Gemini, DeepSeek, Grok, Kimi, Poe, and generic Markdown archives.
- Renders Catppuccin-themed modal dialog allowing thread-by-thread search, selection, and JSON normalization.

---

## 5. Verification & Acceptance Runbook (`scripts/test_module4.sh`)

### Automated Verification Steps:
1. **SearXNG JSON API Probe**:
   `curl -s "http://127.0.0.1:8080/search?q=open+source&format=json"`
   Asserts HTTP 200 and >= 1 result returned without IP ban.
2. **Open WebUI Health Probe**:
   `curl -f http://127.0.0.1:3001/api/version`
   Asserts Open WebUI frontend and backend are alive.
3. **Selective Importer Injection Check**:
   `curl -s http://127.0.0.1:3001/ | grep -q "selective_import.js"`
   Asserts script tag is correctly inserted before `</body>`.
4. **End-to-End Grounded Chat Query**:
   Sends a query requiring current real-time data through Open WebUI with web search toggled ON. Asserts SearXNG citations are included in output.

### Pass/Fail Criteria:
- **PASS:** SearXNG returns structured JSON results; Open WebUI loads; script injected; grounded responses contain web citations.
- **FAIL:** SearXNG returns HTTP 403/429; script tag missing in HTML; WebUI fails to connect to `:8000`.
