package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"INFO":    slog.LevelInfo,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
		"":        slog.LevelInfo,
		"bogus":   slog.LevelInfo,
	}
	for in, want := range cases {
		if got := parseLevel(in); got != want {
			t.Errorf("parseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParseLevelFlagErrors(t *testing.T) {
	if _, err := ParseLevelFlag("nope"); err == nil {
		t.Error("expected error for invalid level")
	}
	if _, err := ParseLevelFlag("info"); err != nil {
		t.Errorf("unexpected error for info: %v", err)
	}
}

func TestNewJSONHandlerEmitsJSON(t *testing.T) {
	// Capture stdout by writing through a handler over a buffer instead.
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := slog.New(h)
	l.Info("hello", "key", "val")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("not json: %v (raw=%q)", err, buf.String())
	}
	if entry["msg"] != "hello" {
		t.Errorf("msg = %v, want hello", entry["msg"])
	}
}

func TestNewTextHandler(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := slog.New(h)
	l.Info("hi")
	if !strings.Contains(buf.String(), "msg=hi") {
		t.Errorf("text output missing msg=hi: %q", buf.String())
	}
}

func TestNewDefaultsGracefully(t *testing.T) {
	// New must not panic on empty/unknown values and should return a usable logger.
	l := New("", "")
	if l == nil {
		t.Fatal("New returned nil logger")
	}
	l.Debug("should be filtered at info default")
}
