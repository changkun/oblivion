package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// errWriter is an io.Writer that always returns an error, used to force app
// error paths in tests.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write error") }

func TestRun(t *testing.T) {
	tests := []struct {
		name               string
		ctx                context.Context
		args               []string
		wantCode           int
		wantStdout         bool   // true = expect non-empty
		wantStderr         bool   // true = expect non-empty
		wantStdoutExact    string // if non-empty, asserts exact stdout content
		wantStderrContains string // if non-empty, asserts stderr contains this string
	}{
		{
			name:       "no args exits 0 and writes to stdout",
			args:       []string{},
			wantCode:   0,
			wantStdout: true,
			wantStderr: false,
		},
		{
			name:       "unknown flag exits non-zero",
			args:       []string{"-unknown"},
			wantCode:   1,
			wantStdout: false,
			wantStderr: true,
		},
		{
			name:               "verbose flag exits 0",
			args:               []string{"-verbose"},
			wantCode:           0,
			wantStdout:         true,
			wantStderr:         true,
			wantStdoutExact:    "oblivion\n",
			wantStderrContains: "verbose:",
		},
		{
			name:       "-help exits 0 and writes to stderr",
			args:       []string{"-help"},
			wantCode:   0,
			wantStdout: false,
			wantStderr: true,
		},
		{
			name:               "positional arg rejected",
			args:               []string{"foo"},
			wantCode:           1,
			wantStderr:         true,
			wantStderrContains: "unexpected arguments",
		},
		{
			name:               "positional arg with flag rejected",
			args:               []string{"-verbose", "bar"},
			wantCode:           1,
			wantStderr:         true,
			wantStderrContains: "unexpected arguments",
		},
		{
			name:       "cancelled context exits 130 with empty stderr",
			ctx:        func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			args:       []string{},
			wantCode:   130,
			wantStdout: false,
			wantStderr: false,
		},
		{
			name:       "pause flag with cancelled context exits 130",
			ctx:        func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			args:       []string{"-pause"},
			wantCode:   130,
			wantStdout: false,
			wantStderr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.ctx
			if ctx == nil {
				ctx = context.Background()
			}
			var stdout, stderr bytes.Buffer
			code := run(ctx, tt.args, &stdout, &stderr)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d (stderr: %q)", code, tt.wantCode, stderr.String())
			}
			if tt.wantStdout && stdout.Len() == 0 {
				t.Error("expected non-empty stdout, got empty")
			}
			if !tt.wantStdout && stdout.Len() != 0 {
				t.Errorf("expected empty stdout, got %q", stdout.String())
			}
			if tt.wantStderr && stderr.Len() == 0 {
				t.Error("expected non-empty stderr, got empty")
			}
			if !tt.wantStderr && stderr.Len() != 0 {
				t.Errorf("expected empty stderr, got %q", stderr.String())
			}
			if tt.wantStdoutExact != "" && stdout.String() != tt.wantStdoutExact {
				t.Errorf("stdout = %q, want %q", stdout.String(), tt.wantStdoutExact)
			}
			if tt.wantStderrContains != "" && !strings.Contains(stderr.String(), tt.wantStderrContains) {
				t.Errorf("stderr %q does not contain %q", stderr.String(), tt.wantStderrContains)
			}
		})
	}
}

// TestRunPauseTimeout verifies that when -pause is supplied with a background
// context, run exits 0 once the pause duration elapses. The duration is
// shortened to 1 ms so the test completes quickly.
func TestRunPauseTimeout(t *testing.T) {
	orig := pauseDuration
	pauseDuration = time.Millisecond
	defer func() { pauseDuration = orig }()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"-pause"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit code = %d, want 0 (stderr: %q)", code, stderr.String())
	}
}

// TestRunAppError verifies that a non-context error from app.Run causes run to
// write "error:" to stderr and return exit code 1.
func TestRunAppError(t *testing.T) {
	var stderr bytes.Buffer
	// errWriter as stdout makes app.Run fail with a non-context write error.
	code := run(context.Background(), []string{}, errWriter{}, &stderr)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "error:") {
		t.Errorf("stderr %q should contain 'error:'", stderr.String())
	}
}

func TestWatchSignal(t *testing.T) {
	t.Run("stores signal and cancels context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		sigCh <- syscall.SIGTERM

		var val atomic.Value
		watchSignal(ctx, cancel, sigCh, &val)

		if ctx.Err() == nil {
			t.Error("expected context to be cancelled after signal")
		}
		stored, ok := val.Load().(syscall.Signal)
		if !ok {
			t.Fatalf("expected stored value to be syscall.Signal, got %T", val.Load())
		}
		if stored != syscall.SIGTERM {
			t.Errorf("stored signal = %v, want SIGTERM", stored)
		}
	})

	t.Run("exits cleanly on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		sigCh := make(chan os.Signal, 1)
		var val atomic.Value

		cancel() // cancel before calling watchSignal
		watchSignal(ctx, cancel, sigCh, &val)

		if val.Load() != nil {
			t.Errorf("expected no stored signal, got %v", val.Load())
		}
	})
}

func TestComputeExitCode(t *testing.T) {
	tests := []struct {
		name     string
		runCode  int
		stored   interface{}
		wantCode int
	}{
		{
			name:     "no signal returns runCode unchanged",
			runCode:  0,
			stored:   nil,
			wantCode: 0,
		},
		{
			name:     "non-zero runCode with no signal",
			runCode:  1,
			stored:   nil,
			wantCode: 1,
		},
		{
			name:     "SIGTERM overrides runCode with 143",
			runCode:  130,
			stored:   syscall.SIGTERM,
			wantCode: 143,
		},
		{
			name:     "SIGINT overrides runCode with 130",
			runCode:  1,
			stored:   syscall.Signal(syscall.SIGINT),
			wantCode: 130,
		},
		{
			name:     "non-signal stored value returns runCode",
			runCode:  42,
			stored:   "not a signal",
			wantCode: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeExitCode(tt.runCode, tt.stored)
			if got != tt.wantCode {
				t.Errorf("computeExitCode(%d, %v) = %d, want %d", tt.runCode, tt.stored, got, tt.wantCode)
			}
		})
	}
}
