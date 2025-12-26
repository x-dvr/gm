/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"log/slog"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		version := args[0]
		if version == versionLatest {
			version, err = toolchain.GetLatestVersion()
			if err != nil {
				slog.Error("Failed to get latest Go version", slog.String("error", err.Error()))
				os.Exit(1)
			}
		}
		if !strings.HasPrefix(version, "go") {
			version = "go" + version
		}

		return sys.SetAsCurrent(version)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
