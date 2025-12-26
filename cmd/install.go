/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Args:    cobra.ExactArgs(1),
	Short:   "Install specified version of Go toolchain",
	Long:    fmt.Sprintf("Use '%s' to install most recent version of toolchain.", versionLatest),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("install called - %s\n", args[0])
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
