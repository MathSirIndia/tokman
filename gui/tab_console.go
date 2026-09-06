package gui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

var allPoolOptions = []string{
	"pool/auto",
	"pool/general",
	"pool/deep-reasoning",
	"pool/agent-coding",
	"pool/document-analysis",
	"pool/web-research",
	"pool/presentation",
	"pool/image-gen",
	"pool/architect",
	"pool/security-tester",
	"pool/stack-optimizer",
	"pool/document-gen",
	"pool/audio-gen",
	"pool/video-conductor",
}

func getPoolOptionsExcluding(selected string) []string {
	opts := make([]string, 0, len(allPoolOptions)-1)
	for _, p := range allPoolOptions {
		if p != selected {
			opts = append(opts, p)
		}
	}
	return opts
}

func getTargetModelNoteText(s string) string {
	switch s {
	case "pool/auto":
		return "Target: Two-Tier Hierarchical Intent Arbitrator (<40ms DAG)"
	case "pool/general":
		return "Target: groq/llama-3.3-70b on Groq LPU (Tier 1 Flagship)"
	case "pool/deep-reasoning":
		return "Target: sambanova/deepseek-r1 (Chain-of-Thought R1)"
	case "pool/agent-coding":
		return "Target: sambanova/qwen-2.5-coder (Autonomous Dev)"
	case "pool/document-analysis":
		return "Target: gemini/gemini-2.0-flash (1M Token Context Window)"
	case "pool/web-research":
		return "Target: cerebras/llama-3.3-70b with SearXNG Grounding"
	case "pool/presentation":
		return "Target: groq/llama-3.3-70b (Headless Marp / Reveal.js)"
	case "pool/image-gen":
		return "Target: cf/flux-1-schnell (Cloudflare Diffusion Engine)"
	case "pool/architect":
		return "Target: gemini/gemini-2.0-flash (Scene Graph Planner)"
	case "pool/security-tester":
		return "Target: sambanova/qwen-coder (AST Static Analysis)"
	case "pool/stack-optimizer":
		return "Target: gemini/gemini-2.0-flash (Token Burn Aggregator)"
	case "pool/document-gen":
		return "Target: gemini/gemini-2.0-flash (Typst Rust Compiler)"
	case "pool/audio-gen":
		return "Target: cf/openai-whisper (Kokoro-82M TTS / Whisper)"
	case "pool/video-conductor":
		return "Target: Deterministic FFmpeg Ken Burns Transform Runner"
	default:
		return "Target: " + s
	}
}

func (a *AppUI) updateConsolePool(selected string) {
	if a.consolePoolSelect != nil {
		a.consolePoolSelect.Options = getPoolOptionsExcluding(selected)
		a.consolePoolSelect.Selected = selected
		a.consolePoolSelect.Refresh()
	}
	if a.consoleTargetNote != nil {
		a.consoleTargetNote.SetText(getTargetModelNoteText(selected))
	}
}

func (a *AppUI) buildConsoleTab() fyne.CanvasObject {
	// =========================================================================
	// Left: Request Composer (Stitch Screen 2)
	// =========================================================================
	a.consoleTargetNote = widget.NewLabelWithStyle(getTargetModelNoteText("pool/general"), fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	initialOptions := getPoolOptionsExcluding("pool/general")
	a.consolePoolSelect = widget.NewSelect(initialOptions, func(s string) {
		a.updateConsolePool(s)
	})
	a.consolePoolSelect.Selected = "pool/general"

	sysPromptEntry := widget.NewMultiLineEntry()
	sysPromptEntry.SetText("You are a concise, helpful assistant.")
	sysPromptEntry.TextStyle = fyne.TextStyle{Monospace: true}
	sysPromptEntry.SetMinRowsVisible(2)

	a.consolePromptEntry = widget.NewMultiLineEntry()
	a.consolePromptEntry.SetText("Explain LRU cache eviction in one paragraph.")
	a.consolePromptEntry.TextStyle = fyne.TextStyle{Monospace: true}
	a.consolePromptEntry.SetMinRowsVisible(3)

	maxTokensSlider := widget.NewSlider(64, 4096)
	maxTokensSlider.SetValue(1024)
	maxTokensLabel := widget.NewLabelWithStyle("1024", fyne.TextAlignTrailing, fyne.TextStyle{Monospace: true})
	maxTokensSlider.OnChanged = func(v float64) {
		maxTokensLabel.SetText(fmt.Sprintf("%d", int(v)))
	}

	tempSlider := widget.NewSlider(0.0, 1.5)
	tempSlider.SetValue(0.7)
	tempLabel := widget.NewLabelWithStyle("0.70", fyne.TextAlignTrailing, fyne.TextStyle{Monospace: true})
	tempSlider.OnChanged = func(v float64) {
		tempLabel.SetText(fmt.Sprintf("%.2f", v))
	}

	streamCheck := widget.NewCheck("Stream SSE", nil)
	streamCheck.SetChecked(false)

	paramsCard := createInsetBox(container.NewVBox(
		streamCheck,
		verticalSpacer(4),
		container.NewBorder(nil, nil, widget.NewLabel("Max Tokens"), maxTokensLabel, maxTokensSlider),
		verticalSpacer(4),
		container.NewBorder(nil, nil, widget.NewLabel("Temperature"), tempLabel, tempSlider),
	))

	// Right: Response Inspector
	statusPill := widget.NewLabelWithStyle("🟢 200 OK", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	latencyPill := widget.NewLabelWithStyle("⏱ 0 ms", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true})
	costPill := widget.NewLabelWithStyle("💲 $0.0000 Free Tier", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	tokensBanner := widget.NewLabelWithStyle("Tokens: 0 prompt + 0 completion = 0 total (0.0 tok/s)", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})

	responseBody := widget.NewMultiLineEntry()
	responseBody.TextStyle = fyne.TextStyle{Monospace: true}
	responseBody.Wrapping = fyne.TextWrapWord
	responseBody.SetMinRowsVisible(8)
	responseBody.Disable()

	sendBtn := widget.NewButtonWithIcon("Send Request", theme.MediaPlayIcon(), nil)
	sendBtn.Importance = widget.HighImportance

	clearBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		a.consolePromptEntry.SetText("")
		responseBody.SetText("")
		if a.consoleRawJsonEntry != nil {
			a.consoleRawJsonEntry.SetText("")
		}
		statusPill.SetText("READY")
		latencyPill.SetText("⏱ 0 ms")
		tokensBanner.SetText("Tokens: 0 prompt + 0 completion = 0 total")
	})
	clearBtn.Importance = widget.LowImportance

	sendBtn.OnTapped = func() {
		promptText := a.consolePromptEntry.Text
		if promptText == "" {
			return
		}

		sendBtn.Disable()
		statusPill.SetText("⏳ DISPATCHING...")
		responseBody.SetText("Connecting to TokMan Gateway (:8000)...")

		go func() {
			startTime := time.Now()

			reqPayload := ChatRequest{
				Model: a.consolePoolSelect.Selected,
				Messages: []ChatMessage{
					{Role: "system", Content: sysPromptEntry.Text},
					{Role: "user", Content: promptText},
				},
				MaxTokens:   int(maxTokensSlider.Value),
				Temperature: tempSlider.Value,
				Stream:      streamCheck.Checked,
			}

			payloadBytes, _ := json.Marshal(reqPayload)
			url := fmt.Sprintf("http://127.0.0.1:%d/v1/chat/completions", a.gwServer.Config.Port)

			httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
			if err != nil {
				a.updateUI(func() {
					sendBtn.Enable()
					statusPill.SetText("❌ ERROR")
					responseBody.SetText(fmt.Sprintf("Failed to create request: %v", err))
				})
				return
			}

			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("Authorization", "Bearer "+a.gwServer.Config.MasterKey)

			client := &http.Client{Timeout: 45 * time.Second}
			resp, err := client.Do(httpReq)
			elapsed := time.Since(startTime)

			if err != nil {
				a.updateUI(func() {
					sendBtn.Enable()
					statusPill.SetText("❌ FAILED")
					responseBody.SetText(fmt.Sprintf("Request failed: %v", err))
				})
				return
			}
			defer resp.Body.Close()

			bodyBytes, _ := io.ReadAll(resp.Body)
			cacheHeader := resp.Header.Get("X-Cache")
			cacheTag := "MISS"
			if cacheHeader == "HIT" {
				cacheTag = "HIT"
			}

			// Format raw JSON for the accordion text field
			var rawJSONText string
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
				rawJSONText = prettyJSON.String()
			} else {
				rawJSONText = string(bodyBytes)
			}

			var chatResp ChatResponse
			if err := json.Unmarshal(bodyBytes, &chatResp); err == nil && len(chatResp.Choices) > 0 {
				content := chatResp.Choices[0].Message.Content
				promptTok := chatResp.Usage.PromptTokens
				compTok := chatResp.Usage.CompletionTokens
				totalTok := chatResp.Usage.TotalTokens

				tokPerSec := 0.0
				if elapsed.Seconds() > 0 && compTok > 0 {
					tokPerSec = float64(compTok) / elapsed.Seconds()
				}

				a.updateUI(func() {
					sendBtn.Enable()
					if resp.StatusCode == 200 {
						statusPill.SetText(fmt.Sprintf("🟢 200 OK (%s)", cacheTag))
					} else {
						statusPill.SetText(fmt.Sprintf("⚠️ %d (%s)", resp.StatusCode, cacheTag))
					}
					latencyPill.SetText(fmt.Sprintf("⏱ %d ms", elapsed.Milliseconds()))
					tokensBanner.SetText(fmt.Sprintf("Tokens: %d prompt + %d completion = %d total (%.1f tok/s)", promptTok, compTok, totalTok, tokPerSec))
					responseBody.SetText(content)
					if a.consoleRawJsonEntry != nil {
						a.consoleRawJsonEntry.SetText(rawJSONText)
					}
					a.refreshLedger()
				})
			} else {
				a.updateUI(func() {
					sendBtn.Enable()
					statusPill.SetText(fmt.Sprintf("HTTP %d", resp.StatusCode))
					latencyPill.SetText(fmt.Sprintf("⏱ %d ms", elapsed.Milliseconds()))
					responseBody.SetText(string(bodyBytes))
					if a.consoleRawJsonEntry != nil {
						a.consoleRawJsonEntry.SetText(rawJSONText)
					}
					a.refreshLedger()
				})
			}
		}()
	}

	copyRespBtn := widget.NewButtonWithIcon("Copy Response", theme.ContentCopyIcon(), func() {
		a.window.Clipboard().SetContent(responseBody.Text)
	})
	copyRespBtn.Importance = widget.LowImportance

	composerScroll := container.NewVScroll(container.NewVBox(
		fieldLabel("Model Pool"),
		verticalSpacer(4),
		a.consolePoolSelect,
		verticalSpacer(2),
		a.consoleTargetNote,
		verticalSpacer(12),
		fieldLabel("System Prompt"),
		verticalSpacer(4),
		sysPromptEntry,
		verticalSpacer(12),
		fieldLabel("User Message"),
		verticalSpacer(4),
		a.consolePromptEntry,
		verticalSpacer(12),
		paramsCard,
	))

	actionsRow := container.NewBorder(
		verticalSpacer(10), nil, nil, clearBtn, sendBtn,
	)

	composerCard := createStitchCard(
		"Request Composer",
		theme.DocumentCreateIcon(),
		nil,
		container.NewBorder(nil, actionsRow, nil, nil, composerScroll),
	)

	metricsRow := container.NewHBox(
		statusPill,
		horizontalSpacer(12),
		latencyPill,
		horizontalSpacer(12),
		costPill,
	)

	infoBannerCard := createInsetBox(container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		horizontalSpacer(6),
		tokensBanner,
	))

	a.consoleRawJsonEntry = widget.NewMultiLineEntry()
	a.consoleRawJsonEntry.TextStyle = fyne.TextStyle{Monospace: true}
	a.consoleRawJsonEntry.Wrapping = fyne.TextWrapWord
	a.consoleRawJsonEntry.SetMinRowsVisible(6)
	a.consoleRawJsonEntry.Disable()
	a.consoleRawJsonEntry.SetText(`{"status": "ready", "protocol": "openai/v1"}`)

	rawJsonBox := createInsetBox(a.consoleRawJsonEntry)
	rawJsonBox.Hide()

	toggleJsonBtn := widget.NewButtonWithIcon("Raw JSON Response", theme.MenuDropDownIcon(), nil)
	toggleJsonBtn.Importance = widget.LowImportance
	toggleJsonBtn.OnTapped = func() {
		if rawJsonBox.Visible() {
			rawJsonBox.Hide()
			toggleJsonBtn.SetIcon(theme.MenuDropDownIcon())
		} else {
			rawJsonBox.Show()
			toggleJsonBtn.SetIcon(theme.MenuDropUpIcon())
		}
	}

	topInspector := container.NewVBox(
		metricsRow,
		verticalSpacer(8),
		infoBannerCard,
		verticalSpacer(10),
		container.NewBorder(
			nil, nil,
			fieldLabel("Output"),
			copyRespBtn,
		),
		verticalSpacer(4),
	)

	bottomInspector := container.NewVBox(
		verticalSpacer(6),
		container.NewBorder(nil, nil, toggleJsonBtn, nil),
		verticalSpacer(4),
		rawJsonBox,
	)

	inspectorContent := container.NewBorder(topInspector, bottomInspector, nil, nil, responseBody)

	inspectorCard := createStitchCard(
		"Response Inspector",
		theme.SearchIcon(),
		nil,
		inspectorContent,
	)

	leftPane := container.NewBorder(verticalSpacer(2), verticalSpacer(4), horizontalSpacer(4), horizontalSpacer(6), composerCard)
	rightPane := container.NewBorder(verticalSpacer(2), verticalSpacer(4), horizontalSpacer(6), horizontalSpacer(4), inspectorCard)

	split := container.NewHSplit(leftPane, rightPane)
	split.SetOffset(0.44) // 44% composer, 56% inspector matching Stitch Screen 2

	return container.NewBorder(
		verticalSpacer(12),
		verticalSpacer(16),
		horizontalSpacer(18),
		horizontalSpacer(18),
		split,
	)
}
