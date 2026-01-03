package efinui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MonocolorTheme struct {
	ThemeColor color.Color
	Variant    string
}

func (t MonocolorTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if t.Variant != "" {
		if t.Variant == "dark" {
			variant = theme.VariantDark
		} else {
			variant = theme.VariantLight
		}
	}

	switch name {
	case theme.ColorNameBackground, theme.ColorNameDisabled, theme.ColorNameInputBackground,
		theme.ColorNamePlaceHolder, theme.ColorNameScrollBar,
		theme.ColorNameScrollBarBackground:

		return theme.DefaultTheme().Color(name, variant)

	case theme.ColorNameShadow:
		return color.RGBA{0, 0, 0, 0}

	case theme.ColorNameHover:
		r, g, b, _ := t.ThemeColor.RGBA()
		return color.RGBA64{uint16(r), uint16(g), uint16(b), 0x3000}

	case theme.ColorNameSelection:
		r, g, b, _ := t.ThemeColor.RGBA()
		f := 0.5
		return color.RGBA64{uint16(float64(r) * f), uint16(float64(g) * f), uint16(float64(b) * f), 0x8FFF}

	case theme.ColorNameHeaderBackground:
		r, g, b, _ := t.ThemeColor.RGBA()
		f := 0.5
		return color.RGBA64{uint16(float64(r) * f), uint16(float64(g) * f), uint16(float64(b) * f), 0xFFFF}

	case theme.ColorNamePrimary, theme.ColorNameForeground, theme.ColorNameSeparator, theme.ColorNameInputBorder:
		return t.ThemeColor
	}

	return t.ThemeColor
}

func (t MonocolorTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t MonocolorTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t MonocolorTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
