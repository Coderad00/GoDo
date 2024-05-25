package main

import (
	"fmt"
	"fyne/driver/desktop"
	"log"
	"os/exec"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bradhe/stopwatch"
)

type TodoItem struct {
	Checkbox      *widget.Check
	Task          *widget.Label
	Duration      *widget.Label
	Timer         *widget.Label
	Running       bool
	Done          chan bool
	RemainingTime time.Duration
	Stopwatch     stopwatch.Watch
}

func main() {
	a := app.NewWithID("GoDo")
	w := a.NewWindow("GoDo")

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("GoDo",
			fyne.NewMenuItem("show", func() {
				w.Show()
			}))
		desk.SetSystemTrayMenu(m)
		desk.SetSystemTrayIcon(resourceIconSystemTrayPng)
	}

	w.SetContent(makeGUI(a, w))
	w.Resize(fyne.NewSize(400, 200))
	w.ShowAndRun()
}

func (t *TodoItem) StartTimer() {
	if t.Running {
		println("TIMER ALREADY STARTED")
		return
	}

	if t.Done == nil {
		t.Done = make(chan bool)
	}

	if t.RemainingTime <= 0 {
		t.RemainingTime, _ = time.ParseDuration(t.Duration.Text)
	}

	t.Running = true
	go t.runTimer()
}

func (t *TodoItem) runTimer() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.RemainingTime -= time.Second
			if t.RemainingTime <= 0 {
				t.StopTimer()
				playSound()
				return
			}
			t.Timer.SetText(formatTime(t.RemainingTime))
		case <-t.Done:
			return
		}
	}
}

func (t *TodoItem) StopTimer() {
	if t.Running {
		t.Running = false
		close(t.Done)
		t.Done = nil
	}
}

func (t *TodoItem) ResetTimer() {
	t.StopTimer()
	t.RemainingTime, _ = time.ParseDuration(t.Duration.Text)
	t.Timer.SetText(formatTime(t.RemainingTime))
}

func formatTime(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

func playSound() {
	cmd := exec.Command("aplay", "sounds/finished.wav")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

var todoList []*TodoItem

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

func clearDoneTasks(a fyne.App, w fyne.Window) {
	var remainingTasks []*TodoItem
	for _, item := range todoList {
		if !item.Checkbox.Checked {
			remainingTasks = append(remainingTasks, item)
		}
	}
	todoList = remainingTasks
	w.SetContent(makeGUI(a, w))
}

func buildTodoList(items []*TodoItem) []fyne.CanvasObject {
	todos := make([]fyne.CanvasObject, len(items))
	for i, item := range items {
		startButton := widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func(item *TodoItem) func() {
			return func() {
				item.StartTimer()
			}
		}(item))
		stopButton := widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func(item *TodoItem) func() {
			return func() {
				item.StopTimer()
			}
		}(item))
		resetButton := widget.NewButtonWithIcon("Reset", theme.ViewRefreshIcon(), func(item *TodoItem) func() {
			return func() {
				item.ResetTimer()
			}
		}(item))

		todos[i] = container.NewHBox(
			item.Checkbox,
			item.Task,
			item.Duration,
			item.Timer,
			startButton,
			stopButton,
			resetButton,
		)
	}
	return todos
}

func makeTodoListContainer() fyne.CanvasObject {
	if len(todoList) > 0 {
		return container.NewVBox(buildTodoList(todoList)...)
	}
	return widget.NewLabel("No tasks available")
}

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

func makeLogo() fyne.CanvasObject {
	logo := canvas.NewImageFromResource(resourcePNGGODOLogoPng)
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(100, 50))
	return logo
}
