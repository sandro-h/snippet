package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	// ColorSnippetContent is the color for snippet content
	ColorSnippetContent fyne.ThemeColorName = "snippetContent"
)

// MyTheme is the custom snippet theme
type MyTheme struct{}

// Color returns the color for the color name.
func (m MyTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case ColorSnippetContent:
		return color.RGBA{128, 128, 128, 255}
	default:
		return theme.DarkTheme().Color(name, theme.VariantDark)
	}
}

// Icon returns the icon for the icon name.
func (m MyTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DarkTheme().Icon(name)
}

// Font returns the font for the text style.
func (m MyTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DarkTheme().Font(style)
}

// Size returns the size for the size name.
func (m MyTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 2
	}
	return theme.DarkTheme().Size(name)
}
