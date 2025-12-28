/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	// set via -ldflags
	version = "dev"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Aliases: []string{"up"},
	Args:    cobra.ExactArgs(0),
	Short:   "Upgrade self",
	Long:    "Upgrade gm to latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("determine path of executable: %w", err)
		}
		fmt.Printf("upgrade called %s\n", exePath)

		info, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Println("Build info not available")
			return nil
		}

		fmt.Println("LD version:", version)
		fmt.Println("Go version:", info.GoVersion)
		fmt.Println("Module path:", info.Path)
		fmt.Println("Main version:", info.Main.Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
