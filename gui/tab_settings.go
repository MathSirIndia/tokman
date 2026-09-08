package gui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
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
	portEntry.Disable() // Socket already bound at startup

	hostBindingSelect := widget.NewSelect([]string{"127.0.0.1 (Localhost only)", "0.0.0.0 (All network interfaces)"}, nil)
	hostBindingSelect.SetSelected("127.0.0.1 (Localhost only)")
	hostBindingSelect.Disable() // Bound at startup

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

	genKeyBtn := widget.NewButtonWithIcon("Gen Key", theme.ContentAddIcon(), func() {
		b := make([]byte, 16)
		rand.Read(b)
		newKey := fmt.Sprintf("sk-tokman-%s", hex.EncodeToString(b))
		keyEntry.SetText(newKey)
		keyEntry.Password = false
		revealBtn.SetIcon(theme.VisibilityOffIcon())
		keyEntry.Refresh()
		if a.endpointsMasterKeyEntry != nil {
			a.endpointsMasterKeyEntry.SetText(newKey)
		}
	})
	genKeyBtn.Importance = widget.LowImportance

	netCol1 := container.NewVBox(
		fieldLabel("HTTP Listen Port"),
		verticalSpacer(4),
		portEntry,
		mutedText("Bound at startup (pass --port to rebind)", 9, false),
	)
	netCol2 := container.NewVBox(
		fieldLabel("Host Interface Binding"),
		verticalSpacer(4),
		hostBindingSelect,
		mutedText("Loopback interface (immutable at runtime)", 9, false),
	)
	netCol3 := container.NewVBox(
		fieldLabel("Master Internal Key"),
		verticalSpacer(4),
		container.NewBorder(nil, nil, nil, container.NewHBox(revealBtn, genKeyBtn), keyEntry),
		mutedText("Bearer token for /v1 routes (live reload)", 9, false),
	)
	netGrid := container.NewGridWithColumns(3, netCol1, netCol2, netCol3)

	netCard := createStitchCard(
		"Local Network & HTTP Binding",
		theme.SettingsIcon(),
		widget.NewLabelWithStyle("OpenAI Protocol Ingress", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true}),
		container.NewVBox(
			netGrid,
			verticalSpacer(6),
			mutedText("ℹ Note: TCP port/host listeners are bound when the process launches. Master Key changes apply immediately upon saving.", 10, true),
		),
	)

	// =========================================================================
	// 2. Upstream Provider Key Management (Interactive Credential Ingestion)
	// =========================================================================
	providersBox := container.NewVBox()

	var refreshProviders func()
	refreshProviders = func() {
		providersBox.Objects = nil
		for idx, p := range allProvidersList {
			spec := p
			isKeyless := spec.EnvVar == "POLLINATIONS_PUBLIC"

			var configured bool
			var status string
			if isKeyless {
				configured = true
				status = "Zero-Auth (Public)"
			} else {
				val := os.Getenv(spec.EnvVar)
				configured = config.IsConfiguredKey(val)
				status = "Not Configured"
				if configured {
					if len(val) > 10 {
						status = fmt.Sprintf("Configured (%s...%s)", val[:4], val[len(val)-4:])
					} else {
						status = "Configured"
					}
				} else if val != "" {
					status = "Placeholder (.env)"
				}
			}

			nameLbl := widget.NewLabelWithStyle(spec.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			envLbl := mutedText(fmt.Sprintf("[%s]", spec.EnvVar), 9, true)
			badge := createPillBadge(status, configured)

			var actionBtn fyne.CanvasObject
			if isKeyless {
				openWebBtn := widget.NewButtonWithIcon("", websiteIcon, func() {
					if u, err := url.Parse(spec.ConsoleURL); err == nil {
						a.app.OpenURL(u)
					}
				})
				openWebBtn.Importance = widget.LowImportance
				actionBtn = container.NewHBox(openWebBtn)
			} else {
				setKeyBtn := widget.NewButtonWithIcon("", keyIcon, func() {
					a.showProviderKeyDialog(spec, refreshProviders)
				})
				setKeyBtn.Importance = widget.LowImportance

				// Website open icon comes AFTER the Set Key icon button
				openWebBtn := widget.NewButtonWithIcon("", websiteIcon, func() {
					if u, err := url.Parse(spec.ConsoleURL); err == nil {
						a.app.OpenURL(u)
					}
					a.showProviderKeyDialog(spec, refreshProviders)
				})
				openWebBtn.Importance = widget.LowImportance

				actionBtn = container.NewHBox(setKeyBtn, horizontalSpacer(4), openWebBtn)
			}

			row := container.NewBorder(
				nil, nil,
				container.NewVBox(nameLbl, envLbl),
				container.NewHBox(badge, horizontalSpacer(8), actionBtn),
			)

			providersBox.Add(row)
			if idx < len(allProvidersList)-1 {
				providersBox.Add(verticalSpacer(4))
				providersBox.Add(widget.NewSeparator())
				providersBox.Add(verticalSpacer(4))
			}
		}
		providersBox.Refresh()
	}

	refreshProviders()

	gotoMatrixBtn := widget.NewButtonWithIcon("Full Provider Matrix ➔", theme.NavigateNextIcon(), func() {
		a.tabs.SelectIndex(1)
	})
	gotoMatrixBtn.Importance = widget.LowImportance

	providerCard := createStitchCard(
		"Upstream Provider Key Registry (10 Providers)",
		theme.ComputerIcon(),
		gotoMatrixBtn,
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

	if a.healthProcessRamLabel == nil {
		a.healthProcessRamLabel = widget.NewLabelWithStyle("Calculating...", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true, Bold: true})
	}
	healthTile1 := createKPITile("RAM", a.healthProcessRamLabel, "Usage / Total Host")
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
		if a.endpointsMasterKeyEntry != nil {
			a.endpointsMasterKeyEntry.SetText(keyEntry.Text)
		}
		_ = config.UpdateDotEnv(".env", "LITELLM_MASTER_KEY", keyEntry.Text)
		_ = config.UpdateDotEnv(".env", "TOKMAN_MASTER_KEY", keyEntry.Text)
		dialog.ShowInformation("Configuration Applied", "Gateway Master Key updated in runtime memory and saved to .env.", a.window)
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
