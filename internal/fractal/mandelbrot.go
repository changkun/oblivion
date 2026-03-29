// Package fractal implements the Mandelbrot-set iteration kernel and ASCII
// rendering helpers used by the oblivion CLI.
//
// The Mandelbrot set is the set of complex numbers c for which the sequence
//
//	z₀ = 0,  zₙ₊₁ = zₙ² + c
//
// remains bounded. Escape counts the iterations before |z| exceeds 2; Frame
// samples this function over a rectangular region of the complex plane.
package fractal

// Escape returns the number of iterations before |z| exceeds 2 (equivalently,
// |z|² > 4), starting from z=0 with the recurrence z = z² + c. If the
// sequence has not escaped after maxIter steps the point is considered inside
// the set and maxIter is returned.
//
// The inner loop is written in component form (real/imag separated) to avoid
// allocating complex128 values on every iteration.
func Escape(c complex128, maxIter int) int {
	cr, ci := real(c), imag(c)
	zr, zi := 0.0, 0.0
	for i := 0; i < maxIter; i++ {
		zr2, zi2 := zr*zr, zi*zi
		if zr2+zi2 > 4 {
			return i
		}
		// z = z² + c:  (a+bi)² = a²-b² + 2abi
		zi = 2*zr*zi + ci
		zr = zr2 - zi2 + cr
	}
	return maxIter
}

// Frame samples the Mandelbrot escape function on a width×height grid covering
// the complex-plane window [xMin, xMax] × [yMin, yMax]. It returns escape
// counts in row-major order: rows[y][x] is the escape count for the grid cell
// at column x, row y (where row 0 is the top, i.e. maximum imaginary part).
//
// Preconditions: width ≥ 2, height ≥ 2, xMin < xMax, yMin < yMax.
func Frame(width, height int, xMin, xMax, yMin, yMax float64, maxIter int) [][]int {
	rows := make([][]int, height)
	for y := 0; y < height; y++ {
		row := make([]int, width)
		// Linearly interpolate: row 0 → yMax, row height-1 → yMin.
		im := yMax - float64(y)*(yMax-yMin)/float64(height-1)
		for x := 0; x < width; x++ {
			re := xMin + float64(x)*(xMax-xMin)/float64(width-1)
			row[x] = Escape(complex(re, im), maxIter)
		}
		rows[y] = row
	}
	return rows
}

// DefaultView returns the standard Mandelbrot view coordinates that show the
// complete set for the given terminal grid dimensions. The imaginary range is
// adjusted by a factor of 0.5 to compensate for terminal characters being
// approximately twice as tall as they are wide.
func DefaultView(width, height int) (xMin, xMax, yMin, yMax float64) {
	const (
		realCenter = -0.75
		imagCenter = 0.0
		realSpan   = 3.5  // covers the full set: real ∈ [-2.5, 1.0]
		charAspect = 0.5  // terminal char width / char height ≈ 0.5
	)
	// Choose imaginary span so the world aspect ratio equals the screen
	// physical aspect ratio (accounting for non-square characters).
	imagSpan := realSpan * float64(height) / float64(width) / charAspect
	xMin = realCenter - realSpan/2
	xMax = realCenter + realSpan/2
	yMin = imagCenter - imagSpan/2
	yMax = imagCenter + imagSpan/2
	return
}
