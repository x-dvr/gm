/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package pbar

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/x-dvr/gm/progress"
)

type tui struct {
	program *tea.Program
	tracker *progress.Tracker
}

func New(title string) tui {
	m := newModel(title)
	p := tea.NewProgram(m)
	t := progress.NewTracker(func(ratio float64) {
		p.Send(ProgressMsg(ratio))
	})

	return tui{program: p, tracker: t}
}

func (t tui) Run() error {
	model, err := t.program.Run()
	if err != nil {
		return err
	}
	if m, ok := model.(Model); ok && m.GetError() != nil {
		return m.GetError()
	}
	return nil
}

func (t tui) SetError(err error) {
	t.program.Send(ErrMsg(err))
}

func (t tui) SetInfo(info string) {
	t.program.Send(InfoMsg(info))
	t.tracker.Reset()
}

func (t tui) Exit(err error) {
	t.SetError(err)
	time.Sleep(500 * time.Millisecond)
	t.program.Send(tea.Quit())
}

func (t tui) GetTracker() progress.IOTracker {
	return t.tracker
}
