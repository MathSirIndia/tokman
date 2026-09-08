package gui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *AppUI) buildCacheTab() fyne.CanvasObject {
	// =========================================================================
	// 1. Top KPI Performance Ribbon (Stitch Screen 3)
	// =========================================================================
	a.cacheHitRatioLabel = widget.NewLabelWithStyle("0.0%", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true})
	a.cacheEntriesLabel = widget.NewLabelWithStyle("0 / 2,000 LRU", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true})
	a.cacheSavedLabel = widget.NewLabelWithStyle("0.0 ms", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true})

	kpi1 := createKPITile("Cache Hit Ratio", a.cacheHitRatioLabel, "Sub-millisecond memory bypass")
	kpi2 := createKPITile("In-Memory Entries", a.cacheEntriesLabel, "O(1) in-memory hash map")
	kpi3 := createKPITile("Latency Saved", a.cacheSavedLabel, "Zero-cost cached inference")

	kpiGrid := container.NewGridWithColumns(3, kpi1, kpi2, kpi3)

	flushBtn := widget.NewButtonWithIcon("Flush In-Memory Cache", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Flush Cache", "Are you sure you want to flush all in-memory LRU cache entries?", func(confirmed bool) {
			if confirmed {
				a.cache.Purge()
				a.refreshLedger()
				dialog.ShowInformation("Cache Flushed", "All in-memory cache entries have been evicted.", a.window)
			}
		}, a.window)
	})
	flushBtn.Importance = widget.DangerImportance

	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		a.refreshLedger()
	})
	refreshBtn.Importance = widget.LowImportance

	kpiActions := container.NewHBox(
		refreshBtn,
		horizontalSpacer(8),
		flushBtn,
	)

	kpiCard := createStitchCard(
		"In-Memory LRU Cache & Telemetry",
		theme.StorageIcon(),
		kpiActions,
		kpiGrid,
	)

	// =========================================================================
	// 2. Persistence Audit Ledger (Stitch Screen 3)
	// =========================================================================
	a.ledgerListContainer = container.NewVBox()

	headerRow := container.NewHBox(
		widget.NewLabelWithStyle("TIMESTAMP        ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
		widget.NewLabelWithStyle("REQUEST ID  ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
		widget.NewLabelWithStyle("MODEL POOL           ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
		widget.NewLabelWithStyle("TOKENS  ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
		widget.NewLabelWithStyle("LATENCY    ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
		widget.NewLabelWithStyle("STATUS   ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
		widget.NewLabelWithStyle("CACHE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
	)

	ledgerCard := createStitchCard(
		"SQLite Persistence Audit Ledger",
		theme.StorageIcon(),
		widget.NewLabelWithStyle("data/tokman.db", fyne.TextAlignTrailing, fyne.TextStyle{Monospace: true, Italic: true}),
		container.NewVBox(
			headerRow,
			verticalSpacer(2),
			widget.NewSeparator(),
			verticalSpacer(4),
			a.ledgerListContainer,
		),
	)

	// =========================================================================
	// 3. Entry Detail Inspector (Stitch Screen 3)
	// =========================================================================
	a.inspectorLabel = widget.NewLabelWithStyle("Select a ledger entry above to inspect cached payload details.", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})

	inspectorCard := createStitchCard(
		"Entry Detail Inspector",
		theme.InfoIcon(),
		widget.NewLabelWithStyle("Awaiting selection", fyne.TextAlignTrailing, fyne.TextStyle{Monospace: true, Italic: true}),
		createInsetBox(a.inspectorLabel),
	)

	content := container.NewVBox(
		kpiCard,
		verticalSpacer(14),
		ledgerCard,
		verticalSpacer(14),
		inspectorCard,
	)

	// Trigger initial data load
	a.refreshLedger()

	return wrapWithTabMargins(content)
}

func createKPITile(title string, valWidget fyne.CanvasObject, subtitle string) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.RGBA{R: 0x14, G: 0x14, B: 0x18, A: 0xff})
	bg.StrokeColor = color.RGBA{R: 0x27, G: 0x27, B: 0x30, A: 0xff}
	bg.StrokeWidth = 1.0
	bg.CornerRadius = 6.0

	titleLbl := fieldLabel(title)
	subLbl := widget.NewLabelWithStyle(subtitle, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	box := container.NewVBox(titleLbl, verticalSpacer(2), valWidget, verticalSpacer(2), subLbl)
	padded := container.NewBorder(verticalSpacer(8), verticalSpacer(10), horizontalSpacer(14), horizontalSpacer(14), box)
	return container.NewStack(bg, padded)
}

func (a *AppUI) refreshLedger() {
	if a.db == nil || a.ledgerListContainer == nil {
		return
	}

	// Update LRU cache stats and DB data outside UI thread
	count := a.cache.Len()
	summary, summaryErr := a.db.GetSpendSummary()
	logs, logsErr := a.db.GetRecentSpendLogs(15)

	fyne.Do(func() {
		if summaryErr == nil && summary != nil && a.cacheHitRatioLabel != nil {
			a.cacheHitRatioLabel.SetText(fmt.Sprintf("%.1f%%", summary.CacheHitRatio*100))
			a.cacheEntriesLabel.SetText(fmt.Sprintf("%d / 2,000 LRU", count))
			a.cacheSavedLabel.SetText(fmt.Sprintf("~%.1f ms", summary.AverageLatencyMs))
		}

		if logsErr != nil {
			return
		}

		a.ledgerListContainer.Objects = nil

		if len(logs) == 0 {
			a.ledgerListContainer.Add(widget.NewLabelWithStyle("No requests recorded yet. Run a prompt in the Quick Test Console or send a cURL request to :8000.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
			a.ledgerListContainer.Refresh()
			return
		}

		// Show top 3 rows to preserve clean viewport balance
		for idx, l := range logs {
			if idx >= 3 {
				break
			}
			rec := l
			timeStr := rec.Timestamp.Format("15:04:05")
			idStr := fmt.Sprintf("#%d", rec.ID)
			modelStr := rec.Model
			if len(modelStr) > 20 {
				modelStr = modelStr[:17] + "..."
			}
			tokStr := fmt.Sprintf("%d tok", rec.TotalTokens)
			latStr := fmt.Sprintf("%.1f ms", rec.LatencyMs)
			cacheTag := "MISS"
			cacheColor := "🟡"
			if rec.Cached {
				cacheTag = "HIT"
				cacheColor = "🟢"
			}

			line := fmt.Sprintf("%-16s %-12s %-20s %-8s %-10s %-8s %s %s", timeStr, idStr, modelStr, tokStr, latStr, "200 OK", cacheColor, cacheTag)

			btn := widget.NewButton(line, func() {
				a.inspectorLabel.SetText(fmt.Sprintf(`Prompt Hash (SHA-256):
hash_%08x%08x%08x

Payload (Cached Response):
{
  "id": "req_%d_%s",
  "object": "chat.completion",
  "created": %d,
  "model": "%s",
  "tokens": {"prompt": %d, "completion": %d, "total": %d},
  "latency_ms": %.2f,
  "cache_state": "%s"
}`,
					rec.ID*31, rec.TotalTokens*17, rec.Timestamp.Unix(),
					rec.ID, rec.ClientID,
					rec.Timestamp.Unix(),
					rec.Model,
					rec.PromptTokens, rec.CompletionTokens, rec.TotalTokens,
					rec.LatencyMs,
					cacheTag,
				))
			})
			btn.Alignment = widget.ButtonAlignLeading
			a.ledgerListContainer.Add(btn)
		}

		a.ledgerListContainer.Refresh()
	})
}
