package main

import (
	"fmt"
	"fyne.io/fyne/v2/widget"
	"github.com/bradhe/stopwatch"
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
