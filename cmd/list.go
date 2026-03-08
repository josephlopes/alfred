package cmd

import (
	"fmt"
	"strings"

	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/store"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Styles used by the list command
var (
	doneStyle    = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("240"))
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	idStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	catStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	dueDayStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
)

var (
	listCategory string
	listDueDay   string
	listDone     bool
	listPending  bool
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

		// Apply filters
		var filtered []model.Task
		for _, t := range tasks {
			if listCategory != "" && !strings.EqualFold(t.Category, listCategory) {
				continue
			}
			if listDueDay != "" && !strings.EqualFold(t.DueDay, listDueDay) {
				continue
			}
			if listDone && !t.Done {
				continue
			}
			if listPending && t.Done {
				continue
			}
			filtered = append(filtered, t)
		}

		if len(filtered) == 0 {
			fmt.Println("No tasks match the given filters.")
			return nil
		}

		for _, t := range filtered {
			checkbox := "[ ]"
			title := pendingStyle.Render(t.Title)
			if t.Done {
				checkbox = "[x]"
				title = doneStyle.Render(t.Title)
			}

			line := fmt.Sprintf("%s %s %s", idStyle.Render(fmt.Sprintf("#%d", t.ID)), checkbox, title)
			if t.Category != "" {
				line += "  " + catStyle.Render("["+t.Category+"]")
			}
			if t.DueDay != "" {
				line += "  " + dueDayStyle.Render(t.DueDay)
			}
			if t.Pomodoros > 0 {
				line += fmt.Sprintf("  🍅×%d", t.Pomodoros)
			}
			fmt.Println(line)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&listCategory, "category", "c", "", "filter by category")
	listCmd.Flags().StringVarP(&listDueDay, "due", "d", "", "filter by due weekday")
	listCmd.Flags().BoolVar(&listDone, "done", false, "show only completed tasks")
	listCmd.Flags().BoolVar(&listPending, "pending", false, "show only pending tasks")
	rootCmd.AddCommand(listCmd)
}
