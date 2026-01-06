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
	"github.com/x-dvr/gm/ui/pbar"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Args:    cobra.MaximumNArgs(1),
	Short:   "Install specified version of Go toolchain",
	Long:    fmt.Sprintf("Use %q to install most recent version of toolchain.", versionLatest),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		version := ""
		if len(args) == 1 {
			version = args[0]
		}
		if version == versionLatest || version == "" {
			version, err = toolchain.GetLatestVersion()
			if err != nil {
				printError("Failed to get latest Go version: %s", err)
				os.Exit(1)
			}
		}
		if !strings.HasPrefix(version, "go") {
			version = "go" + version
		}

		destPath, err := sys.PathForVersion(version)
		if err != nil {
			printError("Failed to determine	destination path for installation: %s", err)
			os.Exit(1)
		}

		unprefixed := strings.TrimPrefix(version, "go")
		tui := pbar.New(fmt.Sprintf("Installing Go %s", unprefixed))

		go func() {
			err = toolchain.Install(version, destPath, tui.GetTracker())
			if err != nil {
				tui.Exit(fmt.Errorf("install toolchain (ver. %s) into path %q: %w", unprefixed, destPath, err))
				return
			}

			if err := sys.SetAsCurrent(version); err != nil {
				tui.Exit(fmt.Errorf("set installed toolchain version %q as current: %w", unprefixed, err))
				return
			}

			tui.Exit(nil)
		}()

		if err := tui.Run(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
