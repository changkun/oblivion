package main

import (
	"bytes"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCode    int
		wantStdout  bool // true = expect non-empty
		wantStderr  bool // true = expect non-empty
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
			name:       "verbose flag exits 0",
			args:       []string{"-verbose"},
			wantCode:   0,
			wantStdout: true,
			wantStderr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr)

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
		})
	}
}
