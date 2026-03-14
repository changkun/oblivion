// Command oblivion is the CLI entrypoint.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/wallfacer/oblivion/internal/app"
)

// internalFlagsEnabled gates flags used only for integration testing.
// These flags are registered in the binary but excluded from -help output.
const internalFlagsEnabled = true

// pauseDuration is how long the -pause flag causes run to wait.
// It is a variable so tests can shorten it without waiting 5 seconds.
var pauseDuration = 5 * time.Second

// run parses args, executes the application, and returns an exit code.
// stdout and stderr are injectable so the function can be tested without
// spawning a subprocess.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("oblivion", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "oblivion – a task-running tool.")
		fmt.Fprintln(fs.Output(), "Usage: oblivion [flags]")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "Flags:")
		fs.VisitAll(func(f *flag.Flag) {
			if f.Usage == "" {
				return // internal flag, hidden from -help
			}
			fmt.Fprintf(fs.Output(), "  -%s\n    \t%s\n", f.Name, f.Usage)
		})
	}
	var verbose bool
	fs.BoolVar(&verbose, "verbose", false, "enable verbose output")
	fs.BoolVar(&verbose, "v", false, "enable verbose output (shorthand)")

	var pause bool
	if internalFlagsEnabled {
		// -pause is an internal flag for integration tests; empty usage hides it
		// from the custom Usage printer above.
		fs.BoolVar(&pause, "pause", false, "")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	if args := fs.Args(); len(args) > 0 {
		fmt.Fprintf(stderr, "error: unexpected arguments: %v\nRun 'oblivion -help' for usage.\n", args)
		return 1
	}

	if pause {
		if ctx.Err() == nil {
			fmt.Fprintln(stderr, "ready")
		}
		select {
		case <-ctx.Done():
		case <-time.After(pauseDuration):
		}
	}

	if err := app.Run(ctx, app.Config{Verbose: verbose, Output: stdout, Log: stderr}); err != nil {
		if ctx.Err() != nil {
			// Context was cancelled by SIGINT/SIGTERM — suppress the message and
			// exit 130 (128+SIGINT) per POSIX signal exit-code convention.
			return 130
		}
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

// watchSignal waits for either a signal on sigCh or ctx cancellation. When a
// signal arrives, it is stored in val and cancel is called. Extracted to allow
// unit testing of the signal-capture logic without spawning subprocesses.
func watchSignal(ctx context.Context, cancel context.CancelFunc, sigCh <-chan os.Signal, val *atomic.Value) {
	select {
	case sig := <-sigCh:
		val.Store(sig)
		cancel()
	case <-ctx.Done():
	}
}

// computeExitCode returns the POSIX exit code for the process. If stored
// contains a syscall.Signal (from watchSignal), the code is 128+signal;
// otherwise runCode (the value returned by run) is used unchanged.
func computeExitCode(runCode int, stored interface{}) int {
	if sig, ok := stored.(syscall.Signal); ok {
		return 128 + int(sig)
	}
	return runCode
}

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	var sigVal atomic.Value
	go watchSignal(ctx, cancel, sigCh, &sigVal)

	code := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	cancel()
	signal.Stop(sigCh)

	os.Exit(computeExitCode(code, sigVal.Load()))
}
