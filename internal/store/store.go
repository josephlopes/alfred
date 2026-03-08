package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/alfred/alfred/internal/model"
)

// configDir returns the alfred config directory path.
func configDir() (string, error) {
	base, err := os.UserConfigDir() // ~/.config on Linux/macOS
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "alfred"), nil
}

// tasksPath returns the path to the JSON tasks file.
func tasksPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tasks.json"), nil
}

// categoriesPath returns the path to the JSON categories file.
func categoriesPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "categories.json"), nil
}

// ensureDir creates the alfred config directory if it does not exist.
func ensureDir() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}

// Load reads all tasks from disk. Returns an empty slice if the file doesn't exist yet.
func Load() ([]model.Task, error) {
	path, err := tasksPath()
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

// Save writes the given tasks slice to disk atomically (temp file + rename).
func Save(tasks []model.Task) error {
	if err := ensureDir(); err != nil {
		return err
	}

	path, err := tasksPath()
	if err != nil {
		return err
	}

	return writeJSON(path, tasks)
}

// LoadCategories reads the saved category list. Returns defaults if the file doesn't exist.
func LoadCategories() ([]string, error) {
	path, err := categoriesPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []string{"Personal", "Work", "Project"}, nil
	}
	if err != nil {
		return nil, err
	}

	var cats []string
	if err := json.Unmarshal(data, &cats); err != nil {
		return nil, err
	}
	return cats, nil
}

// SaveCategories writes the category list to disk.
func SaveCategories(cats []string) error {
	if err := ensureDir(); err != nil {
		return err
	}

	path, err := categoriesPath()
	if err != nil {
		return err
	}

	return writeJSON(path, cats)
}

// writeJSON marshals v and writes it to path using an atomic temp-file swap.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	// Write to a temp file in the same directory, then rename for atomicity.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
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
