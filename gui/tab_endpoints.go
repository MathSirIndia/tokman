package gui

import (
	"fmt"
	"image/color"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"tokman/backend/registry"
	"tokman/backend/types"
)

type ServicePortal struct {
	Name        string
	Icon        fyne.Resource
	URL         string
	Port        string
	Status      string
	IsActive    bool
	Protocol    string
	Description string
	ActionLabel string
	ActionType  string // "open_browser", "open_console"
}

// getMeshPortals constructs the platform service endpoints dynamically after Fyne app has initialized
func getMeshPortals() []ServicePortal {
	return []ServicePortal{
		{
			Name:        "Go Interceptor & Traffic Shaper",
			Icon:        theme.HomeIcon(),
			URL:         "http://localhost:8000/v1",
			Port:        ":8000",
			Status:      "Active (Port 8000)",
			IsActive:    true,
			Protocol:    "OpenAI Protocol Ingress",
			Description: "Fair-share priority queues, token-bucket limiter & dynamic <think> sanitizer",
			ActionLabel: "Quick Test",
			ActionType:  "open_console",
		},
		{
			Name:        "Web Admin Mesh Control Plane",
			Icon:        theme.ComputerIcon(),
			URL:         "http://localhost:8000/admin",
			Port:        ":8000",
			Status:      "Active (Ready)",
			IsActive:    true,
			Protocol:    "HTTP Web Portal",
			Description: "Distributed WireGuard edge fleet, P2P gossip stream & live topology",
			ActionLabel: "Open Dashboard ↗",
			ActionType:  "open_browser",
		},
		{
			Name:        "Open WebUI Conversational Portal",
			Icon:        theme.MailComposeIcon(),
			URL:         "http://localhost:3001",
			Port:        ":3001",
			Status:      "Staged (Module 4)",
			IsActive:    false,
			Protocol:    "Web Application",
			Description: "Multi-user chat interface with SearXNG grounding & model interaction",
			ActionLabel: "Launch UI ↗",
			ActionType:  "open_browser",
		},
		{
			Name:        "Tri-Tier ChatOps Bastion",
			Icon:        theme.StorageIcon(),
			URL:         "http://localhost:8000/bastion",
			Port:        ":3002",
			Status:      "Staged (Module 5)",
			IsActive:    false,
			Protocol:    "Telegram / Discord / WA",
			Description: "Generic assistant, OpenClaw agent & deterministic 0-token admin menu",
			ActionLabel: "View Spec ↗",
			ActionType:  "open_browser",
		},
		{
			Name:        "Video Conductor FIFO Pipeline",
			Icon:        theme.MediaPlayIcon(),
			URL:         "http://localhost:3004",
			Port:        ":3004",
			Status:      "Staged (Module 6)",
			IsActive:    false,
			Protocol:    "Deterministic Video Engine",
			Description: "Flux keyframes + Kokoro-82M TTS + Whisper + FFmpeg Ken Burns transforms",
			ActionLabel: "View Spec ↗",
			ActionType:  "open_browser",
		},
		{
			Name:        "Grafana Observability Dashboard",
			Icon:        theme.ViewRefreshIcon(),
			URL:         "http://localhost:3005",
			Port:        ":3005",
			Status:      "Staged (Module 7)",
			IsActive:    false,
			Protocol:    "Prometheus / Grafana OSS",
			Description: "Real-time cluster telemetry, P50/P90/P99 latency & token burn analytics",
			ActionLabel: "Open Grafana ↗",
			ActionType:  "open_browser",
		},
		{
			Name:        "SearXNG Meta-Search Grounding",
			Icon:        theme.SearchIcon(),
			URL:         "http://localhost:8080",
			Port:        ":8080",
			Status:      "Staged (Module 4)",
			IsActive:    false,
			Protocol:    "Privacy Search Engine",
			Description: "Multi-engine web retrieval & anti-ban dispenser for pool/web-research",
			ActionLabel: "Search API ↗",
			ActionType:  "open_browser",
		},
		{
			Name:        "Go Federation Core Router",
			Icon:        theme.MenuIcon(),
			URL:         "http://localhost:8000/v1/models",
			Port:        ":8000",
			Status:      "Active (In-Process)",
			IsActive:    true,
			Protocol:    "14-Pool Federation Engine",
			Description: "144-slot capability matrix, concentric quota rings & sub-40ms supervisor",
			ActionLabel: "View Models ↗",
			ActionType:  "open_browser",
		},
	}
}

// PoolInfo is aliased to types.PoolInfo
type PoolInfo = types.PoolInfo

// capabilityPools derives from the centralized CanonicalPools registry (BL-1)
var capabilityPools = registry.CanonicalPools

type EndpointItem struct {
	Method      string
	Path        string
	Description string
	IsConsole   bool
}

var gatewayEndpoints = []EndpointItem{
	{
		Method:      "POST",
		Path:        "/v1/chat/completions",
		Description: "OpenAI-compatible inference with LRU caching, telemetry & streaming",
		IsConsole:   true,
	},
	{
		Method:      "GET",
		Path:        "/v1/models",
		Description: "List active capability pools (14 pools) and mapped model targets",
		IsConsole:   false,
	},
	{
		Method:      "GET",
		Path:        "/health/readiness",
		Description: "Deep readiness check: upstream provider keys, SQLite DB, LRU cache",
		IsConsole:   false,
	},
	{
		Method:      "GET",
		Path:        "/health/liveness",
		Description: "Daemon process heartbeat, memory diagnostics & uptime status",
		IsConsole:   false,
	},
	{
		Method:      "GET",
		Path:        "/admin",
		Description: "Web Admin Control Plane (Stitch screen 280ad36a6c9f4585807e6ed2cdc002c0)",
		IsConsole:   false,
	},
}

func (a *AppUI) buildEndpointsTab() fyne.CanvasObject {
	gatewayURL := fmt.Sprintf("http://localhost:%d", a.gwServer.Config.Port)

	// =========================================================================
	// 1. Mesh Service Endpoints (Platform Access Portals - docs/report.md)
	// =========================================================================
	portalsGrid := container.NewGridWithColumns(2)
	for _, portal := range getMeshPortals() {
		p := portal
		if p.Name == "Go Interceptor & Traffic Shaper" || p.Name == "FastAPI Interceptor & Traffic Shaper" {
			p.URL = fmt.Sprintf("http://localhost:%d/v1", a.gwServer.Config.Port)
		} else if p.Name == "Web Admin Mesh Control Plane" {
			p.URL = fmt.Sprintf("http://localhost:%d/admin", a.gwServer.Config.Port)
		} else if p.Name == "Tri-Tier ChatOps Bastion" {
			p.URL = fmt.Sprintf("http://localhost:%d/bastion", a.gwServer.Config.Port)
		} else if p.Name == "Go Federation Core Router" || p.Name == "LiteLLM Core Routing Engine" {
			p.URL = fmt.Sprintf("http://localhost:%d/v1/models", a.gwServer.Config.Port)
		}
		portalsGrid.Add(a.createServicePortalCard(p))
	}

	portalsCard := createStitchCard(
		"Mesh Service Endpoints (Platform Access Portals — docs/report.md)",
		theme.ComputerIcon(),
		mutedText("8 Cluster Service Entrypoints", 11, false),
		portalsGrid,
	)

	// =========================================================================
	// 2. Gateway Credentials & Connection
	// =========================================================================
	urlEntry := widget.NewEntry()
	urlEntry.SetText(gatewayURL)
	urlEntry.TextStyle = fyne.TextStyle{Monospace: true}
	urlEntry.Disable()

	copyURLBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		a.window.Clipboard().SetContent(gatewayURL)
	})
	copyURLBtn.Importance = widget.LowImportance

	a.endpointsMasterKeyEntry = widget.NewEntry()
	a.endpointsMasterKeyEntry.SetText(a.gwServer.Config.MasterKey)
	a.endpointsMasterKeyEntry.TextStyle = fyne.TextStyle{Monospace: true}
	a.endpointsMasterKeyEntry.Password = true
	a.endpointsMasterKeyEntry.Disable()

	revealKeyBtn := widget.NewButtonWithIcon("", theme.VisibilityIcon(), nil)
	revealKeyBtn.Importance = widget.LowImportance
	revealKeyBtn.OnTapped = func() {
		a.endpointsMasterKeyEntry.Password = !a.endpointsMasterKeyEntry.Password
		if a.endpointsMasterKeyEntry.Password {
			revealKeyBtn.SetIcon(theme.VisibilityIcon())
		} else {
			revealKeyBtn.SetIcon(theme.VisibilityOffIcon())
		}
		a.endpointsMasterKeyEntry.Refresh()
	}

	copyKeyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		a.window.Clipboard().SetContent(a.gwServer.Config.MasterKey)
	})
	copyKeyBtn.Importance = widget.LowImportance

	leftCol := container.NewVBox(
		mutedText("BASE INGRESS URL", 10, true),
		verticalSpacer(4),
		container.NewBorder(nil, nil, nil, copyURLBtn, urlEntry),
	)

	rightCol := container.NewVBox(
		mutedText("MESH MASTER INTERNAL KEY (CLIENT INGRESS)", 10, true),
		verticalSpacer(4),
		container.NewBorder(nil, nil, nil, container.NewHBox(revealKeyBtn, copyKeyBtn), a.endpointsMasterKeyEntry),
		mutedText("Bearer token for SDKs & curl (to re-generate or edit, open Daemon Settings)", 9, false),
	)

	ingressGrid := container.NewGridWithColumns(2, leftCol, rightCol)
	statusBadge := createPillBadge(fmt.Sprintf("Running (Port %d)", a.gwServer.Config.Port), true)

	gatewayCard := createStitchCard(
		"Local Gateway Ingress Credentials",
		theme.HomeIcon(),
		statusBadge,
		ingressGrid,
	)

	// =========================================================================
	// 3. Gateway REST Endpoints Directory
	// =========================================================================
	endpointsBox := container.NewVBox(
		a.createEndpointRow(gatewayEndpoints[0], gatewayURL),
		verticalSpacer(4),
		widget.NewSeparator(),
		verticalSpacer(4),
		a.createEndpointRow(gatewayEndpoints[1], gatewayURL),
		verticalSpacer(4),
		widget.NewSeparator(),
		verticalSpacer(4),
		a.createEndpointRow(gatewayEndpoints[2], gatewayURL),
		verticalSpacer(4),
		widget.NewSeparator(),
		verticalSpacer(4),
		a.createEndpointRow(gatewayEndpoints[3], gatewayURL),
		verticalSpacer(4),
		widget.NewSeparator(),
		verticalSpacer(4),
		a.createEndpointRow(gatewayEndpoints[4], gatewayURL),
	)

	endpointsCard := createStitchCard(
		"Gateway REST Endpoints (HTTP Routes)",
		theme.ListIcon(),
		mutedText("OpenAI Protocol Ingress", 11, false),
		endpointsBox,
	)

	// =========================================================================
	// 4. Step-by-Step Gateway REST Integration Guide (docs/report.md)
	// =========================================================================
	step1 := container.NewVBox(
		widget.NewLabelWithStyle("1. Configure Client Base URL", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		mutedText("Set OpenAI API Base URL to http://localhost:8000/v1 in your client (Open WebUI, Cursor, Aider, Claude Code, or LangChain).", 10, false),
	)

	step2 := container.NewVBox(
		widget.NewLabelWithStyle("2. Supply Master Ingress Key", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		mutedText("Pass your Mesh Master Internal Key (copied above) as the Authorization Bearer token header in all requests.", 10, false),
	)

	step3 := container.NewVBox(
		widget.NewLabelWithStyle("3. Dispatch to Capability Pools", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		mutedText("Specify 'pool/auto' for sub-40ms autonomous intent routing, or target pools directly (e.g. pool/agent-coding, pool/general).", 10, false),
	)

	step4 := container.NewVBox(
		widget.NewLabelWithStyle("4. Monitor Cache & Telemetry", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		mutedText("Switch to the Provider Matrix tab to inspect 4×4×4 fallback arrays, or Cache Ledger tab for sub-1ms cache hits & spend logs.", 10, false),
	)

	guideBox := container.NewVBox(
		step1,
		verticalSpacer(6),
		widget.NewSeparator(),
		verticalSpacer(6),
		step2,
		verticalSpacer(6),
		widget.NewSeparator(),
		verticalSpacer(6),
		step3,
		verticalSpacer(6),
		widget.NewSeparator(),
		verticalSpacer(6),
		step4,
	)

	guideCard := createStitchCard(
		"How to Use Gateway REST Endpoints",
		theme.HelpIcon(),
		mutedText("4-Step Integration Flow", 11, true),
		guideBox,
	)

	middleSplit := container.NewGridWithColumns(2, endpointsCard, guideCard)

	// =========================================================================
	// 5. Quick-Start Example (cURL)
	// =========================================================================
	curlSnippet := fmt.Sprintf(`curl %s/v1/chat/completions \
  -H "Content-Type: application/json" -H "Authorization: Bearer %s" \
  -d '{"model": "pool/general", "messages": [{"role": "user", "content": "Initialize mesh diagnostics."}]}'`, gatewayURL, a.gwServer.Config.MasterKey)

	curlLbl := widget.NewLabelWithStyle(curlSnippet, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	curlBox := createInsetBox(curlLbl)

	testInConsoleBtn := widget.NewButtonWithIcon("Open in Console", theme.MediaPlayIcon(), func() {
		a.openInConsole("pool/general", "Initialize mesh diagnostics.")
	})
	testInConsoleBtn.Importance = widget.LowImportance

	copyCurlBtn := widget.NewButtonWithIcon("Copy Snippet", theme.ContentCopyIcon(), func() {
		a.window.Clipboard().SetContent(curlSnippet)
	})
	copyCurlBtn.Importance = widget.HighImportance

	quickStartHeaderActions := container.NewHBox(
		testInConsoleBtn,
		horizontalSpacer(8),
		copyCurlBtn,
	)

	quickStartCard := createStitchCard(
		"Quick-Start Ingress Example (cURL)",
		theme.MediaPlayIcon(),
		quickStartHeaderActions,
		curlBox,
	)

	content := container.NewVBox(
		portalsCard,
		verticalSpacer(14),
		gatewayCard,
		verticalSpacer(14),
		middleSplit,
		verticalSpacer(14),
		quickStartCard,
	)

	return wrapWithTabMargins(content)
}

func (a *AppUI) createServicePortalCard(portal ServicePortal) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.RGBA{R: 0x14, G: 0x14, B: 0x18, A: 0xff})
	bg.StrokeColor = color.RGBA{R: 0x27, G: 0x27, B: 0x30, A: 0xff}
	bg.StrokeWidth = 1.0
	bg.CornerRadius = 6.0

	titleLabel := primaryText(portal.Name, 11, true, false)
	statusBadge := createPillBadge(portal.Status, portal.IsActive)

	topRow := container.NewBorder(
		nil, nil,
		container.NewHBox(widget.NewIcon(portal.Icon), horizontalSpacer(4), titleLabel),
		statusBadge,
	)

	descLabel := mutedText(portal.Description, 10, false)

	urlEntry := widget.NewEntry()
	urlEntry.SetText(portal.URL)
	urlEntry.TextStyle = fyne.TextStyle{Monospace: true}
	urlEntry.Disable()

	copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		a.window.Clipboard().SetContent(portal.URL)
	})
	copyBtn.Importance = widget.LowImportance

	urlRow := container.NewBorder(nil, nil, nil, copyBtn, urlEntry)

	var actionBtn *widget.Button
	if portal.ActionType == "open_console" {
		actionBtn = widget.NewButtonWithIcon(portal.ActionLabel, theme.MediaPlayIcon(), func() {
			a.openInConsole("pool/general", "Testing AI Gateway Ingress")
		})
		actionBtn.Importance = widget.HighImportance
	} else {
		actionBtn = widget.NewButtonWithIcon(portal.ActionLabel, theme.NavigateNextIcon(), func() {
			if u, err := url.Parse(portal.URL); err == nil {
				a.app.OpenURL(u)
			}
		})
		if portal.IsActive {
			actionBtn.Importance = widget.HighImportance
		} else {
			actionBtn.Importance = widget.LowImportance
		}
	}

	protoTag := container.NewHBox(
		mutedText("PORT: "+portal.Port+" | PROTOCOL:", 9, true),
		primaryText(portal.Protocol, 10, false, true),
	)

	bottomRow := container.NewBorder(
		nil, nil,
		protoTag,
		actionBtn,
	)

	cardContent := container.NewVBox(
		topRow,
		verticalSpacer(4),
		descLabel,
		verticalSpacer(6),
		urlRow,
		verticalSpacer(6),
		bottomRow,
	)

	padded := container.NewBorder(
		verticalSpacer(8),
		verticalSpacer(8),
		horizontalSpacer(12),
		horizontalSpacer(12),
		cardContent,
	)

	return container.NewStack(bg, padded)
}

func (a *AppUI) createEndpointRow(ep EndpointItem, gatewayURL string) fyne.CanvasObject {
	methodBadge := createMethodBadge(ep.Method)
	pathLabel := primaryText(ep.Path, 11, true, true)

	leftBox := container.NewHBox(
		methodBadge,
		horizontalSpacer(6),
		pathLabel,
	)

	descLabel := mutedText(ep.Description, 10, false)

	fullURL := gatewayURL + ep.Path
	copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		a.window.Clipboard().SetContent(fullURL)
	})
	copyBtn.Importance = widget.LowImportance

	var actionRow fyne.CanvasObject
	if ep.IsConsole {
		testBtn := widget.NewButtonWithIcon("Test", theme.MediaPlayIcon(), func() {
			a.openInConsole("pool/general", "Testing /v1/chat/completions endpoint")
		})
		testBtn.Importance = widget.LowImportance
		actionRow = container.NewHBox(copyBtn, testBtn)
	} else {
		actionRow = copyBtn
	}

	return container.NewBorder(
		nil, nil,
		container.NewVBox(leftBox, descLabel),
		actionRow,
	)
}

func (a *AppUI) createCompactPoolCard(pool PoolInfo) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.RGBA{R: 0x14, G: 0x14, B: 0x18, A: 0xff})
	bg.StrokeColor = color.RGBA{R: 0x27, G: 0x27, B: 0x30, A: 0xff}
	bg.StrokeWidth = 1.0
	bg.CornerRadius = 5.0

	nameLabel := primaryText(pool.Name, 11, true, true)
	descLabel := mutedText(pool.Description, 10, false)
	statusBadge := createPillBadge(pool.Status, pool.Status == "Active")

	topRow := container.NewBorder(
		nil, nil,
		nameLabel,
		statusBadge,
		nil,
	)

	metaRow := container.NewHBox(
		mutedText("TARGET:", 9, true),
		primaryText(pool.TargetModel, 10, false, false),
		horizontalSpacer(4),
		mutedText("CTX:", 9, true),
		primaryText(pool.ContextMax, 10, false, true),
	)

	testPoolBtn := widget.NewButtonWithIcon("Test Pool", theme.MediaPlayIcon(), func() {
		a.openInConsole(pool.Name, fmt.Sprintf("Testing canonical pool %s", pool.Name))
	})
	testPoolBtn.Importance = widget.LowImportance

	cardContent := container.NewVBox(
		topRow,
		verticalSpacer(4),
		descLabel,
		verticalSpacer(4),
		metaRow,
		verticalSpacer(6),
		testPoolBtn,
	)

	padded := container.NewBorder(
		verticalSpacer(8),
		verticalSpacer(8),
		horizontalSpacer(10),
		horizontalSpacer(10),
		cardContent,
	)

	return container.NewStack(bg, padded)
}

func (a *AppUI) openInConsole(poolName, prompt string) {
	if a.tabs != nil {
		a.updateConsolePool(poolName)
		a.consolePromptEntry.SetText(prompt)
		a.tabs.SelectIndex(2) // Tab 3: Quick Test Console
	}
}
