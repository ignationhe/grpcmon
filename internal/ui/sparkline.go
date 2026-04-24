package ui

import (
	"strings"
)

// sparkChars are the unicode block characters used to render sparklines.
var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline renders a compact single-line bar chart from a slice of float64
// values. width controls the number of characters in the output string.
// Values are normalised against the maximum in the provided slice.
type Sparkline struct {
	Width int
}

// NewSparkline returns a Sparkline with the given display width.
func NewSparkline(width int) *Sparkline {
	if width <= 0 {
		width = 10
	}
	return &Sparkline{Width: width}
}

// Render converts values into a sparkline string.
// If values is empty or all zeros a flat line is returned.
func (s *Sparkline) Render(values []float64) string {
	if len(values) == 0 {
		return strings.Repeat(string(sparkChars[0]), s.Width)
	}

	// Take the last Width samples.
	samples := values
	if len(samples) > s.Width {
		samples = samples[len(samples)-s.Width:]
	}

	max := 0.0
	for _, v := range samples {
		if v > max {
			max = v
		}
	}

	var b strings.Builder
	for _, v := range samples {
		idx := 0
		if max > 0 {
			idx = int((v / max) * float64(len(sparkChars)-1))
		}
		b.WriteRune(sparkChars[idx])
	}

	// Pad with lowest bar if fewer samples than width.
	for b.Len() < s.Width {
		b.WriteRune(sparkChars[0])
	}

	return b.String()
}
