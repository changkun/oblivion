// Package app contains the core application logic for oblivion.
package app

import (
	"context"
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
	// Log is the writer for diagnostic output (verbose mode). Defaults to io.Discard when nil.
	Log io.Writer
}

// Run is the main entrypoint for the application.
// It executes the program with the provided configuration and returns any
// error encountered.
func Run(ctx context.Context, cfg Config) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if cfg.Output == nil {
		return errors.New("config.Output must not be nil")
	}

	n, err := fmt.Fprintln(cfg.Output, "oblivion")
	if err != nil {
		return err
	}

	if cfg.Verbose {
		log := cfg.Log
		if log == nil {
			log = io.Discard
		}
		_, err = fmt.Fprintf(log, "verbose: wrote %d bytes\n", n)
		return err
	}

	return nil
}
