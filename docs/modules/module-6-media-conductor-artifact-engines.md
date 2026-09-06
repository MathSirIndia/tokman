# Module 6: Deterministic Video Conductor & Multimedia Artifact Engines

## 1. Module Overview & Architectural Role

Module 6 provides deterministic, GPU-less media synthesis on consumer edge hardware. It bypasses fragile, resource-heavy text-to-video diffusion models in favor of **Deterministic Motion Synthesis** and headless document/slide generation.

Key capabilities:
1. **Video Conductor Pipeline (`pool/video-conductor`):** Converts prompts into 1080p narrative videos with Flux.1 keyframes, synchronized Kokoro-82M audio, Whisper timing, and programmatic FFmpeg Ken Burns pan/zoom transforms.
2. **Single-Worker FIFO Queue (`task_queue:conductor`):** Prevents CPU/RAM thrashing by strictly serializing heavy render jobs with `MAX_CONCURRENT_WORKERS=1`.
3. **Headless Presentation Engine (`pool/presentation`):** Generates modern Marp / Reveal.js slide decks in HTML and vector PDF.
4. **Formal Document Engine (`pool/document-gen`):** Compiles publication-grade PDFs using the ultra-fast Typst Rust compiler.

```text
                              [ INCOMING VIDEO REQUEST ]
                                         │
                                         ▼
                 ┌────────────────────────────────────────────────┐
                 │ STAGE 1: THE VIDEO CONDUCTOR (pool/architect)  │
                 │ • Evaluates narrative, scene count & pacing    │
                 │ • Compiles `video_manifest.json` (Seed/Colors) │
                 │ • Clamps: Default <=30s / --long <=60s         │
                 └───────────────────────┬────────────────────────┘
                                         │
                 ┌───────────────────────┴────────────────────────┐
                 ▼                                                ▼
┌───────────────────────────────────┐    ┌──────────────────────────────────┐
│ STAGE 2A: VISUAL ANCHORS          │    │ STAGE 2B: AUDIO TRACK SYNTHESIS  │
│ (`pool/image-gen`)                │    │ (`pool/audio-gen`)               │
├───────────────────────────────────┤    ├──────────────────────────────────┤
│ • 1080p base keyframes via Flux.1 │    │ • Narration via Kokoro-82M TTS   │
│ • Fixed visual seed and palette   │    │ • Subtitle alignments via Whisper│
└─────────────────┬─────────────────┘    └────────────────┬─────────────────┘
                  │                                       │
                  └───────────────────┬───────────────────┘
                                      ▼
                 ┌────────────────────────────────────────────────┐
                 │ STAGE 3: DETERMINISTIC MOTION COMPOSITING      │
                 │ • Single-Worker Redis FIFO Queue Execution     │
                 │ • Mathematical Ken Burns Camera Transforms     │
                 │ • Burns subtitles, crossfades, outputs MP4     │
                 └────────────────────────────────────────────────┘
```

---

## 2. Resource Allocation Across Sizing Stages

| Container | Stage 1: Ultra-Edge<br>(< 3GB RAM) | Stage 2: Standard Edge<br>(3–6GB RAM) | Stage 3: Power Node<br>(6–12GB RAM) | Stage 4: Workstation<br>(12–24GB RAM) | Stage 5: Uncapped<br>(> 24GB RAM) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `ai-gateway-video-conductor` | 150 MB (Idle) / 512 MB (Burst) | 200 MB / 768 MB | 250 MB / 1024 MB | 512 MB / 2048 MB | Uncapped |

---

## 3. Inherent Gaps in Base Specification (`docs/report.md`) & Resolutions

| # | Inherent Gap in `report.md` | Architectural Impact | Module 6 Engineering Resolution |
| :-: | :--- | :--- | :--- |
| **G-6.1** | **Unrealistic 150MB FFmpeg Cgroup Cap** | FFmpeg encoding 1080p video with `zoompan` filters and audio muxing will spike to 300–450MB RAM, causing the container to be killed by the Linux kernel. | Enforce cloud-offloaded synthesis for images/audio (`pool/image-gen`, `pool/audio-gen`) and configure Docker `mem_reservation: 150M` with a higher hard burst cap (`mem_limit: 512M`). |
| **G-6.2** | **Missing Typst & Marp Implementations** | `report.md` listed Typst and Marp in tables, but provided no worker or rendering service in the codebase. | Add lightweight headless document generation scripts (`conductor/render_doc.py`) bundling the Typst binary and Marp CLI inside the conductor container. |
| **G-6.3** | **Job Starvation & Timeouts** | If a job crashes mid-way, Redis FIFO queue could leave workers stuck in a dead-lock. | Implement a 120-second heartbeat supervisor on active jobs; if uncompleted, purge scratch files and emit status `{"status": "failed", "reason": "timeout"}`. |

---

## 4. Component Deliverables & Configuration

### A. Video Conductor Worker (`conductor/worker.py`)
- Subscribes to Redis `task_queue:conductor`.
- Clamps duration: `<= 30s` (3–4 scenes) by default, `<= 60s` (6–8 scenes) with `--long`.
- FFmpeg Ken Burns transformation command:
  ```bash
  ffmpeg -y -loop 1 -i scene_0.png -vf "zoompan=z='min(zoom+0.0015,1.2)':d=144:x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':s=1920x1080:fps=24" -t 6.0 -pix_fmt yuv420p -c:v libx264 -preset veryfast clip_0.mp4
  ```
- Subtitle & audio muxing with millisecond synchronization.

### B. Conductor Dockerfile (`conductor/Dockerfile`)
```dockerfile
FROM python:3.11-slim-bookworm

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY worker.py .

CMD ["python", "worker.py"]
```

---

## 5. Verification & Acceptance Runbook (`scripts/test_module6.sh`)

### Automated Verification Steps:
1. **Queue Enqueue Test:** Pushes synthetic job payload to Redis `task_queue:conductor` and verifies immediate queue position response.
2. **Manifest Generation Test:** Probes `pool/architect` with prompt `"A history of the solar system"`, verifies strict JSON manifest returned.
3. **Ken Burns FFmpeg Transform Test:** Renders a test 5-second 1080p MP4 clip from a static PNG image and verifies resolution is exactly 1920x1080 at 24fps.
4. **Muxing & Concat Verification:** Muxes test audio with video clip and verifies final output MP4 in `data/artifacts/`.
5. **Memory Monitoring:** Validates that memory consumption during transcoding remains within the allocated cgroup limit.

### Pass/Fail Criteria:
- **PASS:** Final video renders under duration cap; 1080p video motion is smooth; audio narration is synchronized; memory stays within cgroup bounds.
- **FAIL:** Worker OOM crash, video duration > 60s, or audio desynchronization > 1s.
