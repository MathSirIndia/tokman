package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// TokmanTheme implements fyne.Theme following docs/design.md
type TokmanTheme struct{}

var _ fyne.Theme = (*TokmanTheme)(nil)

func (t *TokmanTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 0x12, G: 0x12, B: 0x14, A: 0xff} // #121214 Canvas
	case theme.ColorNameButton:
		return color.RGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff} // #27272a
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 0x1c, G: 0x1c, B: 0x1e, A: 0xff}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 0x71, G: 0x71, B: 0x7a, A: 0xff} // #71717a Muted
	case theme.ColorNameForeground:
		return color.RGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff} // #f4f4f5 Primary Text
	case theme.ColorNameHover:
		return color.RGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff} // #27272a Hover
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 0x0a, G: 0x0a, B: 0x0c, A: 0xff} // #0a0a0c Inset Entry matching Stitch
	case theme.ColorNameInputBorder:
		return color.RGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff} // #27272a Border
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff} // #2563eb High-Contrast Royal Blue (WCAG AAA with white text)
	case theme.ColorNameSuccess:
		return color.RGBA{R: 0x10, G: 0xb9, B: 0x81, A: 0xff} // #10b981 Emerald
	case theme.ColorNameWarning:
		return color.RGBA{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff} // #f59e0b Amber
	case theme.ColorNameError:
		return color.RGBA{R: 0xef, G: 0x44, B: 0x44, A: 0xff} // #ef4444 Crimson Danger
	case theme.ColorNameShadow:
		return color.Transparent
	case theme.ColorNameSeparator:
		return color.RGBA{R: 0x27, G: 0x27, B: 0x2a, A: 0xff}
	case theme.ColorNameHeaderBackground:
		return color.RGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff} // #18181b Card Surface
	case theme.ColorNameMenuBackground:
		return color.RGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff}
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (t *TokmanTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *TokmanTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *TokmanTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6.0 // Balanced padding for widgets and containers
	case theme.SizeNameInnerPadding:
		return 6.0
	case theme.SizeNameLineSpacing:
		return 4.0 // Standard line spacing to prevent text smudging
	case theme.SizeNameInlineIcon:
		return 18.0
	case theme.SizeNameText:
		return 13.0
	default:
		return theme.DefaultTheme().Size(name)
	}
}
