/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/x-dvr/gm/sys"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Output shell commands to set environment variables",
	Long: `Example usage:
eval $(gm env)
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := sys.PrepareGoEnvs(); err != nil {
			printError("Failed to prepare env variables: %s", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}
