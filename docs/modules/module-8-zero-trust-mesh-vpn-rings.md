# Module 8: Zero-Trust Perimeter, Mesh Ingress & Concentric Quota Rings

## 1. Module Overview & Architectural Role

Module 8 enforces the cluster's network security perimeter and horizontal scaling engine. It establishes:

1. **Absolute Zero Inbound WAN Exposure:** All router/firewall inbound ports remain closed. The gateway binds strictly to `127.0.0.1` and local private subnets.
2. **Encrypted Ingress Mesh (Tailscale Overlay):** Remote devices connect securely over peer-to-peer WireGuard overlays (`tailscale0`) without port forwarding or public DNS.
3. **Concentric Quota Rings (Horizontal Free-Tier Scaling):** Distributes traffic across account reservoirs to multiply daily token limits from 5.5M up to 55M tokens/day:
   - **Ring 0:** Primary Direct Residential / Office ISP IP.
   - **Ring 1–9:** Outbound Gluetun WireGuard VPN sidecars (Private Internet Access / custom WireGuard configurations) providing isolated egress IP shards.
4. **Automated Upstream 429 IP Cycling:** Automatically rotates egress WireGuard tunnel when an upstream provider issues an HTTP 429.

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

---

## 2. Resource Allocation Across Sizing Stages

| Container | Stage 1: Ultra-Edge<br>(< 3GB RAM) | Stage 2: Standard Edge<br>(3–6GB RAM) | Stage 3: Power Node<br>(6–12GB RAM) | Stage 4: Workstation<br>(12–24GB RAM) | Stage 5: Uncapped<br>(> 24GB RAM) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `ai-gateway-tailscale` | 35 MB | 45 MB | 64 MB | 128 MB | Uncapped |
| `ai-gateway-vpn-ring1` | 30 MB / tunnel | 35 MB / tunnel | 45 MB / tunnel | 64 MB / tunnel | Uncapped |
| **Module 8 Subtotal** | **65 MB** | **80 MB** | **109 MB** | **192 MB** | **Host Managed** |

---

## 3. Inherent Gaps in Base Specification (`docs/report.md`) & Resolutions

| # | Inherent Gap in `report.md` | Architectural Impact | Module 8 Engineering Resolution |
| :-: | :--- | :--- | :--- |
| **G-8.1** | **Datacenter/VPN IP Blacklisting by Google & Mistral** | Routing all providers through Gluetun VPN rings causes Google AI Studio and Mistral to return HTTP 403 or trigger SMS re-verification. | Implement **Asymmetric Provider Egress Routing**: Ring 0 (Direct ISP) is strictly pinned for Tier B anti-sybil providers (Google, Mistral), while VPN rings (Rings 1–9) are utilized for Tier A friction-free providers (Pollinations, Hugging Face, OpenRouter, Cerebras). |
| **G-8.2** | **Docker Kernel Privileges (`/dev/net/tun`)** | Tailscale and Gluetun containers fail to boot if `/dev/net/tun` is missing or if user runs rootless Docker. | Document host prerequisites (`sudo modprobe tun`) and configure fallback userspace networking mode (`TS_USERSPACE=true`). |
| **G-8.3** | **VPN Health Stalls** | A stalled WireGuard handshake causes requests to hang indefinitely. | Add an active health check to Gluetun probing public IP resolution every 15 seconds; LiteLLM router automatically bypasses rings failing health checks. |

---

## 4. Component Deliverables & Configuration

### A. Horizontal Sharding Capacity Matrix

| Account Sets Configured | Active Routing Shards | Safe Daily Token Budget | Concurrent Interactive RPM |
| :--- | :--- | :--- | :--- |
| **1 Set (Base Mesh)** | Ring 0 (Direct Host IP) | ~5.5M Tokens/Day | 30 RPM sustained (75 RPM burst) |
| **3 Sets (Small Cluster)** | Ring 0 + Rings 1–2 (Gluetun) | ~16.5M Tokens/Day | 90 RPM sustained (200 RPM burst) |
| **5 Sets (Team Shard)** | Ring 0 + Rings 1–4 (Gluetun) | ~27.5M Tokens/Day | 150 RPM sustained (350 RPM burst) |
| **10 Sets (Maximum Node)** | Ring 0 + Rings 1–9 (Gluetun) | ~55.0M Tokens/Day | 300 RPM sustained (700 RPM burst) |

### B. Modular Compose Profiles (`compose.yml`)
```yaml
  tailscale:
    image: tailscale/tailscale:latest
    container_name: ai-gateway-tailscale
    profiles: ["mesh-ingress", "full"]
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
    profiles: ["vpn-egress", "full"]
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

## 5. Verification & Acceptance Runbook (`scripts/test_module8.sh`)

### Automated Verification Steps:
1. **Firewall WAN Probe:** Verifies no inbound listening ports are exposed to public interfaces (`0.0.0.0`), confirming strict binding to `127.0.0.1`.
2. **Tailscale Connectivity Test:** Executes `tailscale status` inside the container and confirms mesh node registration.
3. **Egress IP Diversity Test:**
   - Probes `curl -s https://ipinfo.io/ip` via Host Ring 0.
   - Probes `curl -s --proxy http://127.0.0.1:1080 https://ipinfo.io/ip` via Gluetun Ring 1.
   - Asserts the two public IPs are distinct and from different ASNs.
4. **429 Auto-Rotation Simulation:** Triggers simulated 429 and asserts Gluetun cycles connection to alternate server region.

### Pass/Fail Criteria:
- **PASS:** Inbound ports sealed; Tailscale operational; Ring 1 egress IP is verified distinct from Ring 0; no IP leaks.
- **FAIL:** Open WAN port detected; WireGuard tunnel fails to establish; IP leak on VPN ring.
