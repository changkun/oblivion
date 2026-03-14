// Command oblivion is the CLI entrypoint.
package main

import (
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
	verbose := fs.Bool("verbose", false, "enable verbose output")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if err := app.Run(app.Config{Verbose: *verbose, Output: stdout}); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
