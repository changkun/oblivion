// Package app contains the core application logic for oblivion.
package app

import (
	"errors"
	"fmt"
	"io"
)

// Config holds the runtime configuration for the application.
type Config struct {
	// Verbose enables detailed logging when true.
	Verbose bool
	// Output is the writer used for all program output.
	Output io.Writer
}

// Run is the main entrypoint for the application.
// It executes the program with the provided configuration and returns any
// error encountered.
func Run(cfg Config) error {
	if cfg.Output == nil {
		return errors.New("config.Output must not be nil")
	}

	if _, err := fmt.Fprintln(cfg.Output, "oblivion"); err != nil {
		return err
	}

	if cfg.Verbose {
		_, err := fmt.Fprintf(cfg.Output, "verbose: output written to %T\n", cfg.Output)
		return err
	}

	return nil
}
