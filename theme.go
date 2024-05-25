//go:generate fyne bundle -o bundled.go assets
package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type GoDoTheme struct {
	fyne.Theme
}

func newGoDoTheme() fyne.Theme {
	return &GoDoTheme{Theme: theme.DefaultTheme()}
}

func (t *GoDoTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(name, theme.VariantDark)
}

func (t *GoDoTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 12
	}
	return t.Theme.Size(name)

}
