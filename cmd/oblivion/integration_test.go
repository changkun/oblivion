//go:build integration

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "oblivion-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(dir, "oblivion")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestBinary(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		wantCode           int
		wantStdout         string
		wantStderr         string
		wantStderrContains string
	}{
		{
			name:       "no args",
			args:       nil,
			wantCode:   0,
			wantStdout: "oblivion\n",
			wantStderr: "",
		},
		{
			name:               "verbose flag",
			args:               []string{"-verbose"},
			wantCode:           0,
			wantStdout:         "oblivion\n",
			wantStderrContains: "verbose:",
		},
		{
			name:               "unknown flag",
			args:               []string{"-nope"},
			wantCode:           1,
			wantStdout:         "",
			wantStderrContains: "flag provided but not defined",
		},
		{
			name:               "help flag",
			args:               []string{"-help"},
			wantCode:           0,
			wantStdout:         "",
			wantStderrContains: "oblivion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			code := 0
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					code = exitErr.ExitCode()
				} else {
					t.Fatalf("unexpected error running binary: %v", err)
				}
			}

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d (stdout: %q, stderr: %q)",
					code, tt.wantCode, stdout.String(), stderr.String())
			}

			if tt.wantStdout != "" {
				if stdout.String() != tt.wantStdout {
					t.Errorf("stdout = %q, want %q", stdout.String(), tt.wantStdout)
				}
			} else if stdout.Len() != 0 {
				t.Errorf("expected empty stdout, got %q", stdout.String())
			}

			if tt.wantStderrContains != "" {
				if !strings.Contains(stderr.String(), tt.wantStderrContains) {
					t.Errorf("stderr = %q, want it to contain %q", stderr.String(), tt.wantStderrContains)
				}
			} else if tt.wantStderr == "" && stderr.Len() != 0 {
				t.Errorf("expected empty stderr, got %q", stderr.String())
			}
		})
	}
}
