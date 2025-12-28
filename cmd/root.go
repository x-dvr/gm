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

const versionLatest = "latest"

var showVersion bool

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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if showVersion {
			info, ok := debug.ReadBuildInfo()
			if !ok {
				fmt.Println("Build info not available")
				os.Exit(0)
			}
			fmt.Println("gm - Go version manager")
			fmt.Println("Version:", info.Main.Version)
			fmt.Println("Built using:", info.GoVersion)
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Print version information")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
