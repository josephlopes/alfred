package model

import "time"

// Task represents a single to-do item.
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`

	// Category groups tasks (e.g. "Work", "Personal"). Empty means uncategorised.
	Category string `json:"category,omitempty"`

	// DueDay is the weekday the task should be completed (e.g. "Monday").
	// Empty means no due day set.
	DueDay string `json:"due_day,omitempty"`

	// Pomodoros counts how many 25-minute pomodoro sessions completed for this task.
	Pomodoros int `json:"pomodoros,omitempty"`
}
