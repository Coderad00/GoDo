package main

import (
	"database/sql"
	"fyne.io/fyne/v2/widget"
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task TEXT,
		duration TEXT,
		remaining_time TEXT,
		remaining_time INT
		completed BOOLEAN
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func saveTodoItem(item *TodoItem) error {
	_, err := db.Exec(`INSERT INTO todos (task, duration, remaining_time, completed) VALUES (?, ?, ?, ?)`,
		item.Task.Text, item.Duration.Text, item.RemainingTime.String(), item.Checkbox.Checked)
	return err
}

func getTodoItems() ([]*TodoItem, error) {
	rows, err := db.Query(`SELECT task, duration, remaining_time, completed FROM todos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*TodoItem
	for rows.Next() {
		var task, duration, remainingTimeStr string
		var completed bool
		err = rows.Scan(&task, &duration, &remainingTimeStr, &completed)
		if err != nil {
			return nil, err
		}

		remainingTime, err := time.ParseDuration(remainingTimeStr)
		if err != nil {
			return nil, err
		}

		item := &TodoItem{
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

func deleteTodoItem(task string) error {
	_, err := db.Exec(`DELETE FROM todos WHERE task = ?`, task)
	return err
}

func updateRemainingTime(item *TodoItem) error {
	_, err := db.Exec(`UPDATE todos SET remaining_time = ? WHERE task = ?`, item.RemainingTime.String(), item.Task.Text)
	return err
}
