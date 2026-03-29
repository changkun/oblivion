package fractal_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/wallfacer/oblivion/internal/fractal"
)

// errWriter always fails writes, used to test error propagation.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write error") }

func TestRender_LineCount(t *testing.T) {
	const (
		w       = 20
		h       = 6
		maxIter = 32
	)
	frame := fractal.Frame(w, h, -2.5, 1.0, -1.0, 1.0, maxIter)
	var sb strings.Builder
	if err := fractal.Render(&sb, frame, maxIter, false); err != nil {
		t.Fatalf("Render returned unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(sb.String(), "\n"), "\n")
	if len(lines) != h {
		t.Errorf("line count = %d, want %d", len(lines), h)
	}
}

func TestRender_LineLengthPlain(t *testing.T) {
	// Without colour, every output character is a single byte, so each line
	// should be exactly w characters long (excluding the newline).
	const (
		w       = 15
		h       = 5
		maxIter = 32
	)
	frame := fractal.Frame(w, h, -2.5, 1.0, -1.0, 1.0, maxIter)
	var sb strings.Builder
	if err := fractal.Render(&sb, frame, maxIter, false); err != nil {
		t.Fatalf("Render error: %v", err)
	}
	for i, line := range strings.Split(strings.TrimRight(sb.String(), "\n"), "\n") {
		if len(line) != w {
			t.Errorf("line %d: len = %d, want %d", i, len(line), w)
		}
	}
}

func TestRender_OnlyPaletteChars(t *testing.T) {
	// Without colour every byte should be a palette character or '\n'.
	const (
		w       = 10
		h       = 4
		maxIter = 16
	)
	frame := fractal.Frame(w, h, -2.5, 1.0, -1.0, 1.0, maxIter)
	var sb strings.Builder
	if err := fractal.Render(&sb, frame, maxIter, false); err != nil {
		t.Fatalf("Render error: %v", err)
	}
	const palette = ` .:,;i+|*%&#@$` + "\n"
	for _, ch := range sb.String() {
		if !strings.ContainsRune(palette, ch) {
			t.Errorf("unexpected character %q in plain output", ch)
		}
	}
}

func TestRender_ColorOutput(t *testing.T) {
	// With colour enabled, the output should contain ANSI escape sequences.
	frame := fractal.Frame(20, 5, -2.5, 1.0, -1.0, 1.0, 32)
	var sb strings.Builder
	if err := fractal.Render(&sb, frame, 32, true); err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(sb.String(), "\x1b[") {
		t.Error("colour output should contain ANSI escape sequences")
	}
}

func TestRender_ColorLineCount(t *testing.T) {
	// Even with colour codes the number of '\n' characters must equal h.
	const h = 5
	frame := fractal.Frame(10, h, -2.5, 1.0, -1.0, 1.0, 32)
	var sb strings.Builder
	if err := fractal.Render(&sb, frame, 32, true); err != nil {
		t.Fatalf("Render error: %v", err)
	}
	got := strings.Count(sb.String(), "\n")
	if got != h {
		t.Errorf("newline count = %d, want %d", got, h)
	}
}

func TestRender_WriteError(t *testing.T) {
	frame := fractal.Frame(5, 3, -2.5, 1.0, -1.0, 1.0, 16)
	err := fractal.Render(errWriter{}, frame, 16, false)
	if err == nil {
		t.Error("Render should return an error when the writer fails")
	}
}

func TestRender_EmptyFrame(t *testing.T) {
	// An empty frame (no rows) should write nothing and return nil.
	var sb strings.Builder
	if err := fractal.Render(&sb, nil, 32, false); err != nil {
		t.Errorf("Render(empty frame) = %v, want nil", err)
	}
	if sb.Len() != 0 {
		t.Errorf("expected empty output for empty frame, got %q", sb.String())
	}
}

func TestRender_InteriorChar(t *testing.T) {
	// A frame where every cell is at maxIter (all inside the set) should
	// produce only the last palette character (plus newlines).
	const (
		w       = 5
		h       = 3
		maxIter = 32
	)
	// Build a synthetic all-interior frame.
	frame := make([][]int, h)
	for y := range frame {
		row := make([]int, w)
		for x := range row {
			row[x] = maxIter
		}
		frame[y] = row
	}
	var sb strings.Builder
	if err := fractal.Render(&sb, frame, maxIter, false); err != nil {
		t.Fatalf("Render error: %v", err)
	}
	// Every character (except newlines) should be '$' — the last palette char.
	for _, ch := range strings.ReplaceAll(sb.String(), "\n", "") {
		if ch != '$' {
			t.Errorf("interior char should be '$', got %q", ch)
		}
	}
}

func TestRender_ExteriorChar(t *testing.T) {
	// A frame where every cell escapes immediately (v=0) should use
	// the first palette character (' ').
	const (
		w       = 5
		h       = 3
		maxIter = 32
	)
	frame := make([][]int, h)
	for y := range frame {
		row := make([]int, w)
		// All zeros: escapes on iteration 0.
		frame[y] = row
	}
	var sb strings.Builder
	if err := fractal.Render(&sb, frame, maxIter, false); err != nil {
		t.Fatalf("Render error: %v", err)
	}
	for _, ch := range strings.ReplaceAll(sb.String(), "\n", "") {
		if ch != ' ' {
			t.Errorf("fast-escape char should be ' ', got %q", ch)
		}
	}
}

// TestRender_DiscardsOK ensures Render works with io.Discard (no error).
func TestRender_DiscardsOK(t *testing.T) {
	frame := fractal.Frame(10, 4, -2.5, 1.0, -1.0, 1.0, 16)
	if err := fractal.Render(io.Discard, frame, 16, false); err != nil {
		t.Errorf("Render to io.Discard returned error: %v", err)
	}
}
