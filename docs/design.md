# TokMan AI Gateway Mesh — Design System Specification

## 1. Overview & Creative North Star

**Creative North Star: "The Utilitarian Instrument"**

TokMan is built for high-throughput distributed AI inference, low-latency stream filtering, and strict zero-cost multi-provider quota management. The interface reflects the precision, speed, and reliability of low-level systems infrastructure (inspired by Ghostty, Linear, JetBrains IDEs, and Raycast).

The design system rejects ornamental fluff, neon gradients, and slow animations in favor of **High-Density Technical Minimalism**. Every pixel serves an operational purpose.

---

## 2. Color Palette & Tonal Hierarchy

The palette is engineered for prolonged operational focus in low-light environments, using dark matte charcoal tones with crisp 1px structural boundaries.

```
+-------------------------------------------------------------------------+
| Level 0: Main Application Canvas        #121214 (Zinc-950 Deep Matte)   |
|   +-----------------------------------------------------------------+   |
|   | Level 1: Surface Containers / Cards  #18181b (Zinc-900 Elevation) |   |
|   |   +---------------------------------------------------------+   |   |
|   |   | Level 2: Inset Input / Logs      #0e0e10 (Recessed Black)|   |   |
|   |   +---------------------------------------------------------+   |   |
|   +-----------------------------------------------------------------+   |
+-------------------------------------------------------------------------+
```

### Color Tokens

| Token | Hex | Usage |
| :--- | :--- | :--- |
| `surface-canvas` | `#121214` | Application root window background |
| `surface-panel` | `#18181b` | Cards (`widget.Card`), sidebars, tab content |
| `surface-inset` | `#0e0e10` | Input entries (`widget.Entry`), code viewports, terminal logs |
| `surface-hover` | `#27272a` | List row hover, subtle button hover |
| `border-subtle` | `#27272a` | 1px dividers, card boundaries, table grid lines |
| `border-focus` | `#52525b` | Active input focus ring (no glow) |
| `text-primary` | `#f4f4f5` | Main titles, active values, high-contrast labels |
| `text-secondary`| `#a1a1aa` | Field labels, timestamps, pool descriptions |
| `text-muted` | `#71717a` | Table headers, units, inactive status |
| `accent-primary`| `#ffffff` | Primary action button fill, key interactive highlights |
| `signal-success`| `#10b981` | 200 OK status, healthy tunnel, active key, cache HIT |
| `signal-warning`| `#f59e0b` | High latency (>1000ms), 429 rate limit approaching, warm cache |
| `signal-error` | `#ef4444` | Safe Mode panic, circuit breaker tripped, 5xx outage |
| `signal-cache` | `#06b6d4` | LRU cache HIT badge, instant zero-latency path |

---

## 3. Typography Scale & Dual-Font Hierarchy

TokMan uses a strict dual-font discipline:
1. **Inter** for human interaction (labels, window chrome, button text, dialog descriptions).
2. **JetBrains Mono** for all machine data (URLs, cURL snippets, tokens, latencies, HTTP codes, SHA-256 hashes).

| Token | Font Family | Size | Weight | Line Height | Usage |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `headline-md` | Inter | 16px | 600 (SemiBold) | 22px | Window title, section headers |
| `body-md` | Inter | 13px | 400 (Regular) | 18px | Descriptive text, tooltips |
| `body-sm` | Inter | 11px | 500 (Medium) | 16px | Field labels, tab titles |
| `code-lg` | JetBrains Mono | 14px | 500 (Medium) | 20px | Interactive response text, playground |
| `code-md` | JetBrains Mono | 12px | 400 (Regular) | 18px | cURL snippets, JSON payload viewer |
| `code-sm` | JetBrains Mono | 11px | 400 (Regular) | 16px | Table data, timestamps, memory stats |
| `badge-caps` | Inter | 10px | 700 (Bold) | 12px | Status pills (ACTIVE, STANDBY, HIT) |

---

## 4. Native Desktop GUI Architecture (Fyne v2)

The desktop Control Panel GUI is implemented as a pure-native Go application using the **Fyne v2** framework (`fyne.io/fyne/v2`).

### Window Specifications
- **Default Size:** `1180px × 760px` (Resizable, min `960px × 640px`).
- **Chrome:** Native OS window controls (Minimize, Maximize, Close).
- **Navigation:** Top-aligned horizontal `widget.AppTabs` with icons.
- **Status Bar:** Fixed 24px bottom bar with real-time host RAM, ROM disk footprint, and uptime.

### Canvas View Layouts

```
+---------------------------------------------------------------------------+
| TokMan AI Gateway Mesh - Local Node                                [-][+][x]|
+---------------------------------------------------------------------------+
| [⚡ Endpoints Hub]   [🧪 Test Console]   [💾 Cache Ledger]   [⚙️ Settings]  |
+---------------------------------------------------------------------------+
|                                                                           |
|  [GATEWAY CONNECTION]                                                     |
|  http://localhost:8000 [Copy URL]   sk-master-... [Copy Key]   🟢 Running |
|                                                                           |
|  [QUICK-START]                                                            |
|  curl http://localhost:8000/v1/chat/completions ... [Copy Snippet] [Test] |
|                                                                           |
|  [CAPABILITY POOLS CATALOG]                                               |
|  +---------------------------------------------------------------------+  |
|  | pool/general          openai/gpt-oss-120b (Groq)   [Copy] [Test]     |  |
|  | pool/deep-reasoning   openai/gpt-oss-120b (Groq)   [Copy] [Test]     |  |
|  | pool/agent-coding     codellama-70b (Compound)    [Copy] [Test]     |  |
|  | pool/fast-general     llama3-8b (Cerebras)        [Copy] [Test]     |  |
|  +---------------------------------------------------------------------+  |
|                                                                           |
+---------------------------------------------------------------------------+
| Host RAM: 42.1 MB / 16.0 GB | ROM: 1.2 MB / 480 GB | Uptime: 01h 14m 22s |
+---------------------------------------------------------------------------+
```

### 1. View 1: Endpoints Hub (Primary / Default Landing)
- **Stitch Screen ID:** `adad6359e73142e294f81311b953f99e`
- **Gateway Quick-Start Banner:** Real-time port status, 1-click clipboard copy for Base URL and Master Internal Key.
- **Snippet Box:** Ready-to-execute cURL / Python client code snippet for immediate testing.
- **Capability Pool Rows:** 14 capability pools with target upstream mapping, token caps, operational state, and inline copy/test triggers.

### 2. View 2: Quick Test Console
- **Stitch Screen ID:** `a3ff17eaf9dc4feab1ed5283bdc5f204`
- **Split Canvas (`widget.SplitContainer`):**
  - **Left (Request Composer):** Pool picker dropdown (`widget.Select`), system prompt, user prompt, SSE streaming switch, token and temperature sliders.
  - **Right (Response Inspector):** Monospace stream viewer, status badge (`200 OK`), latency pill, token audit banner (`Prompt + Completion = Total tok/s`), and raw JSON payload foldout.

### 3. View 3: Cache Ledger & Storage Persistence
- **Stitch Screen ID:** `46a1690d003d49f68ba819c290f032c8`
- **Cache Health Banner:** Live Hit Ratio %, in-memory entry count (LRU capacity), cumulative latency saved, and "Flush Cache" danger button.
- **Persistence Table:** Real-time stream of requests logged to SQLite with timestamp, request ID, model pool, latency, and HIT/MISS tags.
- **Entry Inspector:** Granular SHA-256 prompt hash verification and payload debugging.

### 4. View 4: Daemon & Local Settings
- **Stitch Screen ID:** `e4803cf525f34635ae56d0c896c072dc`
- **Network Form:** Port binding (`:8000`), interface binding (`127.0.0.1` vs `0.0.0.0`), Master Key reveal.
- **Provider Toggles:** Visual switches for Groq, Cerebras, SambaNova, Gemini, and Mistral.
- **Performance Settings:** In-memory LRU max items slider, TTL configuration, and SQLite WAL toggle.
- **Safety Switch:** Emergency Circuit Breaker toggle.

---

## 5. Web Admin Mesh Control Plane Architecture

The Web Admin Dashboard is served over HTTP/HTTPS (`:8000/admin`) for remote cluster operations, multi-node mesh topology monitoring, and quota fleet management.

### Canvas View Layouts

```
+---------------------------------------------------------------------------+
| TOKMAN MESH CONTROL PLANE          [32,450 tok/s] [P99: 142ms] [FLEET: 8/8]|
+---------------------------------------------------------------------------+
| [Mesh Topology]   [Quota Rings & Providers]   [Optimizer]   [Safe Mode]   |
+---------------------------------------------------------------------------+
|                                                                           |
|  [DISTRIBUTED EDGE NODES]                                                 |
|  [node-01 (US-East)]   [node-02 (EU-Central)]   [node-03 (AP-South)]      |
|  Ping: 12ms | 4.2k t/s Ping: 18ms | 5.1k t/s    Ping: 24ms | 3.8k t/s     |
|                                                                           |
|  [REAL-TIME P2P GOSSIP STREAM]                                            |
|  15:10:02 [GOSSIP] node-01 -> node-03 heartbeat sync OK (RTT: 21ms)       |
|  15:10:04 [ROUTER] Rebalanced pool/general: 60% Groq / 40% Cerebras       |
|                                                                           |
+---------------------------------------------------------------------------+
| Protocol: Gossip v2.4 | WireGuard: Full-Mesh Enforced | Memory: 42MB/Node |
+---------------------------------------------------------------------------+
```

### 1. View 1: Global Mesh Topology & Edge Fleet
- **Stitch Project:** `796098453406783157` | **Screen ID:** `280ad36a6c9f4585807e6ed2cdc002c0`
- Global fleet throughput aggregation (30k+ tok/s).
- WireGuard node matrix across geographic edge regions with live tunnel ping and load metrics.
- P2P Gossip terminal stream for inter-node state convergence.

### 2. View 2: Concentric Quota Rings & Multi-Provider Cockpit
- **Stitch Project:** `796098453406783157` | **Screen ID:** `195ad3e336084b2698e04475c0a62877`
- Multi-tier radial concentric rings simultaneously tracking RPM, TPM, Daily Free Tier %, and 429 countdown backoffs.
- 10-Provider fleet matrix table (Groq, Cerebras, SambaNova, Mistral, Gemini, Cloudflare AI, Github Models, OpenRouter, DeepSeek, Local Ollama).
- 0ms automated fallback priority pipeline flow diagram.

### 3. View 3: Stack Optimizer & Live Config Diff Engine
- **Stitch Project:** `796098453406783157` | **Screen ID:** `d14497c1d4024c7398baff665be2eb45`
- Split-pane YAML/JSON configuration diff engine showing current live routing vs AI-optimized staged candidate.
- P50, P90, P99 latency percentile curves across capability pools.
- Cost vs. Latency sliding priority weight rule.

### 4. View 4: Security, WireGuard Egress & Safe Mode Panic Console
- **Stitch Project:** `796098453406783157` | **Screen ID:** `fd6603e21e494d32a02d05c731a8cda6`
- Safe Mode Master Kill-Switch: Instant emergency egress isolation and fallback to local offline Ollama models.
- WireGuard encrypted tunnel health, public keys, and 15-minute IP rotation timer.
- Scoped client access token vault.

---

## 6. Component Rules & Interaction Patterns

1. **Zero-Flicker Monospace Updates:**
   - Numerical metrics (tokens, latency, RAM/ROM) must be rendered in fixed-width `JetBrains Mono` so that value shifts do not cause layout reflows.
2. **Instant Clipboard Feedback:**
   - Any copy button (URL, Master Key, cURL, Pool endpoint) must provide instant visual feedback (icon changes to checkmark `✓` for 1,200ms).
3. **Graceful Degraded States:**
   - If an upstream provider returns 429 or 5xx, the UI reflects the circuit-breaker fallback in real time without blocking user interactions.
