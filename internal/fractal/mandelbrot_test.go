package fractal_test

import (
	"testing"

	"github.com/wallfacer/oblivion/internal/fractal"
)

func TestEscape_InsideSet(t *testing.T) {
	// Origin is inside the set: z stays at 0 forever.
	got := fractal.Escape(0+0i, 100)
	if got != 100 {
		t.Errorf("Escape(0, 100) = %d, want 100 (inside set)", got)
	}
}

func TestEscape_InsideMainCardioid(t *testing.T) {
	// (-0.5+0i) lies inside the main cardioid; should not escape.
	got := fractal.Escape(-0.5+0i, 100)
	if got != 100 {
		t.Errorf("Escape(-0.5, 100) = %d, want 100 (inside main cardioid)", got)
	}
}

func TestEscape_FarOutside(t *testing.T) {
	// (3+0i) is far outside the set; z₁ = 3, z₂ = 12 → escapes after 1 step.
	got := fractal.Escape(3+0i, 100)
	if got == 100 {
		t.Errorf("Escape(3+0i, 100) should not return maxIter (point is outside the set)")
	}
	if got > 5 {
		t.Errorf("Escape(3+0i, 100) = %d, want a small value (fast escape)", got)
	}
}

func TestEscape_NearBoundary(t *testing.T) {
	// (-2+0i) is on the boundary; it should escape eventually but not instantly.
	got := fractal.Escape(-2+0i, 1000)
	// At exactly -2 the sequence is: 0 → -2 → 2 → 2 → … which doesn't escape;
	// the point is actually on the boundary of the set.
	// This just checks the function returns a value in [0, 1000].
	if got < 0 || got > 1000 {
		t.Errorf("Escape(-2+0i, 1000) = %d, want in [0, 1000]", got)
	}
}

func TestEscape_MaxIterOne(t *testing.T) {
	// Even a point far outside should return 0 when maxIter=1
	// because the first check (before any iteration) sees z=0 which doesn't escape.
	got := fractal.Escape(100+0i, 1)
	// i=0: zr=0,zi=0 → 0 not > 4 → advance to zr=100 → loop ends → return maxIter=1
	if got != 1 {
		t.Errorf("Escape(100+0i, 1) = %d, want 1", got)
	}
}

func TestEscape_QuickEscape(t *testing.T) {
	// c = 10+10i: |z₁|² = 200 > 4, so escapes at i=0 after computing z₁.
	// Wait: at i=0, zr=0,zi=0 → zr2+zi2=0 ≤ 4 → advance: zr=10, zi=10.
	// At i=1: zr2+zi2=200 > 4 → return 1.
	got := fractal.Escape(10+10i, 100)
	if got != 1 {
		t.Errorf("Escape(10+10i, 100) = %d, want 1", got)
	}
}

func TestFrame_Dimensions(t *testing.T) {
	frame := fractal.Frame(10, 8, -2.5, 1.0, -1.1, 1.1, 32)
	if len(frame) != 8 {
		t.Fatalf("Frame row count = %d, want 8", len(frame))
	}
	for i, row := range frame {
		if len(row) != 10 {
			t.Errorf("Frame row[%d] length = %d, want 10", i, len(row))
		}
	}
}

func TestFrame_ValuesInRange(t *testing.T) {
	const maxIter = 32
	frame := fractal.Frame(20, 10, -2.5, 1.0, -1.1, 1.1, maxIter)
	for y, row := range frame {
		for x, v := range row {
			if v < 0 || v > maxIter {
				t.Errorf("frame[%d][%d] = %d, want in [0, %d]", y, x, v, maxIter)
			}
		}
	}
}

func TestFrame_SymmetryAboutRealAxis(t *testing.T) {
	// The Mandelbrot set is symmetric about the real axis: Escape(a+bi) == Escape(a-bi).
	// A symmetric window (yMin == -yMax) with an even number of rows should produce
	// a frame where row[y] == row[height-1-y].
	const (
		w       = 20
		h       = 10
		maxIter = 64
	)
	frame := fractal.Frame(w, h, -2.5, 1.0, -1.0, 1.0, maxIter)
	for y := 0; y < h/2; y++ {
		mirror := h - 1 - y
		for x := 0; x < w; x++ {
			if frame[y][x] != frame[mirror][x] {
				t.Errorf("symmetry broken: frame[%d][%d]=%d != frame[%d][%d]=%d",
					y, x, frame[y][x], mirror, x, frame[mirror][x])
			}
		}
	}
}

func TestFrame_InteriorPoint(t *testing.T) {
	// A 3×3 frame centred at (-0.5, 0) with a tiny window should produce
	// maxIter for the centre pixel (which is inside the set).
	const maxIter = 64
	frame := fractal.Frame(3, 3, -0.6, -0.4, -0.1, 0.1, maxIter)
	// Centre pixel → c ≈ -0.5+0i, should be inside the set.
	if frame[1][1] != maxIter {
		t.Errorf("centre of interior window: frame[1][1] = %d, want %d (inside set)", frame[1][1], maxIter)
	}
}

func TestDefaultView(t *testing.T) {
	xMin, xMax, yMin, yMax := fractal.DefaultView(78, 22)
	if xMin >= xMax {
		t.Errorf("DefaultView: xMin (%f) >= xMax (%f)", xMin, xMax)
	}
	if yMin >= yMax {
		t.Errorf("DefaultView: yMin (%f) >= yMax (%f)", yMin, yMax)
	}
	// The full set fits within real [-2.5, 1.0]; check the view covers that.
	if xMin > -2.5 {
		t.Errorf("DefaultView: xMin = %f, want ≤ -2.5", xMin)
	}
	if xMax < 1.0 {
		t.Errorf("DefaultView: xMax = %f, want ≥ 1.0", xMax)
	}
}
