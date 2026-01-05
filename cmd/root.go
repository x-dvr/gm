/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/x-dvr/gm/ui"
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
as default run the following command:
	gm install latest
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if showVersion {
			info, ok := debug.ReadBuildInfo()
			if !ok {
				fmt.Println(sPadLeft.Render(sInfo.Render("Build info not available")))
				os.Exit(0)
			}
			fmt.Println(sPanel.Render(
				lipgloss.JoinVertical(lipgloss.Center,
					sActiveText.Render("gm - Go version manager"),
					sText.Render("Version:", info.Main.Version),
					sSubtext.Render("Built with:", info.GoVersion),
				),
			))
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

func printError(fstr string, args ...any) {
	out := sError.Render(fmt.Sprintf(fstr, args...))
	fmt.Fprintln(os.Stderr, out)
}

var (
	theme     = ui.Catppuccin{}
	sTitleBar = lipgloss.NewStyle().Padding(1, 0, 1, 2)
	sTitle    = lipgloss.NewStyle().
			Background(theme.Accent()).
			Foreground(theme.Background()).
			Padding(0, 1)
	sPadLeft = lipgloss.NewStyle().Padding(0, 0, 0, 1)
	sPanel   = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(theme.Text()).
			Padding(1, 3).
			Margin(1, 1)
	sListItem = lipgloss.NewStyle().
			Padding(0, 0, 0, 2).
			Margin(0, 0, 1)
	sActiveListItem = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(theme.Accent()).
			Padding(0, 0, 0, 1).
			Margin(0, 0, 1)
	sText       = lipgloss.NewStyle().Foreground(theme.Subdued(4))
	sActiveText = lipgloss.NewStyle().Foreground(theme.Accent())
	sSubtext    = lipgloss.NewStyle().Foreground(theme.Surface(2))
	sInfo       = lipgloss.NewStyle().
			Padding(0, 0, 0, 2).
			Foreground(theme.Info())
	sError = lipgloss.NewStyle().
		Padding(1, 1).
		Foreground(theme.Error())
)
