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
	"syscall"
	"testing"
	"time"
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
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
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
		{
			name:               "positional arg rejected",
			args:               []string{"unexpected"},
			wantCode:           1,
			wantStderrContains: "unexpected arguments",
		},
		{
			name:               "positional arg with flag rejected",
			args:               []string{"-verbose", "bar"},
			wantCode:           1,
			wantStderrContains: "unexpected arguments",
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

func TestSIGTERMExits143(t *testing.T) {
	t.Run("SIGTERM exits 143", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(binaryPath, "-pause")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start binary: %v", err)
		}

		// Allow the process time to start up and enter the pause sleep.
		time.Sleep(100 * time.Millisecond)

		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			t.Fatalf("failed to send SIGTERM: %v", err)
		}

		err := cmd.Wait()
		code := 0
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				code = exitErr.ExitCode()
			} else {
				t.Fatalf("unexpected error waiting for binary: %v", err)
			}
		}

		if code != 143 {
			t.Errorf("exit code = %d, want 143 (stdout: %q, stderr: %q)",
				code, stdout.String(), stderr.String())
		}
	})
}

// TestSIGINTExits130 verifies that SIGINT causes the binary to exit with code
// 130 (128+SIGINT) per the POSIX signal exit-code convention.
// TODO: replace time.Sleep with a ready-signal handshake once main.go emits
// "ready" to stderr before entering the pause select.
func TestSIGINTExits130(t *testing.T) {
	t.Run("SIGINT exits 130", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(binaryPath, "-pause")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start binary: %v", err)
		}

		// Allow the process time to start up and enter the pause sleep.
		time.Sleep(100 * time.Millisecond)

		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			t.Fatalf("failed to send SIGINT: %v", err)
		}

		var exitErr *exec.ExitError
		if err := cmd.Wait(); errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 130 {
				t.Errorf("exit code = %d, want 130 (stdout: %q)", exitErr.ExitCode(), stdout.String())
			}
		} else if err != nil {
			t.Fatalf("unexpected error waiting for binary: %v", err)
		} else {
			t.Errorf("exit code = 0, want 130")
		}
	})
}
