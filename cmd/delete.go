package cmd

import (
	"fmt"
	"strconv"

	"github.com/alfred/alfred/internal/store"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
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
				tasks = append(tasks[:i], tasks[i+1:]...)
				if err := store.Save(tasks); err != nil {
					return err
				}
				fmt.Printf("Deleted task #%d\n", id)
				return nil
			}
		}

		return fmt.Errorf("task #%d not found", id)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
