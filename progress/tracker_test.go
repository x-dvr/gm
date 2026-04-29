/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package progress

import (
	"testing"
)

func TestTracker_ResetInvokesCallbackAndClearsCounters(t *testing.T) {
	var resetCalls []string
	var progressCalls []float64
	tr := NewTracker(
		func(p float64) { progressCalls = append(progressCalls, p) },
		func(s string) { resetCalls = append(resetCalls, s) },
	)

	tr.SetSize(100)
	if _, err := tr.Write(make([]byte, 50)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got, want := tr.written.Load(), int64(50); got != want {
		t.Fatalf("written before reset = %d, want %d", got, want)
	}

	tr.Reset("downloading")

	if got, want := tr.written.Load(), int64(0); got != want {
		t.Errorf("written after reset = %d, want %d", got, want)
	}
	if got, want := tr.total.Load(), int64(0); got != want {
		t.Errorf("total after reset = %d, want %d", got, want)
	}
	if len(resetCalls) != 1 || resetCalls[0] != "downloading" {
		t.Errorf("resetCalls = %v, want [downloading]", resetCalls)
	}
	_ = progressCalls
}

func TestTracker_WriteReportsProgress(t *testing.T) {
	var progressCalls []float64
	tr := NewTracker(
		func(p float64) { progressCalls = append(progressCalls, p) },
		func(string) {},
	)
	tr.SetSize(200)

	n, err := tr.Write(make([]byte, 50))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != 50 {
		t.Errorf("Write returned n=%d, want 50", n)
	}
	if _, err := tr.Write(make([]byte, 150)); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if len(progressCalls) != 2 {
		t.Fatalf("progressCalls = %v, want 2 entries", progressCalls)
	}
	if progressCalls[0] != 0.25 {
		t.Errorf("first progress = %v, want 0.25", progressCalls[0])
	}
	if progressCalls[1] != 1.0 {
		t.Errorf("second progress = %v, want 1.0", progressCalls[1])
	}
}

func TestTracker_WriteWithoutTotalDoesNotReport(t *testing.T) {
	called := 0
	tr := NewTracker(
		func(float64) { called++ },
		func(string) {},
	)

	if _, err := tr.Write(make([]byte, 32)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if called != 0 {
		t.Errorf("onProgress called %d times, want 0 when total size is unset", called)
	}
}

func TestTracker_WriterReturnsTracker(t *testing.T) {
	tr := NewTracker(func(float64) {}, func(string) {})
	tr.SetSize(10)
	if _, err := tr.Writer().Write([]byte("hello")); err != nil {
		t.Fatalf("Writer().Write: %v", err)
	}
	if got, want := tr.written.Load(), int64(5); got != want {
		t.Errorf("written = %d, want %d", got, want)
	}
}
