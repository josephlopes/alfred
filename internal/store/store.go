package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/alfred/alfred/internal/model"
)

// filePath returns the path to the JSON storage file.
// It respects XDG_CONFIG_HOME if set, otherwise falls back to ~/.config/alfred/tasks.json.
func filePath() (string, error) {
	configDir, err := os.UserConfigDir() // returns ~/.config on Linux/Mac
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "alfred", "tasks.json"), nil
}

// Load reads all tasks from disk. Returns an empty slice if the file doesn't exist yet.
func Load() ([]model.Task, error) {
	path, err := filePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []model.Task{}, nil
	}
	if err != nil {
		return nil, err
	}

	var tasks []model.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// Save writes the given tasks slice to disk, creating the directory if needed.
func Save(tasks []model.Task) error {
	path, err := filePath()
	if err != nil {
		return err
	}

	// Ensure the directory exists before writing
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// NextID returns max(existing IDs) + 1, so IDs are always monotonically increasing.
func NextID(tasks []model.Task) int {
	max := 0
	for _, t := range tasks {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}
