# Module 5: Tri-Tier Chat Bastions & Deterministic Admin ChatOps

## 1. Module Overview & Architectural Role

Module 5 provides three distinct, decoupled mobile chat experiences over Telegram, WhatsApp, and Discord. It establishes a strict security perimeter separating casual conversation from mission-critical node management.

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

## 2. Resource Allocation Across Sizing Stages

| Container | Stage 1: Ultra-Edge<br>(< 3GB RAM) | Stage 2: Standard Edge<br>(3–6GB RAM) | Stage 3: Power Node<br>(6–12GB RAM) | Stage 4: Workstation<br>(12–24GB RAM) | Stage 5: Uncapped<br>(> 24GB RAM) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `ai-gateway-chat-assistant` | 30 MB | 30 MB | 45 MB | 64 MB | Uncapped |
| `ai-gateway-openclaw-agent` | 95 MB | 95 MB | 128 MB | 256 MB | Uncapped |
| `ai-gateway-admin-chatops` | 35 MB | 35 MB | 50 MB | 64 MB | Uncapped |
| **Module 5 Subtotal** | **160 MB** | **160 MB** | **223 MB** | **384 MB** | **Host Managed** |

---

## 3. Inherent Gaps in Base Specification (`docs/report.md`) & Resolutions

| # | Inherent Gap in `report.md` | Architectural Impact | Module 5 Engineering Resolution |
| :-: | :--- | :--- | :--- |
| **G-5.1** | **Missing Code for Experience A & B** | `report.md` provided code only for `bastion-admin`. `bastion-generic` and `bastion-openclaw` had Docker build contexts in compose, but zero source code. | Provide complete Python implementations for `bastion-generic` (aiogram conversational proxy) and lightweight tool-calling agent wrapper for `bastion-openclaw`. |
| **G-5.2** | **Docker Socket Root Escalation Risk** | Mounting `/var/run/docker.sock` into `admin-chatops` could allow an attacker to escape containers if bot tokens leak. | Enforce strict Telegram User ID authorization drop (`@dp.message(~F.from_user.id == ADMIN_USER_ID)`), require PIN verification for dangerous actions, and restrict shell execution to the Web Admin Panel only. |
| **G-5.3** | **Key Exposure in Chat History** | Pasting an API key into Telegram leaves credentials stored in cloud chat servers. | State machine immediately deletes both the prompt message and the user's secret-containing message via Telegram API upon receipt. |

---

## 4. Component Deliverables & Configuration

### A. The Deterministic Admin ChatOps Menu (`bastion-admin/admin_menu.py`)
- Direct out-of-band communication with `/var/run/docker.sock` and Redis `:6379`.
- Consumes **0 LLM tokens**; latency < 15ms.
- Full navigation tree:
  - `1.0 Quota & Usage:` Live RAM/CPU from `/proc/meminfo`, model RPM counters from Redis.
  - `2.0 Modular Services:` Start/stop toggles for SearXNG, Video Conductor, Tailscale, VPN.
  - `3.0 Model Routing & Key Vault:` Injects new API keys into `.env` and triggers LiteLLM `SIGHUP`.
  - `4.0 Logs & Diagnostics:` Live 40-line sanitized tails of container stdout/stderr.
  - `5.0 Optimizer Proposals:` Inspects staged tuning proposals from Redis and hot-applies diffs.
  - `6.0 Emergency Ops (PIN Protected):` Safe Mode trigger (kills WebUI, Conductor, Grafana, SearXNG to clamp cluster to <800MB RAM).

### B. Security Boundary: Chat Menu vs. Web Admin Panel

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

---

## 5. Verification & Acceptance Runbook (`scripts/test_module5.sh`)

### Automated & Manual Verification Steps:
1. **Mock Telegram Ping:** Sends synthetic `/menu` command via aiogram mock client and asserts root keyboard is returned.
2. **Telemetry Reading:** Fetches telemetry callback (`nav_quota`) and verifies host memory and Redis RPM metrics are rendered in HTML format.
3. **Container Control Test:** Toggles SearXNG container status through Docker socket API, confirming state change from `running` to `stopped` and back.
4. **Key Injection & Redaction Test:** Simulates API key message submission, asserts message is deleted within 500ms, `.env` file is updated, and LiteLLM receives SIGHUP reload signal.
5. **Emergency Safe Mode Test:** Sends security PIN (`ADMIN_SECURITY_PIN`), engages Safe Mode, and verifies non-essential containers (`webui`, `conductor`, `searxng`) are stopped.

### Pass/Fail Criteria:
- **PASS:** Unauthorized users dropped; menu renders instantly with 0 LLM tokens; keys purged; Safe Mode successfully reclaims RAM.
- **FAIL:** Unauthorized messages accepted; Docker socket permission error; unredacted keys left in chat history.
