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
			fmt.Println(sInfo.Render("\nNo updates available"))
			os.Exit(0)
		}

		fmt.Println(sText.Padding(0, 2).Render("\nUpdate available:"), sActiveText.Render(latest.Version))
		asset, err := latest.FindAsset(runtime.GOOS, runtime.GOARCH)
		if err != nil {
			if errors.Is(err, upgrade.ErrPlatformNotSupported) {
				fmt.Println(sError.Render(fmt.Sprintf("Platform %s %s is not supported", runtime.GOOS, runtime.GOARCH)))
				return
			}
			printError("Failed to find update files for this system: %s", err)
			os.Exit(1)
		}
		fmt.Println(sText.Padding(0, 2).Render("Downloading", asset.URL))
		downloadPath, err := asset.Download()
		if err != nil {
			printError("Failed to download update: %s", err)
			os.Exit(1)
		}
		if err := os.Rename(exePath, exePath+".bak"); err != nil {
			printError("Failed to backup current executable: %s", err)
			os.Exit(1)
		}

		if err = upgrade.Extract(downloadPath, installPath); err != nil {
			printError("Failed to extract: %s", err)
			os.Exit(1)
		}
		fmt.Println(sText.Padding(0, 2).Render("Updated"), sActiveText.Render(exePath))

		if err = os.Remove(downloadPath); err != nil {
			printError("Failed to cleanup: %s", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
