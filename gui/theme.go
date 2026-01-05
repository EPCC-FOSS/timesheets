package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme implements a teal, purple, and black color scheme
type CustomTheme struct{}

var _ fyne.Theme = (*CustomTheme)(nil)

// Color palette
var (
	// Primary colors
	primaryPurple = color.NRGBA{R: 156, G: 39, B: 176, A: 255}  // Deep Purple
	lightPurple   = color.NRGBA{R: 186, G: 104, B: 200, A: 255} // Light Purple
	darkPurple    = color.NRGBA{R: 106, G: 27, B: 154, A: 255}  // Dark Purple

	primaryTeal = color.NRGBA{R: 0, G: 150, B: 136, A: 255}   // Teal
	lightTeal   = color.NRGBA{R: 77, G: 182, B: 172, A: 255}  // Light Teal
	darkTeal    = color.NRGBA{R: 0, G: 121, B: 107, A: 255}   // Dark Teal
	accentTeal  = color.NRGBA{R: 128, G: 203, B: 196, A: 255} // Accent Teal

	// Neutral colors
	black         = color.NRGBA{R: 18, G: 18, B: 18, A: 255}    // Near Black
	darkGray      = color.NRGBA{R: 33, G: 33, B: 33, A: 255}    // Dark Gray
	mediumGray    = color.NRGBA{R: 66, G: 66, B: 66, A: 255}    // Medium Gray
	lightGray     = color.NRGBA{R: 158, G: 158, B: 158, A: 255} // Light Gray
	veryLightGray = color.NRGBA{R: 238, G: 238, B: 238, A: 255} // Very Light Gray
	white         = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // White

	// Semantic colors
	errorRed      = color.NRGBA{R: 211, G: 47, B: 47, A: 255} // Error Red
	successGreen  = color.NRGBA{R: 76, G: 175, B: 80, A: 255} // Success Green
	warningOrange = color.NRGBA{R: 255, G: 152, B: 0, A: 255} // Warning Orange
)

func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	// Primary color (buttons, highlights, etc.) - TEAL is now primary
	case theme.ColorNamePrimary:
		return primaryTeal

	// Button colors - TEAL for main buttons
	case theme.ColorNameButton:
		return primaryTeal
	case theme.ColorNameDisabledButton:
		return mediumGray
	case theme.ColorNameHover:
		return lightTeal
	case theme.ColorNamePressed:
		return darkTeal

	// Background colors
	case theme.ColorNameBackground:
		if variant == theme.VariantLight {
			return veryLightGray
		}
		return black
	case theme.ColorNameOverlayBackground:
		return darkGray

	// Foreground (text) colors
	case theme.ColorNameForeground:
		if variant == theme.VariantLight {
			return black
		}
		return white
	case theme.ColorNameDisabled:
		return lightGray
	case theme.ColorNamePlaceHolder:
		return mediumGray

	// Focus and selection - Purple for accents
	case theme.ColorNameFocus:
		return lightPurple
	case theme.ColorNameSelection:
		return lightPurple

	// Input fields
	case theme.ColorNameInputBackground:
		if variant == theme.VariantLight {
			return white
		}
		return darkGray
	case theme.ColorNameInputBorder:
		return primaryTeal

	// Semantic colors
	case theme.ColorNameError:
		return errorRed
	case theme.ColorNameSuccess:
		return successGreen
	case theme.ColorNameWarning:
		return warningOrange

	// Scrollbar
	case theme.ColorNameScrollBar:
		return mediumGray

	// Shadow
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 66}

	// Header background - Teal for headers
	case theme.ColorNameHeaderBackground:
		return darkTeal

	// Menu background
	case theme.ColorNameMenuBackground:
		if variant == theme.VariantLight {
			return white
		}
		return darkGray

	// Hyperlink - Purple as secondary accent
	case theme.ColorNameHyperlink:
		return lightPurple

	default:
		// Fall back to default theme for any unhandled colors
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 8
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInputRadius:
		return 4
	default:
		return theme.DefaultTheme().Size(name)
	}
}
