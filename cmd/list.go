/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/x-dvr/gm/sys"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(0),
	Short:   "List all installed versions of Go toolchain",
	Run: func(cmd *cobra.Command, args []string) {
		installed, err := sys.ListInstalledVersions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed list installed versions: %s", err.Error())
			os.Exit(1)
		}

		if len(installed) == 0 {
			fmt.Println("No Go versions installed")
			return
		}

		current, err := sys.GetCurrentVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to determine current version: %s", err.Error())
			os.Exit(1)
		}

		for _, version := range installed {
			if version == current {
				fmt.Printf("✓ %s (current)\n", version)
			} else {
				fmt.Printf("  %s\n", version)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
