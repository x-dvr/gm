/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package pbar

import (
	tp "github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/x-dvr/gm/ui"
)

type Model struct {
	title    string
	info     string
	progress tp.Model
	err      error
}

func newModel(title string) Model {
	theme := makeTheme()
	m := Model{
		title:    title,
		progress: tp.New(tp.WithGradient(theme.from, theme.to)),
	}
	m.progress.EmptyColor = theme.empty
	m.progress.PercentageStyle = theme.percent

	return m
}

func (m Model) Init() tea.Cmd {
	return m.progress.Init()
}

type (
	ProgressMsg float64
	ErrMsg      error
	InfoMsg     string
)

func (m *Model) GetError() error {
	return m.err
}

const (
	padding  = 2
	maxWidth = 80
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.progress.Width = min(msg.Width-padding*2-4, maxWidth)
		return m, nil

	case ErrMsg:
		m.err = msg
		return m, tea.Quit

	case InfoMsg:
		m.info = string(msg)
		return m, nil

	case ProgressMsg:
		return m, m.progress.SetPercent(float64(msg))

	// FrameMsg is sent when the progress bar wants to animate itself
	case tp.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(tp.Model)
		return m, cmd

	default:
		return m, nil
	}
}

var (
	t       = ui.Catppuccin{}
	sTitle  = lipgloss.NewStyle().Padding(0, 0, 1).Foreground(t.Accent()).Render
	sText   = lipgloss.NewStyle().Foreground(t.Text()).Render
	sError  = lipgloss.NewStyle().Foreground(t.Error()).Render
	sLayout = lipgloss.NewStyle().Padding(1, padding).Render
)

func (m Model) View() string {
	var widgets []string
	widgets = append(widgets,
		sTitle(m.title),
		sText(m.info),
		m.progress.View(),
	)

	if m.err != nil {
		widgets = append(widgets, sError("Error: "+m.err.Error()+"\n"))
	}

	return sLayout(lipgloss.JoinVertical(lipgloss.Left, widgets...))
}
