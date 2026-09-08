package gui

import (
	"fmt"
	"image/color"
	"net/url"
	"os"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"tokman/backend/config"
	"tokman/backend/registry"
	"tokman/backend/supervisor"
	"tokman/backend/types"
)

// hoverCard implements a responsive, hover-interactive container with pointer cursor, mouse-in luminous highlight, and active border glow
type hoverCard struct {
	widget.BaseWidget
	bgRect       *canvas.Rectangle
	content      fyne.CanvasObject
	onTap        func()
	isSelected   bool
	isHovered    bool
	idleFill     color.Color
	hoverFill    color.Color
	activeFill   color.Color
	idleStroke   color.Color
	hoverStroke  color.Color
	activeStroke color.Color
}

func newHoverCard(
	content fyne.CanvasObject,
	isSelected bool,
	idleFill, hoverFill, activeFill color.Color,
	idleStroke, hoverStroke, activeStroke color.Color,
	onTap func(),
) *hoverCard {
	bg := canvas.NewRectangle(idleFill)
	bg.CornerRadius = 6.0
	bg.StrokeWidth = 1.0
	bg.StrokeColor = idleStroke
	if isSelected {
		bg.FillColor = activeFill
		bg.StrokeColor = activeStroke
		bg.StrokeWidth = 2.0
	}

	h := &hoverCard{
		bgRect:       bg,
		content:      content,
		onTap:        onTap,
		isSelected:   isSelected,
		idleFill:     idleFill,
		hoverFill:    hoverFill,
		activeFill:   activeFill,
		idleStroke:   idleStroke,
		hoverStroke:  hoverStroke,
		activeStroke: activeStroke,
	}
	h.ExtendBaseWidget(h)
	return h
}

func (h *hoverCard) CreateRenderer() fyne.WidgetRenderer {
	stack := container.NewStack(h.bgRect, h.content)
	return widget.NewSimpleRenderer(stack)
}

func (h *hoverCard) Tapped(_ *fyne.PointEvent) {
	if h.onTap != nil {
		h.onTap()
	}
}

func (h *hoverCard) MouseIn(_ *desktop.MouseEvent) {
	h.isHovered = true
	h.updateStyle()
}

func (h *hoverCard) MouseMoved(_ *desktop.MouseEvent) {}

func (h *hoverCard) MouseOut() {
	h.isHovered = false
	h.updateStyle()
}

func (h *hoverCard) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (h *hoverCard) updateStyle() {
	if h.isSelected {
		h.bgRect.FillColor = h.activeFill
		h.bgRect.StrokeColor = h.activeStroke
		h.bgRect.StrokeWidth = 2.0
	} else if h.isHovered {
		h.bgRect.FillColor = h.hoverFill
		h.bgRect.StrokeColor = h.hoverStroke
		h.bgRect.StrokeWidth = 1.3
	} else {
		h.bgRect.FillColor = h.idleFill
		h.bgRect.StrokeColor = h.idleStroke
		h.bgRect.StrokeWidth = 1.0
	}
	h.bgRect.Refresh()
}

func getModelProviderKeyEnv(modelID string) (string, string) {
	lower := strings.ToLower(modelID)
	switch {
	case strings.HasPrefix(lower, "groq/"):
		return "Groq Cloud LPU", "GROQ_API_KEY"
	case strings.HasPrefix(lower, "cerebras/"):
		return "Cerebras Cloud", "CEREBRAS_API_KEY"
	case strings.HasPrefix(lower, "gemini/"):
		return "Google AI Studio", "GEMINI_API_KEY"
	case strings.HasPrefix(lower, "sambanova/"):
		return "SambaNova Cloud", "SAMBANOVA_API_KEY"
	case strings.HasPrefix(lower, "mistral/"):
		return "Mistral AI", "MISTRAL_API_KEY"
	case strings.HasPrefix(lower, "github/"):
		return "GitHub Marketplace", "GITHUB_TOKEN"
	case strings.HasPrefix(lower, "cloudflare/") || strings.HasPrefix(lower, "cf/"):
		return "Cloudflare Workers AI", "CLOUDFLARE_API_KEY"
	case strings.HasPrefix(lower, "openrouter/"):
		return "OpenRouter Multi-Gateway", "OPENROUTER_API_KEY"
	case strings.HasPrefix(lower, "huggingface/"):
		return "HuggingFace Hub", "HUGGINGFACE_API_KEY"
	case strings.HasPrefix(lower, "pollinations/"):
		return "Pollinations.ai", "POLLINATIONS_PUBLIC"
	default:
		return "Local / Direct Ingress", ""
	}
}

func isModelReady(modelID string) (bool, string) {
	lower := strings.ToLower(modelID)
	if strings.HasPrefix(lower, "pollinations/") {
		return true, "Ready (Public Capacity)"
	}
	prov, envVar := getModelProviderKeyEnv(modelID)
	if envVar == "" {
		return false, "Direct Ingress"
	}
	val := os.Getenv(envVar)
	if config.IsConfiguredKey(val) {
		return true, fmt.Sprintf("Ready (%s)", prov)
	}
	return false, fmt.Sprintf("Awaiting %s", envVar)
}

func getPoolReadyCount(poolName string) (ready int, total int) {
	def, hasDef := supervisor.MasterPoolCatalog[poolName]
	if !hasDef {
		return 0, 0
	}
	allModels := append(append(def.PrimaryTier.Models, def.BackupTier1.Models...), def.BackupTier2.Models...)
	total = len(allModels)
	for _, m := range allModels {
		if ok, _ := isModelReady(m); ok {
			ready++
		}
	}
	return ready, total
}

func getDistinctModelStats() (totalDistinct int, activeDistinct int, allDistinct []string) {
	seen := make(map[string]bool)
	activeSeen := make(map[string]bool)
	for _, def := range supervisor.MasterPoolCatalog {
		for _, m := range def.PrimaryTier.Models {
			seen[m] = true
			if ok, _ := isModelReady(m); ok {
				activeSeen[m] = true
			}
		}
		for _, m := range def.BackupTier1.Models {
			seen[m] = true
			if ok, _ := isModelReady(m); ok {
				activeSeen[m] = true
			}
		}
		for _, m := range def.BackupTier2.Models {
			seen[m] = true
			if ok, _ := isModelReady(m); ok {
				activeSeen[m] = true
			}
		}
	}
	for m := range seen {
		allDistinct = append(allDistinct, m)
	}
	sort.Strings(allDistinct)
	return len(seen), len(activeSeen), allDistinct
}

type ProviderInfo struct {
	Name         string
	EnvVar       string
	ConsoleURL   string
	Instructions string
	KeyPrefix    string
}

// websiteIcon represents an authentic globe/web portal icon for external platforms
var websiteIcon = fyne.NewStaticResource("website_icon.svg", []byte(
	`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24">`+
		`<circle cx="12" cy="12" r="10" fill="none" stroke="#38bdf8" stroke-width="2"/>`+
		`<line x1="2" y1="12" x2="22" y2="12" stroke="#38bdf8" stroke-width="2"/>`+
		`<path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" fill="none" stroke="#38bdf8" stroke-width="2"/>`+
		`</svg>`))

// keyIcon represents an authentic key icon for configuring credentials
var keyIcon = fyne.NewStaticResource("key_icon.svg", []byte(
	`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24">`+
		`<circle cx="8" cy="15" r="5" fill="none" stroke="#38bdf8" stroke-width="2"/>`+
		`<path d="M11.5 11.5L21 2" stroke="#38bdf8" stroke-width="2" stroke-linecap="round"/>`+
		`<path d="M17.5 5.5L19.5 7.5" stroke="#38bdf8" stroke-width="2" stroke-linecap="round"/>`+
		`<path d="M15 8L17 10" stroke="#38bdf8" stroke-width="2" stroke-linecap="round"/>`+
		`</svg>`))

// 10 Canonical Upstream Free-Tier Providers from docs/report.md Section 4
var allProvidersList = []ProviderInfo{
	{
		Name:         "Groq Cloud LPU",
		EnvVar:       "GROQ_API_KEY",
		ConsoleURL:   "https://console.groq.com/keys",
		KeyPrefix:    "starts with gsk_...",
		Instructions: "1. Sign in to console.groq.com (Google or GitHub)\n2. Click 'Create API Key' in the API Keys tab\n3. Set a name (e.g. 'tokman') and copy the key (starts with 'gsk_')",
	},
	{
		Name:         "SambaNova Cloud",
		EnvVar:       "SAMBANOVA_API_KEY",
		ConsoleURL:   "https://cloud.sambanova.ai/apis",
		KeyPrefix:    "UUID token format",
		Instructions: "1. Sign up or log in at cloud.sambanova.ai\n2. Navigate to 'API Keys' in the developer sidebar\n3. Click 'Create Key' and copy your authentication secret",
	},
	{
		Name:         "Google Gemini AI Studio",
		EnvVar:       "GEMINI_API_KEY",
		ConsoleURL:   "https://aistudio.google.com/app/apikey",
		KeyPrefix:    "starts with AIzaSy...",
		Instructions: "1. Sign in with your Google account at aistudio.google.com\n2. Click 'Create API key' (or choose an existing project)\n3. Copy the key (free tier includes 15 RPM / 1M TPM for Gemini models)",
	},
	{
		Name:         "Cerebras Cloud (Ultra-Fast)",
		EnvVar:       "CEREBRAS_API_KEY",
		ConsoleURL:   "https://cloud.cerebras.ai/",
		KeyPrefix:    "starts with csk-...",
		Instructions: "1. Sign up or log in at cloud.cerebras.ai\n2. Navigate to 'API Keys' in the left menu\n3. Click 'Create API Key' and copy the token (1,800+ tokens/sec throughput)",
	},
	{
		Name:         "Mistral AI",
		EnvVar:       "MISTRAL_API_KEY",
		ConsoleURL:   "https://console.mistral.ai/api-keys/",
		KeyPrefix:    "Secret token",
		Instructions: "1. Sign in to console.mistral.ai (complete SMS phone verification)\n2. Navigate to 'API Keys'\n3. Click 'Create new key' and copy the secret token",
	},
	{
		Name:         "GitHub Marketplace",
		EnvVar:       "GITHUB_TOKEN",
		ConsoleURL:   "https://github.com/settings/tokens",
		KeyPrefix:    "starts with ghp_... or github_pat_...",
		Instructions: "1. Go to github.com/settings/tokens (Personal Access Tokens)\n2. Generate a token with 'read:packages' or Marketplace Models\n3. Copy the token (starts with 'ghp_' or 'github_pat_')",
	},
	{
		Name:         "Cloudflare Workers AI",
		EnvVar:       "CLOUDFLARE_API_KEY",
		ConsoleURL:   "https://dash.cloudflare.com/",
		KeyPrefix:    "API token",
		Instructions: "1. Log in to dash.cloudflare.com -> AI -> Workers AI\n2. Go to User Profile -> API Tokens -> 'Create Token'\n3. Use 'Workers AI' template with Read/Edit access and copy token",
	},
	{
		Name:         "OpenRouter Multi-Gateway",
		EnvVar:       "OPENROUTER_API_KEY",
		ConsoleURL:   "https://openrouter.ai/keys",
		KeyPrefix:    "starts with sk-or-v1-...",
		Instructions: "1. Sign in at openrouter.ai\n2. Navigate to 'Keys' and click 'Create Key'\n3. Free tier models (:free) require no credits but need a valid key",
	},
	{
		Name:         "HuggingFace Hub",
		EnvVar:       "HUGGINGFACE_API_KEY",
		ConsoleURL:   "https://huggingface.co/settings/tokens",
		KeyPrefix:    "starts with hf_...",
		Instructions: "1. Sign in to huggingface.co\n2. Navigate to Settings -> Access Tokens\n3. Click 'Create new token' with 'Read' role for Serverless Inference",
	},
	{
		Name:         "Pollinations.ai (Public / Keyless)",
		EnvVar:       "POLLINATIONS_PUBLIC",
		ConsoleURL:   "https://pollinations.ai",
		KeyPrefix:    "None (Zero-Auth)",
		Instructions: "Zero-Auth Public Endpoint — No API key, registration, or credit card required! Requests are served anonymously.",
	},
}

func (a *AppUI) showProviderKeyDialog(spec ProviderInfo, onSaved func()) {
	currentVal := os.Getenv(spec.EnvVar)
	entry := widget.NewPasswordEntry()
	entry.SetPlaceHolder(fmt.Sprintf("Paste %s API key...", spec.Name))
	entry.SetText(currentVal)
	entry.TextStyle = fyne.TextStyle{Monospace: true}

	openPortalBtn := widget.NewButtonWithIcon(fmt.Sprintf("Open %s Console ↗", spec.Name), websiteIcon, func() {
		if u, err := url.Parse(spec.ConsoleURL); err == nil {
			a.app.OpenURL(u)
		}
	})
	openPortalBtn.Importance = widget.LowImportance

	instructionsLbl := widget.NewLabel(spec.Instructions)
	instructionsLbl.Wrapping = fyne.TextWrapWord

	instrHeader := widget.NewLabelWithStyle("HOW TO OBTAIN YOUR KEY:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	formatLbl := mutedText(fmt.Sprintf("Expected key format: %s", spec.KeyPrefix), 10, true)

	instrBox := container.NewVBox(
		instrHeader,
		verticalSpacer(2),
		instructionsLbl,
		verticalSpacer(4),
		formatLbl,
	)
	instrPadded := container.NewBorder(verticalSpacer(8), verticalSpacer(8), horizontalSpacer(12), horizontalSpacer(12), instrBox)
	instrBg := canvas.NewRectangle(color.RGBA{R: 0x14, G: 0x18, B: 0x24, A: 0xff})
	instrBg.CornerRadius = 5.0
	instrBg.StrokeColor = color.RGBA{R: 0x27, G: 0x36, B: 0x54, A: 0xff}
	instrBg.StrokeWidth = 1.0
	instrCard := container.NewStack(instrBg, instrPadded)

	titleRow := container.NewBorder(
		nil, nil,
		container.NewVBox(
			widget.NewLabelWithStyle(spec.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			mutedText(fmt.Sprintf("Environment Variable: [%s]", spec.EnvVar), 10, true),
		),
		openPortalBtn,
	)

	box := container.NewVBox(
		titleRow,
		verticalSpacer(8),
		instrCard,
		verticalSpacer(10),
		fieldLabel("Enter or Paste API Key:"),
		verticalSpacer(4),
		entry,
		verticalSpacer(4),
		mutedText("Key will be written to .env and loaded immediately into gateway memory.", 10, true),
	)

	d := dialog.NewCustomConfirm(
		fmt.Sprintf("Configure %s", spec.Name),
		"Save Key",
		"Cancel",
		box,
		func(confirmed bool) {
			if confirmed && strings.TrimSpace(entry.Text) != "" {
				keyVal := strings.TrimSpace(entry.Text)
				config.UpdateDotEnv(".env", spec.EnvVar, keyVal)
				if spec.EnvVar == "GROQ_API_KEY" {
					a.gwServer.SetGroqAPIKey(keyVal)
				}
				dialog.ShowInformation("Key Saved", fmt.Sprintf("API key for %s updated in runtime memory and written to .env.", spec.Name), a.window)
				if onSaved != nil {
					onSaved()
				}
			}
		},
		a.window,
	)
	d.Resize(fyne.NewSize(620, 390))
	d.Show()
}

func (a *AppUI) buildModelsTab() fyne.CanvasObject {
	pools := registry.CanonicalPools
	selectedPoolName := "pool/general"
	currentSubView := 0 // 0: Cluster Pools, 1: Distinct Model Catalog, 2: Upstream Providers

	mainContentContainer := container.NewStack()
	ribbonContainer := container.NewStack()

	var refreshAllViews func()

	// -------------------------------------------------------------------------
	// Subview 1: Cluster Pools & Model Fallback Hierarchy (Default View)
	// -------------------------------------------------------------------------
	buildPoolsView := func() fyne.CanvasObject {
		detailContainer := container.NewVBox()

		renderPoolDetail := func(poolName string) {
			detailContainer.Objects = nil

			var currentPool types.PoolInfo
			for _, p := range pools {
				if p.Name == poolName {
					currentPool = p
					break
				}
			}

			def, hasDef := supervisor.MasterPoolCatalog[poolName]
			if !hasDef {
				detailContainer.Add(widget.NewLabelWithStyle(fmt.Sprintf("Catalog definition for %s is staged for future modules.", poolName), fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
				detailContainer.Refresh()
				return
			}

			readyCount, totalCount := getPoolReadyCount(poolName)

			nameLabel := widget.NewLabelWithStyle(currentPool.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			nameLabel.Truncation = fyne.TextTruncateEllipsis
			tierLabel := mutedText(fmt.Sprintf("%s  •  Context: %s  •  %d/%d Models Ready", currentPool.Tier, currentPool.ContextMax, readyCount, totalCount), 10, true)
			descLabel := widget.NewLabelWithStyle(currentPool.Description, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
			descLabel.Wrapping = fyne.TextWrapWord
			statusBadge := createPillBadge(currentPool.Status, currentPool.Status == "Active")

			testBtn := widget.NewButtonWithIcon("Test in Console", theme.MediaPlayIcon(), func() {
				a.openInConsole(currentPool.Name, fmt.Sprintf("Testing canonical capability pool %s", currentPool.Name))
			})

			// Disable test in console if pool is staged or has 0 ready models (Fix Point 6)
			if currentPool.Status != "Active" {
				testBtn.SetText("Staged Engine")
				testBtn.Disable()
				testBtn.Importance = widget.LowImportance
			} else if readyCount == 0 {
				testBtn.SetText("Awaiting Upstream Key")
				testBtn.Disable()
				testBtn.Importance = widget.LowImportance
			} else {
				testBtn.SetText("Test in Console")
				testBtn.Enable()
				testBtn.Importance = widget.HighImportance
			}

			headerRow := container.NewBorder(
				nil, nil,
				container.NewVBox(nameLabel, tierLabel),
				container.NewHBox(statusBadge, horizontalSpacer(8), testBtn),
			)

			detailContainer.Add(headerRow)
			detailContainer.Add(verticalSpacer(4))
			detailContainer.Add(descLabel)
			detailContainer.Add(verticalSpacer(10))
			detailContainer.Add(widget.NewSeparator())
			detailContainer.Add(verticalSpacer(10))

			renderTierCard := func(title, desc string, tier supervisor.ModelTier) fyne.CanvasObject {
				tierGrid := container.NewGridWithColumns(2)
				for idx, modelID := range tier.Models {
					mID := modelID
					provName, envVar := getModelProviderKeyEnv(mID)
					isReady, statusText := isModelReady(mID)

					mBg := canvas.NewRectangle(color.RGBA{R: 0x14, G: 0x14, B: 0x18, A: 0xff})
					mBg.StrokeColor = color.RGBA{R: 0x27, G: 0x27, B: 0x30, A: 0xff}
					mBg.StrokeWidth = 1.0
					mBg.CornerRadius = 5.0

					mName := widget.NewLabelWithStyle(fmt.Sprintf("%d. %s", idx+1, mID), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true, Bold: true})
					mName.Truncation = fyne.TextTruncateEllipsis
					mProv := mutedText(provName, 9, false)
					mBadge := createPillBadge(statusText, isReady)

					var mEnv fyne.CanvasObject
					if envVar != "" {
						mEnv = mutedText(fmt.Sprintf("[%s]", envVar), 8, true)
					} else {
						mEnv = horizontalSpacer(1)
					}

					mBox := container.NewVBox(
						container.NewBorder(nil, nil, mName, mBadge),
						verticalSpacer(2),
						container.NewBorder(nil, nil, mProv, mEnv),
					)

					padded := container.NewBorder(verticalSpacer(6), verticalSpacer(6), horizontalSpacer(10), horizontalSpacer(10), mBox)
					tierGrid.Add(container.NewStack(mBg, padded))
				}

				tierHeader := container.NewHBox(
					widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					horizontalSpacer(8),
					mutedText(desc, 10, true),
				)

				return container.NewVBox(
					tierHeader,
					verticalSpacer(6),
					tierGrid,
				)
			}

			// Consistent naming: Primary, Secondary Fallback, Tertiary Resiliency (Fix Point 6)
			detailContainer.Add(renderTierCard("Tier 0: Primary Array (4× Models)", "Targeted first; sub-millisecond failover", def.PrimaryTier))
			detailContainer.Add(verticalSpacer(12))
			detailContainer.Add(renderTierCard("Tier 1: Secondary Fallback Array (4× Models)", "Dispatched on Primary 429/5xx exhaustion", def.BackupTier1))
			detailContainer.Add(verticalSpacer(12))
			detailContainer.Add(renderTierCard("Tier 2: Tertiary Fallback Array (4× Models)", "Terminal resiliency ring prior to Bastion fallback", def.BackupTier2))

			detailContainer.Refresh()
		}

		poolsListContainer := container.NewVBox()

		var refreshPoolButtons func()
		refreshPoolButtons = func() {
			poolsListContainer.Objects = nil
			for _, p := range pools {
				poolItem := p
				isSelected := poolItem.Name == selectedPoolName
				readyModels, totalModels := getPoolReadyCount(poolItem.Name)

				nameLabel := widget.NewLabelWithStyle(poolItem.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true})
				nameLabel.Truncation = fyne.TextTruncateEllipsis
				statusBadge := createPillBadge(poolItem.Status, poolItem.Status == "Active")

				modelsStatStr := fmt.Sprintf("%d/%d Models Ready", readyModels, totalModels)
				if totalModels == 0 {
					modelsStatStr = "Staged (Module X)"
				}
				modelsStatLbl := mutedText(modelsStatStr, 9, true)
				tierLbl := mutedText(poolItem.Tier, 9, false)

				rowContent := container.NewBorder(
					nil, nil,
					container.NewVBox(nameLabel, container.NewHBox(tierLbl, horizontalSpacer(6), modelsStatLbl)),
					statusBadge,
				)
				rowPadded := container.NewBorder(verticalSpacer(8), verticalSpacer(8), horizontalSpacer(12), horizontalSpacer(12), rowContent)

				// Responsive, hover-interactive pool switch with hand pointer & illuminated border
				clickablePoolCard := newHoverCard(
					rowPadded,
					isSelected,
					color.RGBA{R: 0x14, G: 0x14, B: 0x18, A: 0xff},
					color.RGBA{R: 0x1c, G: 0x21, B: 0x2e, A: 0xff},
					color.RGBA{R: 0x1a, G: 0x25, B: 0x3c, A: 0xff},
					color.RGBA{R: 0x25, G: 0x26, B: 0x30, A: 0xff},
					color.RGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xff},
					color.RGBA{R: 0x38, G: 0xbd, B: 0xf8, A: 0xff},
					func() {
						selectedPoolName = poolItem.Name
						refreshPoolButtons()
						renderPoolDetail(poolItem.Name)
					},
				)

				poolsListContainer.Add(clickablePoolCard)
				poolsListContainer.Add(verticalSpacer(4))
			}
			poolsListContainer.Refresh()
		}

		refreshPoolButtons()
		renderPoolDetail(selectedPoolName)

		leftScroll := container.NewVScroll(poolsListContainer)
		leftScroll.SetMinSize(fyne.NewSize(360, 500))

		leftCard := createStitchCard(
			"Capability Pools (14)",
			theme.StorageIcon(),
			nil, // Clean: removed patronizing instruction
			leftScroll,
		)

		rightScroll := container.NewVScroll(detailContainer)
		rightScroll.SetMinSize(fyne.NewSize(450, 500))

		rightCard := createStitchCard(
			"Multi-Model Fallback Matrix (4×4×4)",
			theme.ListIcon(),
			mutedText("Primary ➔ Secondary Fallback ➔ Tertiary Resiliency", 10, false),
			rightScroll,
		)

		split := container.NewHSplit(leftCard, rightCard)
		if a.poolSplitOffset <= 0.1 || a.poolSplitOffset >= 0.9 {
			a.poolSplitOffset = 0.32
		}
		split.SetOffset(a.poolSplitOffset)
		return split
	}

	// -------------------------------------------------------------------------
	// Subview 2: Distinct Model Catalog
	// -------------------------------------------------------------------------
	buildCatalogView := func() fyne.CanvasObject {
		_, _, distinctModels := getDistinctModelStats()

		catalogBox := container.NewVBox()

		headerGrid := container.NewGridWithColumns(5,
			widget.NewLabelWithStyle("MODEL IDENTIFIER", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
			widget.NewLabelWithStyle("UPSTREAM PROVIDER", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
			widget.NewLabelWithStyle("ENV KEY VARIABLE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
			widget.NewLabelWithStyle("MAPPED POOLS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
			widget.NewLabelWithStyle("CREDENTIAL STATUS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
		)
		catalogBox.Add(headerGrid)
		catalogBox.Add(widget.NewSeparator())
		catalogBox.Add(verticalSpacer(4))

		for idx, mID := range distinctModels {
			modelName := mID
			provName, envVar := getModelProviderKeyEnv(modelName)
			isReady, statusStr := isModelReady(modelName)

			// Find pools using this model
			var mappedPools []string
			for pName, def := range supervisor.MasterPoolCatalog {
				allM := append(append(def.PrimaryTier.Models, def.BackupTier1.Models...), def.BackupTier2.Models...)
				for _, cand := range allM {
					if cand == modelName {
						mappedPools = append(mappedPools, pName)
						break
					}
				}
			}
			sort.Strings(mappedPools)
			mappedStr := strings.Join(mappedPools, ", ")
			if len(mappedStr) > 30 {
				mappedStr = mappedStr[:27] + "..."
			}

			mLabel := widget.NewLabelWithStyle(fmt.Sprintf("%d. %s", idx+1, modelName), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true, Bold: true})
			provLabel := widget.NewLabelWithStyle(provName, fyne.TextAlignLeading, fyne.TextStyle{})
			envLabel := mutedText(fmt.Sprintf("[%s]", envVar), 9, true)
			poolsLabel := mutedText(mappedStr, 9, false)
			badge := createPillBadge(statusStr, isReady)

			rowGrid := container.NewGridWithColumns(5,
				mLabel,
				provLabel,
				envLabel,
				poolsLabel,
				container.NewHBox(badge),
			)

			rowBg := canvas.NewRectangle(color.RGBA{R: 0x14, G: 0x14, B: 0x18, A: 0xff})
			rowBg.CornerRadius = 4.0
			rowPadded := container.NewBorder(verticalSpacer(4), verticalSpacer(4), horizontalSpacer(8), horizontalSpacer(8), rowGrid)
			catalogBox.Add(container.NewStack(rowBg, rowPadded))
			catalogBox.Add(verticalSpacer(3))
		}

		scroll := container.NewVScroll(catalogBox)
		return createStitchCard(
			fmt.Sprintf("Distinct Model Catalog (%d Unique Models in Mesh)", len(distinctModels)),
			theme.ListIcon(),
			mutedText("Cross-Pool Model Federation Inventory", 11, true),
			scroll,
		)
	}

	// -------------------------------------------------------------------------
	// Subview 3: Upstream Provider Registry & Credentials (10 Providers)
	// -------------------------------------------------------------------------
	buildProvidersView := func() fyne.CanvasObject {
		providersBox := container.NewVBox()

		headerGrid := container.NewGridWithColumns(5,
			widget.NewLabelWithStyle("PROVIDER NAME", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
			widget.NewLabelWithStyle("ENVIRONMENT VARIABLE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
			widget.NewLabelWithStyle("MAPPED MODELS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
			widget.NewLabelWithStyle("CREDENTIAL STATUS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
			widget.NewLabelWithStyle("ACTIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
		)
		providersBox.Add(headerGrid)
		providersBox.Add(widget.NewSeparator())
		providersBox.Add(verticalSpacer(6))

		for _, p := range allProvidersList {
			spec := p
			isKeyless := spec.EnvVar == "POLLINATIONS_PUBLIC"

			var configured bool
			var statusStr string
			if isKeyless {
				configured = true
				statusStr = "Zero-Auth (Public)"
			} else {
				val := os.Getenv(spec.EnvVar)
				configured = config.IsConfiguredKey(val)
				statusStr = "Not Configured"
				if configured {
					if len(val) > 10 {
						statusStr = fmt.Sprintf("Configured (%s...%s)", val[:4], val[len(val)-4:])
					} else {
						statusStr = "Configured"
					}
				} else if val != "" {
					statusStr = "Placeholder (.env)"
				}
			}

			// Count models under this provider
			_, _, distinctModels := getDistinctModelStats()
			provModelCount := 0
			for _, m := range distinctModels {
				prov, _ := getModelProviderKeyEnv(m)
				if strings.Contains(strings.ToLower(prov), strings.ToLower(strings.Split(spec.Name, " ")[0])) {
					provModelCount++
				}
			}

			pNameLbl := widget.NewLabelWithStyle(spec.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			pEnvLbl := widget.NewLabelWithStyle(fmt.Sprintf("[%s]", spec.EnvVar), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
			countLbl := widget.NewLabelWithStyle(fmt.Sprintf("%d Models", provModelCount), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
			badge := createPillBadge(statusStr, configured)

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
				// Matching Key icon button and Website Globe icon button side by side
				setKeyBtn := widget.NewButtonWithIcon("", keyIcon, func() {
					a.showProviderKeyDialog(spec, refreshAllViews)
				})
				setKeyBtn.Importance = widget.LowImportance

				openWebBtn := widget.NewButtonWithIcon("", websiteIcon, func() {
					if u, err := url.Parse(spec.ConsoleURL); err == nil {
						a.app.OpenURL(u)
					}
					a.showProviderKeyDialog(spec, refreshAllViews)
				})
				openWebBtn.Importance = widget.LowImportance

				actionBtn = container.NewHBox(setKeyBtn, horizontalSpacer(4), openWebBtn)
			}

			rowGrid := container.NewGridWithColumns(5,
				pNameLbl,
				pEnvLbl,
				countLbl,
				container.NewHBox(badge),
				actionBtn,
			)

			rowBg := canvas.NewRectangle(color.RGBA{R: 0x14, G: 0x14, B: 0x18, A: 0xff})
			rowBg.CornerRadius = 4.0
			rowPadded := container.NewBorder(verticalSpacer(4), verticalSpacer(4), horizontalSpacer(8), horizontalSpacer(8), rowGrid)
			providersBox.Add(container.NewStack(rowBg, rowPadded))
			providersBox.Add(verticalSpacer(4))
		}

		scroll := container.NewVScroll(providersBox)
		return createStitchCard(
			"Upstream Provider Registry (10 Providers — docs/report.md Section 4)",
			theme.ComputerIcon(),
			mutedText("Free-Tier Identity & Credential Vault", 11, true),
			scroll,
		)
	}

	// -------------------------------------------------------------------------
	// Top Ribbon: 1 Single Cohesive Navigation Switcher (Fix Point 4)
	// -------------------------------------------------------------------------
	buildTopRibbon := func() fyne.CanvasObject {
		totalDistinct, activeDistinct, _ := getDistinctModelStats()
		configuredKeys := 0
		for _, p := range allProvidersList {
			if p.EnvVar == "POLLINATIONS_PUBLIC" || config.IsConfiguredKey(os.Getenv(p.EnvVar)) {
				configuredKeys++
			}
		}

		createNavCard := func(index int, title, mainVal, subVal string) fyne.CanvasObject {
			isSelected := currentSubView == index

			var titleLbl fyne.CanvasObject
			if isSelected {
				titleLbl = widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			} else {
				titleLbl = mutedText(title, 11, true)
			}
			valLbl := widget.NewLabelWithStyle(mainVal, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true, Bold: true})
			subLbl := mutedText(subVal, 10, false)

			box := container.NewVBox(
				titleLbl,
				verticalSpacer(4),
				valLbl,
				verticalSpacer(2),
				subLbl,
			)
			padded := container.NewBorder(verticalSpacer(10), verticalSpacer(12), horizontalSpacer(16), horizontalSpacer(16), box)

			return newHoverCard(
				padded,
				isSelected,
				color.RGBA{R: 0x13, G: 0x14, B: 0x18, A: 0xff},
				color.RGBA{R: 0x1c, G: 0x22, B: 0x30, A: 0xff},
				color.RGBA{R: 0x17, G: 0x24, B: 0x3b, A: 0xff},
				color.RGBA{R: 0x27, G: 0x27, B: 0x32, A: 0xff},
				color.RGBA{R: 0x47, G: 0x55, B: 0x69, A: 0xff},
				color.RGBA{R: 0x38, G: 0xbd, B: 0xf8, A: 0xff},
				func() {
					currentSubView = index
					refreshAllViews()
				},
			)
		}

		tile1 := createNavCard(0, "1. CLUSTER CAPABILITY POOLS", "6 Active / 14 Total", "4×4×4 Multi-Tier Fallback Arrays")
		tile2 := createNavCard(1, "2. DISTINCT MODEL CATALOG", fmt.Sprintf("%d Active / %d Distinct", activeDistinct, totalDistinct), "Cross-Pool Global Inventory")
		tile3 := createNavCard(2, "3. UPSTREAM PROVIDER VAULT", fmt.Sprintf("%d / 10 Ready", configuredKeys), "Zero-Trust Identity & Keys")

		return container.NewGridWithColumns(3, tile1, tile2, tile3)
	}

	refreshAllViews = func() {
		ribbonContainer.Objects = []fyne.CanvasObject{buildTopRibbon()}
		ribbonContainer.Refresh()

		switch currentSubView {
		case 0:
			mainContentContainer.Objects = []fyne.CanvasObject{buildPoolsView()}
		case 1:
			mainContentContainer.Objects = []fyne.CanvasObject{buildCatalogView()}
		case 2:
			mainContentContainer.Objects = []fyne.CanvasObject{buildProvidersView()}
		}
		mainContentContainer.Refresh()
	}

	ribbonContainer.Objects = []fyne.CanvasObject{buildTopRibbon()}
	mainContentContainer.Objects = []fyne.CanvasObject{buildPoolsView()}

	content := container.NewBorder(
		container.NewVBox(ribbonContainer, verticalSpacer(8)),
		nil, nil, nil,
		mainContentContainer,
	)

	return wrapWithTabMargins(content)
}
