package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// blockChars maps heights 0-8 to Unicode block characters.
// Index 0 = empty (space), 8 = full block.
var blockChars = [9]string{" ", "\u2581", "\u2582", "\u2583", "\u2584", "\u2585", "\u2586", "\u2587", "\u2588"}

// RenderCompactHistogram renders a sparkline from sampled mixture PDF values.
func RenderCompactHistogram(pdf []float64, lo, hi float64, width int) string {
	if len(pdf) == 0 || width <= 0 {
		return ""
	}

	values := resamplePDF(pdf, width)

	maxVal := 0.0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal <= 0 {
		return strings.Repeat(" ", width)
	}

	var sb strings.Builder
	for _, v := range values {
		level := int(math.Round(v / maxVal * 8))
		if level < 0 {
			level = 0
		}
		if level > 8 {
			level = 8
		}
		sb.WriteString(blockChars[level])
	}

	return sb.String()
}

// RenderCompactHistogramSimple renders a simple single-normal sparkline (fallback).
func RenderCompactHistogramSimple(mean, stddev float64, width int) string {
	if stddev <= 0 || width <= 0 {
		return ""
	}

	lo := math.Max(0, mean-3*stddev)
	hi := mean + 3*stddev
	if hi <= lo {
		return blockChars[8]
	}

	step := (hi - lo) / float64(width)
	values := make([]float64, width)
	maxVal := 0.0

	for i := 0; i < width; i++ {
		x := lo + (float64(i)+0.5)*step
		values[i] = scaledNormalPDF(x, mean, stddev)
		if values[i] > maxVal {
			maxVal = values[i]
		}
	}

	if maxVal <= 0 {
		return strings.Repeat(" ", width)
	}

	var sb strings.Builder
	for _, v := range values {
		level := int(math.Round(v / maxVal * 8))
		if level < 0 {
			level = 0
		}
		if level > 8 {
			level = 8
		}
		sb.WriteString(blockChars[level])
	}

	return sb.String()
}

// RenderTallHistogram renders a multi-row vertical bar chart from sampled PDF values.
// Each column is a vertical bar using full blocks (█) with a fractional cap (▁▂▃...▇).
// Returns a slice of strings, one per row (top to bottom).
func RenderTallHistogram(pdf []float64, lo, hi float64, width, height int, color lipgloss.Color) []string {
	if len(pdf) == 0 || width <= 0 || height <= 0 {
		blank := strings.Repeat(" ", width)
		rows := make([]string, height)
		for i := range rows {
			rows[i] = blank
		}
		return rows
	}

	// Resample to target width
	values := resamplePDF(pdf, width)

	maxVal := 0.0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal <= 0 {
		blank := strings.Repeat(" ", width)
		rows := make([]string, height)
		for i := range rows {
			rows[i] = blank
		}
		return rows
	}

	// Total vertical resolution: height rows * 8 sub-levels per row
	totalLevels := height * 8
	barStyle := lipgloss.NewStyle().Foreground(color)

	rows := make([]string, height)
	for row := 0; row < height; row++ {
		var sb strings.Builder
		// Row 0 = top, row height-1 = bottom
		// For each column, determine what character to show at this row
		rowBase := (height - 1 - row) * 8 // bottom of this row in sub-levels

		for col := 0; col < width; col++ {
			level := int(math.Round(values[col] / maxVal * float64(totalLevels)))
			if level < 0 {
				level = 0
			}
			if level > totalLevels {
				level = totalLevels
			}

			// How many sub-levels fill into this row?
			fill := level - rowBase
			if fill <= 0 {
				sb.WriteString(" ")
			} else if fill >= 8 {
				sb.WriteString(blockChars[8]) // full block
			} else {
				sb.WriteString(blockChars[fill]) // partial cap
			}
		}
		rows[row] = barStyle.Render(sb.String())
	}

	return rows
}

// resamplePDF resamples a PDF slice to a target width using nearest-neighbor.
func resamplePDF(pdf []float64, width int) []float64 {
	if len(pdf) == width {
		return pdf
	}
	values := make([]float64, width)
	for i := 0; i < width; i++ {
		srcIdx := float64(i) / float64(width) * float64(len(pdf))
		idx := int(srcIdx)
		if idx >= len(pdf) {
			idx = len(pdf) - 1
		}
		values[i] = pdf[idx]
	}
	return values
}

// RenderHorizontalHistogram renders a horizontal bar chart for the four degrees
// of success with degree-specific colors.
func RenderHorizontalHistogram(labels [4]string, probs [4]float64, maxBarWidth int) string {
	if maxBarWidth <= 0 {
		maxBarWidth = 30
	}

	maxLabelLen := 0
	for _, l := range labels {
		if len(l) > maxLabelLen {
			maxLabelLen = len(l)
		}
	}

	maxProb := 0.0
	for _, p := range probs {
		if p > maxProb {
			maxProb = p
		}
	}

	var sb strings.Builder
	for i, label := range labels {
		pct := probs[i] * 100
		barLen := 0
		if maxProb > 0 {
			barLen = int(math.Round(probs[i] / maxProb * float64(maxBarWidth)))
		}
		if barLen < 0 {
			barLen = 0
		}

		barStyle := lipgloss.NewStyle().Foreground(degreeColors[i])
		bar := barStyle.Render(strings.Repeat(blockChars[8], barLen))
		padding := strings.Repeat(" ", maxLabelLen-len(label))

		pctStyle := lipgloss.NewStyle().Foreground(degreeColors[i])
		sb.WriteString(fmt.Sprintf("  %s%s %s %s\n", padding, label, pctStyle.Render(fmt.Sprintf("%5.1f%%", pct)), bar))
	}

	return sb.String()
}
