package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

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
			name:       "positional arg rejected",
			args:       []string{"foo"},
			wantCode:   1,
			wantStderr: true,
		},
		{
			name:       "positional arg with flag rejected",
			args:       []string{"-verbose", "bar"},
			wantCode:   1,
			wantStderr: true,
		},
		{
			name:       "cancelled context exits 130 with empty stderr",
			ctx:        func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			args:       []string{},
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
