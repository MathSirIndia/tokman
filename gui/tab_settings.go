package gui

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"tokman/backend/config"
)

func (a *AppUI) buildSettingsTab() fyne.CanvasObject {
	// =========================================================================
	// 1. Local Network & HTTP Binding (Stitch Screen 4)
	// =========================================================================
	portEntry := widget.NewEntry()
	portEntry.SetText(fmt.Sprintf("%d", a.gwServer.Config.Port))
	portEntry.TextStyle = fyne.TextStyle{Monospace: true}

	hostBindingSelect := widget.NewSelect([]string{"127.0.0.1 (Localhost only)", "0.0.0.0 (All network interfaces)"}, nil)
	hostBindingSelect.SetSelected("127.0.0.1 (Localhost only)")

	keyEntry := widget.NewEntry()
	keyEntry.SetText(a.gwServer.Config.MasterKey)
	keyEntry.TextStyle = fyne.TextStyle{Monospace: true}
	keyEntry.Password = true

	revealBtn := widget.NewButtonWithIcon("", theme.VisibilityIcon(), nil)
	revealBtn.Importance = widget.LowImportance
	revealBtn.OnTapped = func() {
		keyEntry.Password = !keyEntry.Password
		if keyEntry.Password {
			revealBtn.SetIcon(theme.VisibilityIcon())
		} else {
			revealBtn.SetIcon(theme.VisibilityOffIcon())
		}
		keyEntry.Refresh()
	}

	netCol1 := container.NewVBox(
		fieldLabel("HTTP Listen Port"),
		verticalSpacer(4),
		portEntry,
	)
	netCol2 := container.NewVBox(
		fieldLabel("Host Interface Binding"),
		verticalSpacer(4),
		hostBindingSelect,
	)
	netCol3 := container.NewVBox(
		fieldLabel("Master Internal Key"),
		verticalSpacer(4),
		container.NewBorder(nil, nil, nil, revealBtn, keyEntry),
	)
	netGrid := container.NewGridWithColumns(3, netCol1, netCol2, netCol3)

	netCard := createStitchCard(
		"Local Network & HTTP Binding",
		theme.SettingsIcon(),
		widget.NewLabelWithStyle("OpenAI Protocol Ingress", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true}),
		netGrid,
	)

	// =========================================================================
	// 2. Upstream Provider Key Toggles (Stitch Screen 4)
	// =========================================================================
	groqKey := os.Getenv("GROQ_API_KEY")
	groqStatus := "Not Configured"
	groqConfigured := false
	if groqKey != "" {
		if len(groqKey) > 10 {
			groqStatus = fmt.Sprintf("Configured (%s...%s)", groqKey[:4], groqKey[len(groqKey)-4:])
		} else {
			groqStatus = "Configured"
		}
		groqConfigured = true
	}

	createProviderToggle := func(name, status string, configured bool) fyne.CanvasObject {
		nameLbl := widget.NewLabelWithStyle(name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		badge := createPillBadge(status, configured)
		chk := widget.NewCheck("", nil)
		chk.SetChecked(configured)
		return container.NewBorder(nil, nil, nameLbl, chk, container.NewCenter(badge))
	}

	providersBox := container.NewVBox(
		createProviderToggle("Groq Cloud LPU", groqStatus, groqConfigured),
		verticalSpacer(4),
		widget.NewSeparator(),
		verticalSpacer(4),
		createProviderToggle("Cerebras Cloud (Ultra-Fast)", "Not Configured", false),
		verticalSpacer(4),
		widget.NewSeparator(),
		verticalSpacer(4),
		createProviderToggle("Google Gemini AI Studio", "Not Configured", false),
	)

	providerCard := createStitchCard(
		"Upstream Provider Key Toggles",
		theme.ComputerIcon(),
		widget.NewLabelWithStyle("Multi-Provider Fleet (Tier 10)", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true}),
		providersBox,
	)

	// =========================================================================
	// 3. 2-Column Row: Cache Engine & System Health (Stitch Screen 4)
	// =========================================================================
	lruSlider := widget.NewSlider(100, 5000)
	lruSlider.SetValue(2000)
	lruLabel := widget.NewLabelWithStyle("2,000", fyne.TextAlignTrailing, fyne.TextStyle{Monospace: true})
	lruSlider.OnChanged = func(v float64) {
		lruLabel.SetText(fmt.Sprintf("%d", int(v)))
	}

	walCheck := widget.NewCheck("SQLite WAL Persistence (High Concurrency)", nil)
	walCheck.SetChecked(true)

	cacheBox := container.NewVBox(
		container.NewBorder(nil, nil, fieldLabel("LRU Capacity"), lruLabel, lruSlider),
		verticalSpacer(8),
		container.NewBorder(nil, nil, fieldLabel("Default TTL"), widget.NewLabelWithStyle("3600 seconds", fyne.TextAlignTrailing, fyne.TextStyle{Monospace: true})),
		verticalSpacer(8),
		walCheck,
	)

	cacheEngineCard := createStitchCard(
		"Cache & Performance Engine",
		theme.StorageIcon(),
		nil,
		cacheBox,
	)

	safeModeCheck := widget.NewCheck("Force Safe Mode", nil)
	safeModeCheck.SetChecked(false)

	healthTile1 := createKPITile("Process vs Host", widget.NewLabelWithStyle("42.4 MB / 16.0 GB", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true, Bold: true}), "Resident memory")
	healthTile2 := createKPITile("Persistence Disk", widget.NewLabelWithStyle("1.4 MB / tokman.db", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true, Bold: true}), "SQLite WAL ledger")
	healthTilesGrid := container.NewGridWithColumns(2, healthTile1, healthTile2)

	circuitBreakerRow := container.NewBorder(
		nil, nil,
		container.NewVBox(
			widget.NewLabelWithStyle("Emergency Circuit Breaker", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Force Safe Mode (<50MB RAM clamp)", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		),
		safeModeCheck,
	)

	systemHealthCard := createStitchCard(
		"System Health & Circuit Breaker",
		theme.WarningIcon(),
		createPillBadge("Online", true),
		container.NewVBox(
			healthTilesGrid,
			verticalSpacer(8),
			circuitBreakerRow,
		),
	)

	bottomGrid := container.NewGridWithColumns(2, cacheEngineCard, systemHealthCard)

	// =========================================================================
	// 4. Save & Reload Actions (Stitch Screen 4)
	// =========================================================================
	saveBtn := widget.NewButtonWithIcon("Save & Apply Config", theme.DocumentSaveIcon(), func() {
		a.gwServer.Config.MasterKey = keyEntry.Text
		dialog.ShowInformation("Configuration Applied", "Gateway settings updated in runtime memory.", a.window)
	})
	saveBtn.Importance = widget.HighImportance

	reloadBtn := widget.NewButtonWithIcon("Reload .env File", theme.ViewRefreshIcon(), func() {
		config.LoadDotEnv(".env")
		dialog.ShowInformation(".env Reloaded", "Environment variables reloaded from .env file.", a.window)
	})
	reloadBtn.Importance = widget.LowImportance

	actionsBar := container.NewBorder(
		nil, nil,
		nil,
		container.NewHBox(reloadBtn, horizontalSpacer(12), saveBtn),
	)

	bodyContent := container.NewVBox(
		netCard,
		verticalSpacer(14),
		providerCard,
		verticalSpacer(14),
		bottomGrid,
	)

	return container.NewBorder(
		verticalSpacer(12),
		container.NewVBox(verticalSpacer(10), actionsBar, verticalSpacer(8)),
		horizontalSpacer(18),
		horizontalSpacer(18),
		container.NewVScroll(bodyContent),
	)
}
