/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/x-dvr/gm/sys"
	"github.com/x-dvr/gm/toolchain"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use",
	Args:  cobra.ExactArgs(1),
	Short: "Set specified version of Go toolchain as current",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		version := args[0]
		if version == versionLatest {
			version, err = toolchain.GetLatestVersion()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get latest Go version: %s", err.Error())
				os.Exit(1)
			}
		}
		if !strings.HasPrefix(version, "go") {
			version = "go" + version
		}

		if err := sys.SetAsCurrent(version); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to prepare env variables: %s", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
