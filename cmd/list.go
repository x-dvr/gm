/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/x-dvr/gm/sys"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(0),
	Short:   "List all installed versions of Go toolchain",
	RunE: func(cmd *cobra.Command, args []string) error {
		installed, err := sys.ListInstalledVersions()
		if err != nil {
			return err
		}

		if len(installed) == 0 {
			fmt.Println("No Go versions installed")
			return nil
		}

		current, err := sys.GetCurrentVersion()
		if err != nil {
			return err
		}

		for _, version := range installed {
			if version == current {
				fmt.Printf("✓ %s (current)\n", version)
			} else {
				fmt.Printf("  %s\n", version)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
