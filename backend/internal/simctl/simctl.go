// Package simctl implements simulation speed control: the run mode (paused or
// auto), the speed (1/5/30 ticks per real second), and an auto-runner that
// advances the simulation by invoking a caller-provided tick function on a
// ticker. The runner is per-company and safe for concurrent control changes.
package simctl

import (
	"context"
	"sync"
	"time"
)

// Mode is the run state of a company's simulation.
type Mode string

const (
	ModePaused Mode = "paused"
	ModeAuto   Mode = "auto"
)

// Speed is the number of simulated ticks per real second.
type Speed int

const (
	Speed1x  Speed = 1
	Speed5x  Speed = 5
	Speed30x Speed = 30
)

// IsValidMode reports whether m is a recognized mode.
func IsValidMode(m Mode) bool { return m == ModePaused || m == ModeAuto }

// IsValidSpeed reports whether s is a recognized speed.
func IsValidSpeed(s Speed) bool { return s == Speed1x || s == Speed5x || s == Speed30x }

// Interval returns the real-time duration between ticks for a speed.
func (s Speed) Interval() time.Duration {
	switch s {
	case Speed30x:
		return time.Second / 30
	case Speed5x:
		return time.Second / 5
	default:
		return time.Second
	}
}

// Control is the desired run state for a company.
type Control struct {
	Mode  Mode
	Speed Speed
}

// TickFunc advances one simulation day. The runner calls it on each tick.
type TickFunc func(ctx context.Context) error

// Runner auto-advances a company's simulation when mode=auto. It is safe for
// concurrent Apply calls and Start/Stop.
type Runner struct {
	mu     sync.Mutex
	ctrl   Control
	tick   TickFunc
	cancel context.CancelFunc
}

// NewRunner returns a Runner in paused mode at 1x.
func NewRunner(tick TickFunc) *Runner {
	return &Runner{ctrl: Control{Mode: ModePaused, Speed: Speed1x}, tick: tick}
}

// Control returns a snapshot of the current run state.
func (r *Runner) Control() Control {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ctrl
}

// Apply sets the mode and/or speed, (re)starting or stopping the auto-loop as
// needed. Passing zero values leaves the field unchanged.
func (r *Runner) Apply(ctx context.Context, mode Mode, speed Speed) {
	r.mu.Lock()
	if mode != "" && IsValidMode(mode) {
		r.ctrl.Mode = mode
	}
	if speed != 0 && IsValidSpeed(speed) {
		r.ctrl.Speed = speed
	}
	ctrl := r.ctrl
	r.mu.Unlock()

	if ctrl.Mode == ModeAuto {
		r.start(ctx, ctrl.Speed)
	} else {
		r.stop()
	}
}

// start (re)launches the auto-loop at the given speed, replacing any prior loop.
func (r *Runner) start(parent context.Context, speed Speed) {
	r.stop()
	ctx, cancel := context.WithCancel(parent)
	r.mu.Lock()
	r.cancel = cancel
	tick := r.tick
	r.mu.Unlock()

	interval := speed.Interval()
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := tick(ctx); err != nil {
					// On error (e.g. bankrupt), stop the loop.
					return
				}
			}
		}
	}()
}

// stop cancels the current auto-loop if any.
func (r *Runner) stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
}

// Shutdown stops the runner. Safe to call multiple times.
func (r *Runner) Shutdown() {
	r.stop()
}
