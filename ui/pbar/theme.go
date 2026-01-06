/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package pbar

import (
	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/lipgloss"
)

type theme struct {
	empty   string
	from    string
	to      string
	percent lipgloss.Style
}

var (
	renderer = lipgloss.DefaultRenderer()
	latte    = catppuccin.Latte
	frappe   = catppuccin.Frappe
)

func makeTheme() theme {
	t := latte
	if renderer.HasDarkBackground() {
		t = frappe
	}

	return theme{
		empty:   t.Surface2().Hex,
		from:    t.Blue().Hex,
		to:      t.Lavender().Hex,
		percent: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: latte.Text().Hex, Dark: frappe.Text().Hex}),
	}
}
