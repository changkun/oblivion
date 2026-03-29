// Package app contains the core application logic for oblivion.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/wallfacer/oblivion/internal/fractal"
)

// Config holds the runtime configuration for the application.
type Config struct {
	// Verbose enables detailed logging when true.
	Verbose bool
	// Output is the writer used for all program output.
	Output io.Writer
	// Log is the writer for diagnostic output (verbose mode). Defaults to io.Discard when nil.
	Log io.Writer

	// Fractal, when true, renders an ASCII Mandelbrot set instead of the
	// default greeting output.
	Fractal bool
	// FractalWidth is the number of columns in the rendered output.
	// Values ≤ 0 default to 78.
	FractalWidth int
	// FractalHeight is the number of rows in the rendered output.
	// Values ≤ 0 default to 22.
	FractalHeight int
	// FractalIter is the maximum iteration count. Higher values reveal more
	// detail near the boundary at the cost of additional computation.
	// Values ≤ 0 default to 64.
	FractalIter int
	// FractalColor enables ANSI 256-colour output when true.
	FractalColor bool
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

	if cfg.Fractal {
		return runFractal(ctx, cfg)
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

// runFractal renders the Mandelbrot set to cfg.Output and returns any write error.
func runFractal(ctx context.Context, cfg Config) error {
	w := cfg.FractalWidth
	h := cfg.FractalHeight
	iter := cfg.FractalIter
	if w <= 0 {
		w = 78
	}
	if h <= 0 {
		h = 22
	}
	if iter <= 0 {
		iter = 64
	}

	xMin, xMax, yMin, yMax := fractal.DefaultView(w, h)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	frame := fractal.Frame(w, h, xMin, xMax, yMin, yMax, iter)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return fractal.Render(cfg.Output, frame, iter, cfg.FractalColor)
}
