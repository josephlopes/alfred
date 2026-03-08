package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/store"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new task",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		tasks, err := store.Load()
		if err != nil {
			return err
		}

		t := model.Task{
			ID:        store.NextID(tasks),
			Title:     title,
			Done:      false,
			CreatedAt: time.Now(),
		}
		tasks = append(tasks, t)

		if err := store.Save(tasks); err != nil {
			return err
		}

		fmt.Printf("Added task #%d: %s\n", t.ID, t.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
