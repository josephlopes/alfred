package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command when called without any subcommands.
// When no subcommand is given, it launches the interactive TUI.
var rootCmd = &cobra.Command{
	Use:   "alfred",
	Short: "A simple to-do CLI",
	Long:  "Manage your to-do list from the terminal. Run without arguments to open the interactive UI.",
	// RunE is called when no subcommand is provided
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}

// Execute is the entry point called by main.go.
// cobra handles routing to the correct subcommand automatically.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
