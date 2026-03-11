package cmd

import (
	"fmt"
	"strings"

	"github.com/alfred/alfred/internal/store"
	"github.com/spf13/cobra"
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Manage categories",
}

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		cats, err := store.LoadCategories()
		if err != nil {
			return err
		}
		if len(cats) == 0 {
			fmt.Println("No categories defined.")
			return nil
		}
		for _, c := range cats {
			fmt.Println(" •", c)
		}
		return nil
	},
}

var categoryAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new category",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(strings.Join(args, " "))
		if name == "" {
			return fmt.Errorf("category name cannot be empty")
		}

		cats, err := store.LoadCategories()
		if err != nil {
			return err
		}

		// Deduplicate (case-insensitive)
		for _, c := range cats {
			if strings.EqualFold(c, name) {
				fmt.Printf("Category %q already exists.\n", c)
				return nil
			}
		}

		cats = append(cats, name)
		if err := store.SaveCategories(cats); err != nil {
			return err
		}
		fmt.Printf("Added category %q.\n", name)
		return nil
	},
}

var categoryDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cats, err := store.LoadCategories()
		if err != nil {
			return err
		}

		updated := cats[:0]
		found := false
		for _, c := range cats {
			if strings.EqualFold(c, name) {
				found = true
				continue
			}
			updated = append(updated, c)
		}

		if !found {
			return fmt.Errorf("category %q not found", name)
		}

		if err := store.SaveCategories(updated); err != nil {
			return err
		}
		fmt.Printf("Deleted category %q.\n", name)
		return nil
	},
}

func init() {
	categoryCmd.AddCommand(categoryListCmd)
	categoryCmd.AddCommand(categoryAddCmd)
	categoryCmd.AddCommand(categoryDeleteCmd)
	rootCmd.AddCommand(categoryCmd)
}
