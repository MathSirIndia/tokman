# Module 7: Stack Optimizer (`pool/stack-optimizer`) & Observability Control Plane

## 1. Module Overview & Architectural Role

Module 7 implements the intelligence governance and observability layer of the cluster. It pairs the **Zero-Touch Stack Optimizer** with **Grafana Telemetry** and the **Web Admin Panel (`:3000`)**.

Key capabilities:
1. **Human-in-the-Loop Stack Optimizer:** Evaluates 7-day rolling token burn patterns to detect expiring free-tier quota (e.g. unused SambaNova/Google capacity) and stages non-destructive routing tuning proposals in Redis with 48-hour TTL.
2. **Zero-Touch Safety Boundary:** No routing tables or model priorities are modified automatically unless the operator clicks `[Approve]` or explicitly enables `AUTO_APPROVE_TUNING=true`.
3. **Grafana Real-Time Dashboard (`:3005`):** Visualizes cluster metrics (RPM per provider, token burn rates, latency percentiles, error codes, and memory utilization).
4. **Web Admin Panel (`:3000`):** Secure LAN-only interface for high blast-radius and cryptographic management tasks that are restricted from mobile chat interfaces.

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

---

## 2. Resource Allocation Across Sizing Stages

| Container | Stage 1: Ultra-Edge<br>(< 3GB RAM) | Stage 2: Standard Edge<br>(3–6GB RAM) | Stage 3: Power Node<br>(6–12GB RAM) | Stage 4: Workstation<br>(12–24GB RAM) | Stage 5: Uncapped<br>(> 24GB RAM) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `ai-gateway-admin` | 95 MB | 128 MB | 160 MB | 256 MB | Uncapped |
| `ai-gateway-grafana` | 110 MB | 128 MB | 160 MB | 256 MB | Uncapped |
| **Module 7 Subtotal** | **205 MB** | **256 MB** | **320 MB** | **512 MB** | **Host Managed** |

---

## 3. Inherent Gaps in Base Specification (`docs/report.md`) & Resolutions

| # | Inherent Gap in `report.md` | Architectural Impact | Module 7 Engineering Resolution |
| :-: | :--- | :--- | :--- |
| **G-7.1** | **Missing Optimizer Calculation Logic** | `report.md` detailed the concept of proposal generation, but lacked the script calculating token waste from PostgreSQL logs. | Implement `daemon/optimizer.py` which aggregates `LiteLLM_SpendLogs` over a 7-day rolling window, identifies underutilized quotas, and publishes proposals to Redis. |
| **G-7.2** | **Missing Web Admin Panel Source** | `compose.yml` had a build context for `./dashboard`, but no frontend or server code existed in the repository. | Provide a lightweight, zero-dependency FastAPI/HTML5 dashboard on `:3000` with direct Docker API access. |
| **G-7.3** | **Grafana Provisioning File Drift** | Grafana fails to launch if `datasources.yaml` or `dashboards.yaml` fail to connect to PostgreSQL. | Create standardized declarative provisioning templates with health check retries. |

---

## 4. Component Deliverables & Configuration

### A. Staged Review Card Format (Admin Panel & ChatOps Menu)
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

### B. Hot-Reload Architecture
When approved:
1. Interceptor updates `config/config.yaml` with the staged diff.
2. Interceptor emits a non-destructive `SIGHUP` signal to `ai-gateway-litellm`.
3. LiteLLM hot-reloads model weights and routing tables without dropping active SSE streams or restarting containers.

---

## 5. Verification & Acceptance Runbook (`scripts/test_module7.sh`)

### Automated Verification Steps:
1. **Telemetry Ingestion Test:** Queries PostgreSQL spend log view and verifies spend data format.
2. **Optimizer Proposal Generation:** Runs `daemon/optimizer.py --dry-run` and asserts a valid proposal JSON structure is generated and written to Redis key `proposal:opt-test`.
3. **Approval Hot-Reload Test:** Simulates approval click, verifies YAML diff is applied to `config/config.yaml`, and LiteLLM hot-reloads without downtime.
4. **Grafana Health Check:** Probes `http://127.0.0.1:3005/api/health` and verifies PostgreSQL datasource is connected.
5. **Web Admin Interface Probe:** Validates HTTP 200 on `http://127.0.0.1:3000`.

### Pass/Fail Criteria:
- **PASS:** Proposals generated from real telemetry; approval hot-reloads LiteLLM with 0 downtime; Grafana connects to Postgres.
- **FAIL:** Unapproved changes automatically applied; hot-reload causes connection drop; dashboard fails to boot.
