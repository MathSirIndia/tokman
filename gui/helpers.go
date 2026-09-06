package gui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// verticalSpacer creates an empty vertical gap of specified pixel height
func verticalSpacer(height float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(1, height))
	return rect
}

// horizontalSpacer creates an empty horizontal gap of specified pixel width
func horizontalSpacer(width float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(width, 1))
	return rect
}

// sectionDivider returns a subtle horizontal separator line with vertical breathing room
func sectionDivider(spacing float32) fyne.CanvasObject {
	return container.NewVBox(
		verticalSpacer(spacing),
		widget.NewSeparator(),
		verticalSpacer(spacing),
	)
}

// fieldLabel creates a clean, uppercase muted label matching Stitch design
func fieldLabel(text string) *widget.Label {
	lbl := widget.NewLabelWithStyle(strings.ToUpper(text), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return lbl
}

// mutedText returns a compact canvas text element without widget padding
func mutedText(text string, size float32, bold bool) *canvas.Text {
	t := canvas.NewText(text, color.RGBA{R: 0x8e, G: 0x8e, B: 0x9a, A: 0xff})
	t.TextSize = size
	t.TextStyle = fyne.TextStyle{Bold: bold}
	return t
}

// primaryText returns crisp white canvas text
func primaryText(text string, size float32, bold bool, mono bool) *canvas.Text {
	t := canvas.NewText(text, color.RGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff})
	t.TextSize = size
	t.TextStyle = fyne.TextStyle{Bold: bold, Monospace: mono}
	return t
}

// createStitchCard constructs a distinct elevated dark container with a visible 1px border (#2a2a33),
// 6px rounded corners, header bar with icon and title, right-side badge/actions, and compact padded content.
func createStitchCard(title string, icon fyne.Resource, headerRight fyne.CanvasObject, content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.RGBA{R: 0x16, G: 0x16, B: 0x1a, A: 0xff})
	bg.StrokeColor = color.RGBA{R: 0x2a, G: 0x2a, B: 0x33, A: 0xff}
	bg.StrokeWidth = 1.0
	bg.CornerRadius = 6.0

	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	var leftHeader fyne.CanvasObject
	if icon != nil {
		leftHeader = container.NewHBox(widget.NewIcon(icon), horizontalSpacer(6), titleLabel)
	} else {
		leftHeader = container.NewHBox(titleLabel)
	}

	headerBorder := container.NewBorder(nil, nil, leftHeader, headerRight)

	headerBox := container.NewVBox(
		headerBorder,
		verticalSpacer(6),
		widget.NewSeparator(),
		verticalSpacer(10),
	)

	innerBox := container.NewBorder(
		headerBox,
		nil,
		nil,
		nil,
		content,
	)

	padded := container.NewBorder(
		verticalSpacer(10),
		verticalSpacer(12),
		horizontalSpacer(16),
		horizontalSpacer(16),
		innerBox,
	)

	return container.NewStack(bg, padded)
}

// createInsetBox creates an inset, darker bordered card (#0c0c0f with #24242c border)
func createInsetBox(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.RGBA{R: 0x0c, G: 0x0c, B: 0x0f, A: 0xff})
	bg.StrokeColor = color.RGBA{R: 0x24, G: 0x24, B: 0x2c, A: 0xff}
	bg.StrokeWidth = 1.0
	bg.CornerRadius = 4.0

	padded := container.NewBorder(
		verticalSpacer(8),
		verticalSpacer(8),
		horizontalSpacer(12),
		horizontalSpacer(12),
		content,
	)

	return container.NewStack(bg, padded)
}

// createPillBadge creates a Stitch-style pill badge (e.g. Active, Standby, Configured)
func createPillBadge(text string, isSuccess bool) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.RGBA{R: 0x12, G: 0x12, B: 0x16, A: 0xff})
	bg.CornerRadius = 4.0
	bg.StrokeWidth = 1.0

	var dotColor color.Color
	var textColor color.Color
	if isSuccess {
		bg.StrokeColor = color.RGBA{R: 0x1b, G: 0x47, B: 0x2e, A: 0xff}
		dotColor = color.RGBA{R: 0x10, G: 0xb9, B: 0x81, A: 0xff}
		textColor = color.RGBA{R: 0x34, G: 0xd3, B: 0x99, A: 0xff}
	} else {
		bg.StrokeColor = color.RGBA{R: 0x47, G: 0x33, B: 0x1b, A: 0xff}
		dotColor = color.RGBA{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff}
		textColor = color.RGBA{R: 0xfb, G: 0xbf, B: 0x24, A: 0xff}
	}

	dot := canvas.NewCircle(dotColor)
	dot.StrokeWidth = 0
	dotContainer := container.NewStack(dot)
	dotContainer.Resize(fyne.NewSize(6, 6))

	lbl := canvas.NewText(text, textColor)
	lbl.TextSize = 11.0
	lbl.TextStyle = fyne.TextStyle{Bold: true}

	content := container.NewHBox(
		horizontalSpacer(2),
		container.NewCenter(dot),
		lbl,
		horizontalSpacer(2),
	)

	padded := container.NewBorder(
		verticalSpacer(2),
		verticalSpacer(2),
		horizontalSpacer(6),
		horizontalSpacer(6),
		content,
	)

	return container.NewStack(bg, padded)
}

// createMethodBadge creates an HTTP method pill (e.g. POST in emerald, GET in purple/indigo)
func createMethodBadge(method string) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.RGBA{R: 0x12, G: 0x12, B: 0x16, A: 0xff})
	bg.CornerRadius = 3.0
	bg.StrokeWidth = 1.0

	var textColor color.Color
	if method == "POST" {
		bg.StrokeColor = color.RGBA{R: 0x1b, G: 0x47, B: 0x2e, A: 0xff}
		textColor = color.RGBA{R: 0x34, G: 0xd3, B: 0x99, A: 0xff} // Emerald
	} else {
		bg.StrokeColor = color.RGBA{R: 0x2e, G: 0x2d, B: 0x5a, A: 0xff}
		textColor = color.RGBA{R: 0x81, G: 0x8c, B: 0xf8, A: 0xff} // Indigo
	}

	lbl := canvas.NewText(method, textColor)
	lbl.TextSize = 10.0
	lbl.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	padded := container.NewBorder(
		verticalSpacer(1),
		verticalSpacer(1),
		horizontalSpacer(5),
		horizontalSpacer(5),
		lbl,
	)
	return container.NewStack(bg, padded)
}

// wrapWithTabMargins encloses tab content in standard compact margins
func wrapWithTabMargins(content fyne.CanvasObject) fyne.CanvasObject {
	bordered := container.NewBorder(
		verticalSpacer(12),
		verticalSpacer(16),
		horizontalSpacer(18),
		horizontalSpacer(18),
		content,
	)
	return container.NewVScroll(bordered)
}

