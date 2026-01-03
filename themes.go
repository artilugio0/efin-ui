package efinui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	ColorNameActiveBackground fyne.ThemeColorName = "activeBackground"
)

type CustomTheme struct {
	ActiveBackground    color.Color
	Background          color.Color
	Disabled            color.Color
	Foreground          color.Color
	HeaderBackground    color.Color
	Hover               color.Color
	InputBackground     color.Color
	InputBorder         color.Color
	PlaceHolder         color.Color
	Primary             color.Color
	ScrollBar           color.Color
	ScrollBarBackground color.Color
	Selection           color.Color
	Separator           color.Color
	Shadow              color.Color
}

func (ct CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case ColorNameActiveBackground:
		return ct.ActiveBackground

	case theme.ColorNameBackground:
		return ct.Background

	case theme.ColorNameDisabled:
		return ct.Disabled

	case theme.ColorNameForeground:
		return ct.Foreground

	case theme.ColorNameHeaderBackground:
		return ct.HeaderBackground

	case theme.ColorNameHover:
		return ct.Hover

	case theme.ColorNameInputBackground:
		return ct.InputBackground

	case theme.ColorNameInputBorder:
		return ct.InputBorder

	case theme.ColorNamePlaceHolder:
		return ct.PlaceHolder

	case theme.ColorNamePrimary:
		return ct.Primary

	case theme.ColorNameScrollBar:
		return ct.ScrollBar

	case theme.ColorNameScrollBarBackground:
		return ct.ScrollBarBackground

	case theme.ColorNameSelection:
		return ct.Selection

	case theme.ColorNameSeparator:
		return ct.Separator

	case theme.ColorNameShadow:
		return ct.Shadow

	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (ct CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (ct CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (ct CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
