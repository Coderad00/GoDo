package main

import (
	"fyne"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func makeGUI(a fyne.App, w fyne.Window) fyne.CanvasObject {
	banner := makeBanner(a, w)
	logo := makeLogo()
	todoListContainer := makeTodoListContainer()

	return container.NewVBox(
		logo,
		banner,
		container.NewStack(todoListContainer),
	)
}

func makeBanner(a fyne.App, w fyne.Window) fyne.CanvasObject {
	toolbar := makeToolbar(a, w)
	return container.NewVBox(toolbar)
}

func makeToolbar(a fyne.App, w fyne.Window) fyne.CanvasObject {
	addButton := widget.NewToolbarAction(theme.ContentAddIcon(), func() {
		showNewTodoWindow(a, w)
	})
	clearDoneButton := widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
		clearDoneTasks(a, w)
	})
	return widget.NewToolbar(addButton, clearDoneButton)
}

func makeTodoListContainer() fyne.CanvasObject {
	if len(todoList) > 0 {
		return container.NewVBox(buildTodoList(todoList)...)
	}
	return widget.NewLabel("No tasks available")
}

func makeLogo() fyne.CanvasObject {
	logo := canvas.NewImageFromResource(resourcePNGGODOLogoPng)
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(100, 50))
	return logo
}
