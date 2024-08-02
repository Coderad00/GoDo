//go:generate fyne bundle -o bundled.go assets

package main

import (
	"database/sql"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bradhe/stopwatch"
	_ "github.com/mattn/go-sqlite3"
	"image/color"
	"log"
	"os/exec"
	"time"
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

var todoList []*TodoItem

func main() {
	initDB()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println("db-error", err)
			return
		}
	}(db)

	a := app.NewWithID("GoDo")
	w := a.NewWindow("GoDo")
	w.SetIcon(resourceLogoWindowmanagerWhitePng)

	todoList, _ = getTodoItems()

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("GoDo", fyne.NewMenuItem("show", func() { w.Show() }))
		desk.SetSystemTrayMenu(m)
		desk.SetSystemTrayIcon(resourceLogoWindowmanagerWhitePng)
	}

	w.SetContent(makeGUI(a, w))
	w.Resize(fyne.NewSize(400, 200))

	w.SetCloseIntercept(func() {
		w.Close()
	})

	w.ShowAndRun()
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
		saveTodoItem(newItem)
		w.SetContent(makeGUI(a, w))
		inputWindow.Close()
	}

	inputContainer := container.NewVBox(taskEntry, durationSelect, widget.NewButton("Save", saveCallback))
	inputWindow.SetContent(inputContainer)
	inputWindow.Show()
}

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

func (item *TodoItem) StartTimer() {
	if item.Running {
		println("TIMER ALREADY STARTED")
		return
	}

	if item.Done == nil {
		item.Done = make(chan bool)
	}

	if item.RemainingTime <= 0 {
		item.RemainingTime, _ = time.ParseDuration(item.Duration.Text)
	}

	item.Running = true
	go item.runTimer()
}

func (item *TodoItem) runTimer() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			item.RemainingTime -= time.Second
			if item.RemainingTime <= 0 {
				item.StopTimer()
				playSound()
				return
			}
			item.Timer.SetText(formatTime(item.RemainingTime))
		case <-item.Done:
			return
		}
	}
}

func (item *TodoItem) StopTimer() {
	updateRemainingTime(item)
	if item.Running {
		item.Running = false
		close(item.Done)
		item.Done = nil
	}
}

func (item *TodoItem) ResetTimer() {
	item.StopTimer()
	item.RemainingTime, _ = time.ParseDuration(item.Duration.Text)
	item.Timer.SetText(formatTime(item.RemainingTime))
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

func clearDoneTasks(a fyne.App, w fyne.Window) {
	var remainingTasks []*TodoItem
	for _, item := range todoList {
		if !item.Checkbox.Checked {
			remainingTasks = append(remainingTasks, item)
		} else {
			deleteTodoItem(item.Task.Text)
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
