/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const versionLatest = "latest"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gm",
	Short: "Go manager",
	Long: `Go version manager.
Helps to install and use multiple versions of Go at the same time.

To install latest version of Go toolchain and use it
as default run the following set of commands:
	gm install latest
	gm use latest
`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
