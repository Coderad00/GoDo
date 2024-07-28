package main

import (
	"fyne"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
)

func main() {
	a := app.NewWithID("GoDo")
	w := a.NewWindow("GoDo")
	w.SetIcon(ResourceLogoWindowmanagerWhitePng)
	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("GoDo",
			fyne.NewMenuItem("show", func() {
				w.Show()
			}))
		desk.SetSystemTrayMenu(m)
		desk.SetSystemTrayIcon(ResourceLogoWindowmanagerWhitePng)
	}

	w.SetContent(makeGUI(a, w))
	w.Resize(fyne.NewSize(400, 200))
	w.ShowAndRun()
}
