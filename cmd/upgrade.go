/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/x-dvr/gm/ui/pbar"
	"github.com/x-dvr/gm/upgrade"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Aliases: []string{"up"},
	Args:    cobra.ExactArgs(0),
	Short:   "Upgrade self",
	Long:    "Upgrade gm to latest version",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		exePath, err := os.Executable()
		if err != nil {
			printError("Failed to determine path of executable: %s", err)
			os.Exit(1)
		}
		installPath := filepath.Dir(exePath)

		latest, err := upgrade.GetUpdate(ctx)
		if err != nil {
			printError("Failed to determine latest version: %s", err)
			os.Exit(1)
		}
		if latest == nil {
			fmt.Println(sInfo.Render("No updates available"))
			os.Exit(0)
		}

		tui := pbar.New(fmt.Sprintf("Update available: %s", latest.Version))

		go func() {
			asset, err := latest.FindAsset(runtime.GOOS, runtime.GOARCH)
			if err != nil {
				if errors.Is(err, upgrade.ErrPlatformNotSupported) {
					tui.Exit(fmt.Errorf("platform %s %s is not supported", runtime.GOOS, runtime.GOARCH))
					return
				}
				tui.Exit(fmt.Errorf("find update archive: %w", err))
				return
			}

			expectedChecksum, err := latest.GetChecksum(asset.Name)
			if err != nil && !errors.Is(err, upgrade.ErrChecksumNotFound) {
				tui.Exit(fmt.Errorf("get checksum: %w", err))
				return
			}

			downloadPath, err := asset.Download(tui.GetTracker(), expectedChecksum)
			if err != nil {
				tui.Exit(fmt.Errorf("download update archive: %w", err))
				return
			}
			if err := os.Rename(exePath, exePath+".bak"); err != nil {
				tui.Exit(fmt.Errorf("backup old version: %w", err))
				return
			}
			if err = upgrade.Extract(downloadPath, installPath, tui.GetTracker()); err != nil {
				tui.Exit(fmt.Errorf("extract update archive: %w", err))
				return
			}
			if err = os.Remove(downloadPath); err != nil {
				tui.Exit(fmt.Errorf("cleanup: %w", err))
				return
			}
			tui.SetInfo("Successfully updated!")
			tui.Exit(nil)
		}()

		if err := tui.Run(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
