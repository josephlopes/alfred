package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/store"
	"github.com/spf13/cobra"
)

var validDays = map[string]bool{
	"monday": true, "tuesday": true, "wednesday": true,
	"thursday": true, "friday": true, "saturday": true, "sunday": true,
}

var (
	addCategory string
	addDueDay   string
)

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new task",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")

		// Validate due day if provided
		dueDay := ""
		if addDueDay != "" {
			normalized := strings.ToLower(strings.TrimSpace(addDueDay))
			if !validDays[normalized] {
				return fmt.Errorf("invalid due day %q — use Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, or Sunday", addDueDay)
			}
			// Capitalise first letter for consistent storage
			dueDay = strings.ToUpper(normalized[:1]) + normalized[1:]
		}

		tasks, err := store.Load()
		if err != nil {
			return err
		}

		t := model.Task{
			ID:        store.NextID(tasks),
			Title:     title,
			Done:      false,
			CreatedAt: time.Now(),
			Category:  strings.TrimSpace(addCategory),
			DueDay:    dueDay,
		}
		tasks = append(tasks, t)

		if err := store.Save(tasks); err != nil {
			return err
		}

		msg := fmt.Sprintf("Added task #%d: %s", t.ID, t.Title)
		if t.Category != "" {
			msg += fmt.Sprintf(" [%s]", t.Category)
		}
		if t.DueDay != "" {
			msg += fmt.Sprintf(" due %s", t.DueDay)
		}
		fmt.Println(msg)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addCategory, "category", "c", "", "category for the task (e.g. Work, Personal)")
	addCmd.Flags().StringVarP(&addDueDay, "due", "d", "", "due weekday (Monday–Sunday)")
	rootCmd.AddCommand(addCmd)
}
