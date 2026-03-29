package fractal

import (
	"io"
	"strings"
)

// palette maps escape-iteration bands to ASCII characters. The palette is
// ordered from sparse (few iterations → far outside the set) to dense (many
// iterations → near the set boundary). The last character is used exclusively
// for points inside the set (escape == maxIter).
const palette = ` .:,;i+|*%&#@$`

// paletteLen is the total number of palette characters including the interior.
const paletteLen = len(palette)

// exteriorBands is the number of bands available for exterior (escaping) points.
const exteriorBands = paletteLen - 1

// ansiColors is a handpicked gradient of ANSI 256-colour codes that transition
// through blue → cyan → green when painted from the outer edge inward.
var ansiColors = [...]int{
	17, 18, 19, 20, 21,
	27, 33, 39, 45, 51,
	50, 49, 48, 47, 46,
	82, 118, 154, 190, 226,
	220, 214, 208, 202, 196,
}

const ansiReset = "\x1b[0m"

// Render writes an ASCII representation of frame to w. Each row of frame
// becomes one line of output terminated by '\n'. When color is true, ANSI
// 256-colour escape codes are prepended to each non-interior character to
// paint a blue→cyan gradient; interior points are always uncoloured.
func Render(w io.Writer, frame [][]int, maxIter int, color bool) error {
	var sb strings.Builder
	for _, row := range frame {
		sb.Reset()
		for _, v := range row {
			ch := charFor(v, maxIter)
			if color && v < maxIter {
				band := v * exteriorBands / maxIter
				colorIdx := band * (len(ansiColors) - 1) / exteriorBands
				writeColorChar(&sb, ansiColors[colorIdx], ch)
			} else {
				sb.WriteByte(ch)
			}
		}
		sb.WriteByte('\n')
		if _, err := io.WriteString(w, sb.String()); err != nil {
			return err
		}
	}
	return nil
}

// charFor maps a single escape value to its palette character.
// Points inside the set (v == maxIter) always receive the last palette char.
func charFor(v, maxIter int) byte {
	if v == maxIter {
		return palette[paletteLen-1]
	}
	// Map [0, maxIter) onto [0, exteriorBands) linearly.
	idx := v * exteriorBands / maxIter
	return palette[idx]
}

// writeColorChar appends an ANSI 256-colour painted character to sb.
func writeColorChar(sb *strings.Builder, colorCode int, ch byte) {
	// ESC[38;5;<n>m sets the foreground to 256-colour index n.
	sb.WriteString("\x1b[38;5;")
	writeInt(sb, colorCode)
	sb.WriteByte('m')
	sb.WriteByte(ch)
	sb.WriteString(ansiReset)
}

// writeInt appends the decimal representation of n to sb without allocating.
func writeInt(sb *strings.Builder, n int) {
	if n == 0 {
		sb.WriteByte('0')
		return
	}
	var buf [3]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	sb.Write(buf[pos:])
}
