// Command oblivion is the CLI entrypoint.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/wallfacer/oblivion/internal/app"
)

// run parses args, executes the application, and returns an exit code.
// stdout and stderr are injectable so the function can be tested without
// spawning a subprocess.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("oblivion", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "oblivion – a task-running tool.")
		fmt.Fprintln(fs.Output(), "Usage: oblivion [flags]")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "Flags:")
		fs.PrintDefaults()
	}
	var verbose bool
	fs.BoolVar(&verbose, "verbose", false, "enable verbose output")
	fs.BoolVar(&verbose, "v", false, "enable verbose output (shorthand)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	if err := app.Run(app.Config{Verbose: verbose, Output: stdout}); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
