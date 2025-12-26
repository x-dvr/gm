/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return sys.SetGoEnvs()
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}
