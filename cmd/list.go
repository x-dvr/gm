/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/x-dvr/gm/sys"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(0),
	Short:   "List all installed versions of Go toolchain",
	Run: func(cmd *cobra.Command, args []string) {
		installed, err := sys.ListInstalledVersions()
		if err != nil {
			printError("Failed to list installed versions: %s", err)
			os.Exit(1)
		}

		fmt.Println(sTitleBar.Render(sTitle.Render("Installed versions of Go")))
		if len(installed) == 0 {
			fmt.Println(sPadLeft.Render(sInfo.Render("No Go versions found")))
			return
		}

		current, err := sys.GetCurrentVersion()
		if err != nil {
			printError("Failed to determine current version: %s", err)
			os.Exit(1)
		}

		items := make([]string, 0, len(installed))

		for _, toolchain := range installed {
			if toolchain.Version == current.Version {
				text := sActiveText.Render(toolchain.Version + " - current")
				sub := sSubtext.Render(toolchain.Path)
				items = append(items, sActiveListItem.Render(text+"\n"+sub))
			} else {
				text := sText.Render(toolchain.Version)
				sub := sSubtext.Render(toolchain.Path)
				items = append(items, sListItem.Render(text+"\n"+sub))
			}
		}

		fmt.Println(sPadLeft.Render(lipgloss.JoinVertical(lipgloss.Left, items...)))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
