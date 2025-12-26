/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/x-dvr/gm/sys"
	"github.com/x-dvr/gm/toolchain"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Args:    cobra.ExactArgs(1),
	Short:   "Install specified version of Go toolchain",
	Long:    fmt.Sprintf("Use '%s' to install most recent version of toolchain.", versionLatest),
	Run: func(cmd *cobra.Command, args []string) {
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

		destPath, err := sys.PathForVersion(version)
		if err != nil {
			slog.Error("Failed to determine	destination path for installation", slog.String("error", err.Error()))
			os.Exit(1)
		}

		err = toolchain.Install(version, destPath)
		if err != nil {
			slog.Error("Failed to download toolchain", slog.String("error", err.Error()), slog.String("version", version), slog.String("path", destPath))
			os.Exit(1)
		}

		if err := sys.SetAsCurrent(version); err != nil {
			slog.Error("Failed to set installed toolchain as current", slog.String("error", err.Error()), slog.String("version", version))
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
