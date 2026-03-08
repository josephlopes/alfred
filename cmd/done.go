package cmd

import (
	"fmt"
	"strconv"

	"github.com/alfred/alfred/internal/store"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as done",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid id: %s", args[0])
		}

		tasks, err := store.Load()
		if err != nil {
			return err
		}

		for i, t := range tasks {
			if t.ID == id {
				tasks[i].Done = true
				if err := store.Save(tasks); err != nil {
					return err
				}
				fmt.Printf("Marked task #%d as done\n", id)
				return nil
			}
		}

		return fmt.Errorf("task #%d not found", id)
	},
}

func init() {
	rootCmd.AddCommand(doneCmd)
}
