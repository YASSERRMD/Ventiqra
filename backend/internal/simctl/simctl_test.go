package simctl

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsValidModeAndSpeed(t *testing.T) {
	if !IsValidMode(ModePaused) || !IsValidMode(ModeAuto) {
		t.Error("valid modes rejected")
	}
	if IsValidMode("fast") {
		t.Error("invalid mode accepted")
	}
	if !IsValidSpeed(Speed1x) || !IsValidSpeed(Speed5x) || !IsValidSpeed(Speed30x) {
		t.Error("valid speeds rejected")
	}
	if IsValidSpeed(7) {
		t.Error("invalid speed accepted")
	}
}

func TestSpeedInterval(t *testing.T) {
	if Speed1x.Interval() != time.Second {
		t.Errorf("1x interval = %v", Speed1x.Interval())
	}
	if Speed5x.Interval() != time.Second/5 {
		t.Errorf("5x interval = %v", Speed5x.Interval())
	}
	if Speed30x.Interval() != time.Second/30 {
		t.Errorf("30x interval = %v", Speed30x.Interval())
	}
}

func TestRunnerPausedByDefault(t *testing.T) {
	r := NewRunner(func(context.Context) error { return nil })
	if c := r.Control(); c.Mode != ModePaused || c.Speed != Speed1x {
		t.Errorf("default control = %+v", c)
	}
}

func TestRunnerAutoTicksUntilPaused(t *testing.T) {
	var ticks int32
	r := NewRunner(func(context.Context) error {
		atomic.AddInt32(&ticks, 1)
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r.Apply(ctx, ModeAuto, Speed30x)
	time.Sleep(120 * time.Millisecond)
	r.Apply(ctx, ModePaused, 0)
	count := atomic.LoadInt32(&ticks)
	time.Sleep(120 * time.Millisecond)
	afterPause := atomic.LoadInt32(&ticks)

	if count < 1 {
		t.Fatalf("expected ticks while auto, got %d", count)
	}
	if afterPause != count {
		t.Errorf("ticks continued after pause: was %d, now %d", count, afterPause)
	}
}

func TestRunnerStopsOnError(t *testing.T) {
	var ticks int32
	r := NewRunner(func(context.Context) error {
		n := atomic.AddInt32(&ticks, 1)
		if n >= 2 {
			return errBoom
		}
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r.Apply(ctx, ModeAuto, Speed30x)
	time.Sleep(200 * time.Millisecond)
	final := atomic.LoadInt32(&ticks)
	if final > 3 {
		t.Errorf("runner did not stop on error: ticks=%d", final)
	}
}

func TestRunnerChangesSpeed(t *testing.T) {
	r := NewRunner(func(context.Context) error { return nil })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r.Apply(ctx, ModeAuto, Speed1x)
	if c := r.Control(); c.Speed != Speed1x {
		t.Errorf("speed = %v, want 1x", c.Speed)
	}
	r.Apply(ctx, Mode(""), Speed30x)
	if c := r.Control(); c.Speed != Speed30x {
		t.Errorf("speed = %v, want 30x", c.Speed)
	}
	r.Shutdown()
}

var errBoom = boom{}

type boom struct{}

func (boom) Error() string { return "boom" }
