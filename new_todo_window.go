package main

import (
	"fmt"
	"fyne"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"time"
)

func showNewTodoWindow(a fyne.App, w fyne.Window) {
	inputWindow := a.NewWindow("New Todo")
	inputWindow.Resize(fyne.NewSize(300, 200))

	taskEntry := widget.NewEntry()
	taskEntry.SetPlaceHolder("Enter your task...")

	durationSelect := widget.NewSelect([]string{"10s", "1m", "15m", "30m", "1h", "3h"}, func(selected string) {
		fmt.Println("Selected duration:", selected)
	})

	saveCallback := func() {
		if taskEntry.Text == "" {
			fmt.Println("Please enter a task description.")
			return
		}

		duration, err := time.ParseDuration(durationSelect.Selected)
		if err != nil {
			fmt.Println("Invalid duration:", err)
			return
		}

		newItem := &TodoItem{
			Checkbox: widget.NewCheck("", func(checked bool) {
				fmt.Println("Checkbox state changed:", checked)
			}),
			Task:          widget.NewLabel(taskEntry.Text),
			Duration:      widget.NewLabel(durationSelect.Selected),
			Timer:         widget.NewLabel(formatTime(duration)),
			RemainingTime: duration,
		}

		todoList = append(todoList, newItem)
		w.SetContent(makeGUI(a, w))
		inputWindow.Close()
	}

	inputContainer := container.NewVBox(
		taskEntry,
		durationSelect,
		widget.NewButton("Save", saveCallback),
	)

	inputWindow.SetContent(inputContainer)
	inputWindow.Show()
}
