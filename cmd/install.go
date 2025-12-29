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

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Args:    cobra.MaximumNArgs(1),
	Short:   "Install specified version of Go toolchain",
	Long:    fmt.Sprintf("Use '%s' to install most recent version of toolchain.", versionLatest),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		version := ""
		if len(args) == 1 {
			version = args[0]
		}
		if version == versionLatest || version == "" {
			version, err = toolchain.GetLatestVersion()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get latest Go version: %s", err.Error())
				os.Exit(1)
			}
		}
		if !strings.HasPrefix(version, "go") {
			version = "go" + version
		}

		destPath, err := sys.PathForVersion(version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to determine	destination path for installation: %s", err.Error())
			os.Exit(1)
		}

		err = toolchain.Install(version, destPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to download toolchain (ver. %s) into path %q: %s", version, destPath, err.Error())
			os.Exit(1)
		}

		if err := sys.SetAsCurrent(version); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set installed toolchain version %q as current: %s", version, err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
