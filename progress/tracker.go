/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package progress

import (
	"io"
	"sync/atomic"
)

type Tracker struct {
	total      atomic.Int64
	written    atomic.Int64
	onProgress func(float64)
	onReset    func(string)
}

var _ IOTracker = (*Tracker)(nil)

func NewTracker(onProgress func(float64), onReset func(string)) *Tracker {
	return &Tracker{
		onProgress: onProgress,
		onReset:    onReset,
	}
}

func (t *Tracker) Proxy(r io.Reader) io.Reader {
	return io.TeeReader(r, t)
}

func (t *Tracker) Reset(info string) {
	t.total.Store(0)
	t.written.Store(0)
	t.onReset(info)
}

func (t *Tracker) SetSize(total int64) {
	t.total.Store(total)
}

func (t *Tracker) Write(p []byte) (int, error) {
	written := t.written.Add(int64(len(p)))
	total := t.total.Load()
	if total > 0 {
		t.onProgress(float64(written) / float64(total))
	}
	return len(p), nil
}
