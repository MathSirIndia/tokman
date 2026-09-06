package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"tokman/backend/gateway"
	"tokman/backend/storage"
)

type AppUI struct {
	app      fyne.App
	window   fyne.Window
	gwServer *gateway.Server
	db       *storage.DB
	cache    *storage.LRUCache

	startTime time.Time

	tabs                *container.AppTabs
	consolePoolSelect   *widget.Select
	consolePromptEntry  *widget.Entry
	consoleTargetNote   *widget.Label
	consoleRawJsonEntry *widget.Entry

	cacheHitRatioLabel  *widget.Label
	cacheEntriesLabel   *widget.Label
	cacheSavedLabel     *widget.Label
	ledgerListContainer *fyne.Container
	inspectorLabel      *widget.Label

	statusBarLabel *widget.Label
}

func loadDotEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, "\"'")
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

// Run initializes and executes the Fyne native desktop GUI
func Run(gwServer *gateway.Server, db *storage.DB, cache *storage.LRUCache, initialTab int) {
	fyneApp := app.NewWithID("com.tokman.mesh")
	fyneApp.Settings().SetTheme(&TokmanTheme{})

	window := fyneApp.NewWindow("TokMan AI Gateway Mesh — Local Node")
	window.Resize(fyne.NewSize(1340, 980))
	window.CenterOnScreen()

	ui := &AppUI{
		app:       fyneApp,
		window:    window,
		gwServer:  gwServer,
		db:        db,
		cache:     cache,
		startTime: time.Now(),
	}

	tab1 := container.NewTabItemWithIcon("Endpoints Hub", theme.HomeIcon(), ui.buildEndpointsTab())
	tab2 := container.NewTabItemWithIcon("Quick Test Console", theme.MediaPlayIcon(), ui.buildConsoleTab())
	tab3 := container.NewTabItemWithIcon("Cache Ledger", theme.StorageIcon(), ui.buildCacheTab())
	tab4 := container.NewTabItemWithIcon("Daemon Settings", theme.SettingsIcon(), ui.buildSettingsTab())

	ui.tabs = container.NewAppTabs(tab1, tab2, tab3, tab4)
	if initialTab > 0 && initialTab < len(ui.tabs.Items) {
		ui.tabs.SelectIndex(initialTab)
	}

	// Status Bar
	ui.statusBarLabel = widget.NewLabelWithStyle(
		"Initializing TokMan Telemetry...",
		fyne.TextAlignLeading,
		fyne.TextStyle{Monospace: true},
	)

	statusBarContent := container.NewBorder(
		nil, nil,
		horizontalSpacer(16),
		container.NewHBox(
			widget.NewLabelWithStyle("● Mode: Local Standalone", fyne.TextAlignTrailing, fyne.TextStyle{Monospace: true, Bold: true}),
			horizontalSpacer(16),
		),
		ui.statusBarLabel,
	)

	statusBar := container.NewVBox(
		widget.NewSeparator(),
		container.NewBorder(
			verticalSpacer(4),
			verticalSpacer(6),
			nil,
			nil,
			statusBarContent,
		),
	)

	mainLayout := container.NewBorder(
		nil,
		statusBar,
		nil,
		nil,
		ui.tabs,
	)

	window.SetContent(mainLayout)

	// Background ticker for 1-second status updates
	go ui.startStatusTicker()

	window.ShowAndRun()
}

func (a *AppUI) updateUI(fn func()) {
	fyne.Do(fn)
}

func (a *AppUI) startStatusTicker() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Memory usage
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		procRAMMB := float64(m.Alloc) / 1024 / 1024

		hostRAMTotalGB := 16.0
		if meminfo, err := os.ReadFile("/proc/meminfo"); err == nil {
			lines := strings.Split(string(meminfo), "\n")
			for _, l := range lines {
				if strings.HasPrefix(l, "MemTotal:") {
					var kb uint64
					fmt.Sscanf(l, "MemTotal: %d kB", &kb)
					if kb > 0 {
						hostRAMTotalGB = float64(kb) / 1024 / 1024
					}
					break
				}
			}
		}

		// ROM / Persistent Storage usage (TokMan binary + embedded SQLite + data directory)
		var appStorageBytes int64
		if execPath, err := os.Executable(); err == nil {
			if info, err := os.Stat(execPath); err == nil {
				appStorageBytes += info.Size()
			}
		}
		_ = filepath.Walk("data", func(_ string, info os.FileInfo, err error) error {
			if err == nil && info != nil && !info.IsDir() {
				appStorageBytes += info.Size()
			}
			return nil
		})
		appStorageMB := float64(appStorageBytes) / 1024 / 1024
		if appStorageMB < 0.1 {
			if info, err := os.Stat("data/tokman.db"); err == nil {
				appStorageMB = float64(info.Size()) / 1024 / 1024
			}
		}

		diskFreeGB := getDiskFreeGB()

		// Uptime
		elapsed := time.Since(a.startTime)
		hours := int(elapsed.Hours())
		mins := int(elapsed.Minutes()) % 60
		secs := int(elapsed.Seconds()) % 60
		uptimeStr := fmt.Sprintf("%02dh %02dm %02ds", hours, mins, secs)

		// Active Pools & Models metrics (Module 1 Runtime State vs Roadmap Architecture)
		totalPools := len(capabilityPools)
		activePools := 0
		for _, p := range capabilityPools {
			if p.Status == "Active" {
				activePools++
			}
		}
		totalModelSlots := 144 // Master 144-Model Capability Matrix target (docs/report.md Section 5)
		activeModels := activePools // 1 active flagship model running on Groq in Module 1

		statusText := fmt.Sprintf(
			"Host RAM: %.1f MB / %.1f GB   |   ROM: %.1f MB / %.0f GB   |   Pools: %d/%d Active   |   Models: %d/%d Active   |   Uptime: %s",
			procRAMMB, hostRAMTotalGB, appStorageMB, diskFreeGB, activePools, totalPools, activeModels, totalModelSlots, uptimeStr,
		)

		if a.statusBarLabel != nil {
			fyne.Do(func() {
				a.statusBarLabel.SetText(statusText)
			})
		}
	}
}
