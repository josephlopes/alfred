package cmd

import (
	"fmt"

	"github.com/alfred/alfred/internal/store"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Styles using lipgloss
var (
	doneStyle    = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("240"))
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	idStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := store.Load()
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks yet. Add one with: alfred add <title>")
			return nil
		}

		for _, t := range tasks {
			checkbox := "[ ]"
			title := pendingStyle.Render(t.Title)
			if t.Done {
				checkbox = "[x]"
				title = doneStyle.Render(t.Title)
			}
			fmt.Printf("%s %s %s\n", idStyle.Render(fmt.Sprintf("#%d", t.ID)), checkbox, title)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
