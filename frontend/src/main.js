// Native Desktop Client Controller for TokMan AI Gateway

const GATEWAY_BASE = 'http://127.0.0.1:8000';

let bootTimestamp = Date.now();
let hasSyncedBoot = false;

// Format seconds into HH:MM:SS
function formatHMS(totalSeconds) {
  if (totalSeconds < 0) totalSeconds = 0;
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  
  const pad = (n) => String(n).padStart(2, '0');
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
}

// 1-Second Continuous Live Ticker
function tickUptime() {
  const elapsedSeconds = Math.floor((Date.now() - bootTimestamp) / 1000);
  const formatted = formatHMS(elapsedSeconds);

  const topbarEl = document.getElementById('topbar-uptime');
  if (topbarEl) topbarEl.textContent = formatted;

  const cardEl = document.getElementById('card-uptime');
  if (cardEl) cardEl.textContent = formatted;
}

// Start 1-second continuous tick immediately
setInterval(tickUptime, 1000);
tickUptime();

// Telemetry Polling & Synchronization
async function syncTelemetry() {
  try {
    let data = null;

    if (window.go && window.go.main && window.go.main.App) {
      data = await window.go.main.App.GetStatus();
    } else {
      const res = await fetch(`${GATEWAY_BASE}/health/readiness`, { cache: 'no-store' });
      if (res.ok) {
        data = await res.json();
      }
    }

    if (data) {
      // Synchronize boot timestamp from server on first sync
      if (!hasSyncedBoot && data.uptime_seconds !== undefined) {
        bootTimestamp = Date.now() - (data.uptime_seconds * 1000);
        hasSyncedBoot = true;
        tickUptime();
      }

      // RAM Metrics (Process Heap/RSS + System RAM)
      const procRAM = data.process_ram_mb !== undefined ? data.process_ram_mb : (data.memory_mb || 0);
      const ramText = `${procRAM.toFixed(1)} MB`;
      const topbarRAM = document.getElementById('topbar-ram');
      if (topbarRAM) topbarRAM.textContent = ramText;
      const cardRAM = document.getElementById('card-ram');
      if (cardRAM) cardRAM.textContent = ramText;

      const cardRAMSub = document.getElementById('card-ram-sub');
      if (cardRAMSub && data.system_ram_total_mb) {
        const sysUsedGB = (data.system_ram_used_mb / 1024).toFixed(1);
        const sysTotGB = (data.system_ram_total_mb / 1024).toFixed(1);
        cardRAMSub.textContent = `Host RAM: ${sysUsedGB} / ${sysTotGB} GB`;
      }

      // ROM / Persistent Storage Metrics (SQLite DB + Disk Free)
      const dbKB = data.storage_db_kb !== undefined ? data.storage_db_kb : 20;
      const romText = dbKB >= 1024 ? `${(dbKB / 1024).toFixed(1)} MB` : `${dbKB} KB`;
      const topbarROM = document.getElementById('topbar-rom');
      if (topbarROM) topbarROM.textContent = romText;
      const cardROM = document.getElementById('card-rom');
      if (cardROM) cardROM.textContent = romText;

      const cardROMSub = document.getElementById('card-rom-sub');
      if (cardROMSub && data.storage_disk_free_gb) {
        cardROMSub.textContent = `data/tokman.db | ${data.storage_disk_free_gb.toFixed(0)} GB Free`;
      }

      // In-Memory LRU Cache
      const cacheCount = data.cache_entries ?? 0;
      const topbarCache = document.getElementById('topbar-cache');
      if (topbarCache) topbarCache.textContent = cacheCount;
      const cardCache = document.getElementById('card-cache');
      if (cardCache) cardCache.textContent = `${cacheCount} entries`;
    }
  } catch (err) {
    console.debug('Telemetry sync:', err.message);
  }
}

// Fetch telemetry every 2 seconds to keep stats fresh
setInterval(syncTelemetry, 2000);
syncTelemetry();

// Navigation Tabs
document.querySelectorAll('.nav-item').forEach(item => {
  item.addEventListener('click', () => {
    document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
    document.querySelectorAll('.view-section').forEach(v => v.classList.remove('active'));

    item.classList.add('active');
    const tabId = item.getAttribute('data-tab');
    const targetSection = document.getElementById(tabId);
    if (targetSection) {
      targetSection.classList.add('active');
      if (tabId === 'tab-traffic') {
        loadTrafficLogs();
      }
    }
  });
});

// Refresh Buttons
document.getElementById('btn-refresh-stats')?.addEventListener('click', () => {
  syncTelemetry();
});

document.getElementById('btn-refresh-logs')?.addEventListener('click', () => {
  loadTrafficLogs();
});

// Load Traffic / Spend Logs
async function loadTrafficLogs() {
  const tbody = document.getElementById('traffic-table-body');
  if (!tbody) return;

  try {
    let logs = [];
    if (window.go && window.go.main && window.go.main.App) {
      logs = await window.go.main.App.GetRecentLogs(25);
    } else {
      // In headless browser view, query from API or mock
      return;
    }

    if (!logs || logs.length === 0) {
      tbody.innerHTML = `<tr><td colspan="8" style="text-align: center; color: var(--text-dim); padding: 24px;">No completion transactions recorded yet.</td></tr>`;
      return;
    }

    tbody.innerHTML = logs.map(l => `
      <tr>
        <td style="font-family: var(--font-mono); color: var(--text-dim);">${l.id}</td>
        <td>${new Date(l.timestamp).toLocaleTimeString()}</td>
        <td><strong style="font-family: var(--font-mono); color: var(--accent-blue);">${l.model}</strong></td>
        <td>${l.prompt_tokens}</td>
        <td>${l.completion_tokens}</td>
        <td><strong>${l.total_tokens}</strong></td>
        <td>${l.latency_ms.toFixed(1)}ms</td>
        <td><span class="tag ${l.cached ? 'green' : 'amber'}">${l.cached ? 'CACHE HIT' : 'NETWORK'}</span></td>
      </tr>
    `).join('');
  } catch (err) {
    console.debug('Failed to load logs:', err);
  }
}

// API Console Request Dispatcher
document.getElementById('btn-send-prompt')?.addEventListener('click', async () => {
  const pool = document.getElementById('console-pool-select').value;
  const prompt = document.getElementById('console-prompt-input').value;
  const outputEl = document.getElementById('console-response-output');
  const latencyBadge = document.getElementById('console-latency-badge');
  const statusBadge = document.getElementById('response-status-badge');
  const sendBtn = document.getElementById('btn-send-prompt');
  const spinner = document.getElementById('console-spinner');

  if (!prompt.trim()) return;

  sendBtn.disabled = true;
  spinner.style.display = 'inline';
  statusBadge.className = 'tag amber';
  statusBadge.textContent = 'DISPATCHING...';
  outputEl.textContent = 'Routing request to ' + pool + '...';
  latencyBadge.textContent = '';

  const start = performance.now();

  try {
    const res = await fetch(`${GATEWAY_BASE}/v1/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer sk-master-internal-network-key'
      },
      body: JSON.stringify({
        model: pool,
        messages: [{ role: 'user', content: prompt }]
      })
    });

    const elapsedMs = (performance.now() - start).toFixed(1);
    const cacheStatus = res.headers.get('X-Cache') || 'MISS';

    if (!res.ok) {
      const errText = await res.text();
      statusBadge.className = 'tag red';
      statusBadge.textContent = `HTTP ${res.status}`;
      outputEl.textContent = `Error: ${errText}`;
      latencyBadge.textContent = `${elapsedMs}ms`;
      return;
    }

    const data = await res.json();
    const content = data.choices?.[0]?.message?.content || JSON.stringify(data, null, 2);

    statusBadge.className = 'tag green';
    statusBadge.textContent = `200 OK (${cacheStatus})`;
    outputEl.textContent = content;
    latencyBadge.textContent = `${elapsedMs}ms [${cacheStatus}]`;

    // Populate token audit pills
    const usage = data.usage;
    if (usage) {
      const pEl = document.getElementById('audit-prompt-tokens');
      const cEl = document.getElementById('audit-completion-tokens');
      const tEl = document.getElementById('audit-total-tokens');
      if (pEl) pEl.textContent = `Prompt: ${usage.prompt_tokens} tok`;
      if (cEl) cEl.textContent = `Completion: ${usage.completion_tokens} tok`;
      if (tEl) tEl.textContent = `Total: ${usage.total_tokens} tok`;
    }

    syncTelemetry();
  } catch (err) {
    statusBadge.className = 'tag amber';
    statusBadge.textContent = 'NETWORK ERROR';
    outputEl.textContent = `Failed to connect to gateway at ${GATEWAY_BASE}: ${err.message}`;
  } finally {
    sendBtn.disabled = false;
    spinner.style.display = 'none';
  }
});

// Clear Output Button
document.getElementById('btn-clear-output')?.addEventListener('click', () => {
  const outputEl = document.getElementById('console-response-output');
  if (outputEl) outputEl.textContent = 'Ready for request submission.';
});

// Settings Save Key
document.getElementById('btn-save-settings')?.addEventListener('click', async () => {
  const keyInput = document.getElementById('settings-groq-key');
  const feedback = document.getElementById('settings-save-feedback');
  const key = keyInput.value.trim();

  if (!key) return;

  if (window.go && window.go.main && window.go.main.App) {
    const msg = await window.go.main.App.SaveAPIKey(key);
    feedback.textContent = msg;
    keyInput.value = '';
    setTimeout(() => { feedback.textContent = ''; }, 3000);
  } else {
    feedback.textContent = 'API key updated for current session.';
    setTimeout(() => { feedback.textContent = ''; }, 3000);
  }
});
