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
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("determine path of executable: %w", err)
		}
		installPath := filepath.Dir(exePath)

		latest, err := upgrade.GetUpdate(ctx)
		if err != nil {
			return fmt.Errorf("determine latest version: %w", err)
		}
		if latest == nil {
			fmt.Println("No updates available")
			return nil
		}

		fmt.Println("Update available:", latest.Version)
		asset, err := latest.FindAsset(runtime.GOOS, runtime.GOARCH)
		if err != nil {
			if errors.Is(err, upgrade.ErrPlatformNotSupported) {
				fmt.Printf("Platform %s %s is not supported\n", runtime.GOOS, runtime.GOARCH)
				return nil
			}
			return err
		}
		fmt.Println("Downloading", asset.URL)
		downloadPath, err := asset.Download()
		if err != nil {
			return err
		}
		if err := os.Rename(exePath, exePath+".bak"); err != nil {
			return fmt.Errorf("rename executable: %w", err)
		}

		if err = upgrade.Extract(downloadPath, installPath); err != nil {
			return err
		}

		fmt.Println("Updated", exePath)

		if err = os.Remove(downloadPath); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
