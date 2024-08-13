package main

import (
	"database/sql"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./todos.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS todos (
		id TEXT PRIMARY KEY,
		task TEXT,
		duration TEXT,
		remaining_time TEXT,
		completed BOOLEAN
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func saveTodoItem(item *TodoItem) error {
	_, err := db.Exec(`INSERT INTO todos (id, task, duration, remaining_time, completed) VALUES (?, ?, ?, ?, ?)`,
		item.ID.String(), item.Task.Text, item.Duration.Text, item.RemainingTime.String(), item.Checkbox.Checked)
	return err
}

func getTodoItems() ([]*TodoItem, error) {
	rows, err := db.Query(`SELECT id, task, duration, remaining_time, completed FROM todos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*TodoItem
	for rows.Next() {

		var id uuid.UUID
		var task, duration, remainingTimeStr string
		var completed bool

		err = rows.Scan(&id, &task, &duration, &remainingTimeStr, &completed)
		if err != nil {
			return nil, err
		}

		remainingTime, err := time.ParseDuration(remainingTimeStr)
		if err != nil {
			return nil, err
		}

		item := &TodoItem{
			ID:            id,
			Checkbox:      widget.NewCheck("", nil),
			Task:          widget.NewLabel(task),
			Duration:      widget.NewLabel(duration),
			Timer:         widget.NewLabel(formatTime(remainingTime)),
			RemainingTime: remainingTime,
			Running:       false,
			Done:          make(chan bool),
		}
		item.Checkbox.SetChecked(completed)
		items = append(items, item)
	}
	return items, nil
}

func deleteTodoItem(item *TodoItem) error {
	_, err := db.Exec(`DELETE FROM todos WHERE id = ?`, item.ID.String())
	return err
}

func updateRemainingTime(item *TodoItem) error {
	_, err := db.Exec(`UPDATE todos SET remaining_time = ? WHERE id = ?`, item.RemainingTime.String(), item.ID.String())
	return err
}
