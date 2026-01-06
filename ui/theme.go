package ui

import (
	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/lipgloss"
)

type Theme interface {
	// Text returns base text color
	Text() lipgloss.TerminalColor
	// Subdued returns a muted text color with reduced contrast.
	// Higher level values produce more subdued (less prominent) text.
	// Valid levels: 1-4, where 1 is slightly subdued and 4 is most subdued.
	Subdued(level byte) lipgloss.TerminalColor
	// Background returns background color
	Background() lipgloss.TerminalColor
	// Surface returns an elevated background color for UI elements.
	// Higher level values produce surfaces with more elevation (more contrast from base background).
	// Valid levels: 1-4, where 1 is slightly elevated and 4 is most elevated.
	Surface(level byte) lipgloss.TerminalColor
	// Inset returns a recessed background color for UI elements.
	// Higher level values produce more recessed backgrounds (more contrast from base background).
	// Valid levels: 1-2, where 1 is slightly recessed and 2 is most recessed.
	Inset(level byte) lipgloss.TerminalColor
	// Accent returns accent color for highlights and focus states
	Accent() lipgloss.TerminalColor
	// Success returns color for positive/successful states
	Success() lipgloss.TerminalColor
	// Warning returns color for warning states
	Warning() lipgloss.TerminalColor
	// Error returns color for error/destructive states
	Error() lipgloss.TerminalColor
	// Info returns color for informational states
	Info() lipgloss.TerminalColor
}

type Catppuccin struct{}

var _ Theme = Catppuccin{}

var (
	latte  = catppuccin.Latte
	frappe = catppuccin.Frappe
)

func (Catppuccin) Text() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: latte.Text().Hex, Dark: frappe.Text().Hex}
}

func (Catppuccin) Subdued(level byte) lipgloss.TerminalColor {
	switch level {
	case 1:
		return lipgloss.AdaptiveColor{Light: latte.Subtext1().Hex, Dark: frappe.Subtext1().Hex}
	case 2:
		return lipgloss.AdaptiveColor{Light: latte.Subtext0().Hex, Dark: frappe.Subtext0().Hex}
	case 3:
		return lipgloss.AdaptiveColor{Light: latte.Overlay2().Hex, Dark: frappe.Overlay2().Hex}
	case 4:
		return lipgloss.AdaptiveColor{Light: latte.Overlay1().Hex, Dark: frappe.Overlay1().Hex}
	default:
		return lipgloss.AdaptiveColor{Light: latte.Subtext1().Hex, Dark: frappe.Subtext1().Hex}
	}
}

func (Catppuccin) Background() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: latte.Base().Hex, Dark: frappe.Base().Hex}
}

func (Catppuccin) Surface(level byte) lipgloss.TerminalColor {
	switch level {
	case 1:
		return lipgloss.AdaptiveColor{Light: latte.Surface0().Hex, Dark: frappe.Surface0().Hex}
	case 2:
		return lipgloss.AdaptiveColor{Light: latte.Surface1().Hex, Dark: frappe.Surface1().Hex}
	case 3:
		return lipgloss.AdaptiveColor{Light: latte.Surface2().Hex, Dark: frappe.Surface2().Hex}
	case 4:
		return lipgloss.AdaptiveColor{Light: latte.Overlay0().Hex, Dark: frappe.Overlay0().Hex}
	default:
		return lipgloss.AdaptiveColor{Light: latte.Overlay0().Hex, Dark: frappe.Overlay0().Hex}
	}
}

func (Catppuccin) Inset(level byte) lipgloss.TerminalColor {
	switch level {
	case 1:
		return lipgloss.AdaptiveColor{Light: latte.Mantle().Hex, Dark: frappe.Mantle().Hex}
	case 2:
		return lipgloss.AdaptiveColor{Light: latte.Crust().Hex, Dark: frappe.Crust().Hex}
	default:
		return lipgloss.AdaptiveColor{Light: latte.Crust().Hex, Dark: frappe.Crust().Hex}
	}
}

func (Catppuccin) Accent() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: latte.Lavender().Hex, Dark: frappe.Lavender().Hex}
}

func (Catppuccin) Success() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: latte.Green().Hex, Dark: frappe.Green().Hex}
}

func (Catppuccin) Warning() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: latte.Yellow().Hex, Dark: frappe.Yellow().Hex}
}

func (Catppuccin) Error() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: latte.Red().Hex, Dark: frappe.Red().Hex}
}

func (Catppuccin) Info() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: latte.Blue().Hex, Dark: frappe.Blue().Hex}
}
