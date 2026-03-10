package main

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// blockChars maps heights 0-8 to Unicode block characters.
// Index 0 = empty (space), 8 = full block.
var blockChars = [9]string{" ", "\u2581", "\u2582", "\u2583", "\u2584", "\u2585", "\u2586", "\u2587", "\u2588"}

// brailleDotBits maps (row, col) to the bit position in a braille character.
// Braille cell is 2 cols x 4 rows. Unicode base: U+2800.
//
//	Row 0: dot1 (0x01)  dot4 (0x08)
//	Row 1: dot2 (0x02)  dot5 (0x10)
//	Row 2: dot3 (0x04)  dot6 (0x20)
//	Row 3: dot7 (0x40)  dot8 (0x80)
var brailleDotBits = [4][2]byte{
	{0x01, 0x08}, // row 0 (top)
	{0x02, 0x10}, // row 1
	{0x04, 0x20}, // row 2
	{0x40, 0x80}, // row 3 (bottom)
}

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

// RenderBrailleChart renders a filled area chart using braille characters.
// Each character cell is 2 dots wide x 4 dots tall, giving 2*width horizontal
// and 4*height vertical resolution. Returns a slice of strings, one per row
// (top to bottom), each already styled with a vertical color gradient.
func RenderBrailleChart(pdf []float64, lo, hi float64, width, height int, gradient []lipgloss.Color) []string {
	if len(pdf) == 0 || width <= 0 || height <= 0 {
		blank := strings.Repeat(" ", width)
		rows := make([]string, height)
		for i := range rows {
			rows[i] = blank
		}
		return rows
	}

	// Resample to 2*width (2 x-positions per braille cell)
	dotsW := width * 2
	values := resamplePDF(pdf, dotsW)

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

	// Total vertical dot resolution: height rows * 4 dots per row
	totalDots := height * 4

	// Precompute column heights in dot units
	colHeights := make([]int, dotsW)
	for i, v := range values {
		h := int(math.Round(v / maxVal * float64(totalDots)))
		if h < 0 {
			h = 0
		}
		if h > totalDots {
			h = totalDots
		}
		colHeights[i] = h
	}

	rows := make([]string, height)
	for row := 0; row < height; row++ {
		// Row 0 = top, row height-1 = bottom
		rowBaseDot := (height - 1 - row) * 4 // bottom dot index of this row

		var sb strings.Builder
		for col := 0; col < width; col++ {
			var braille byte
			leftH := colHeights[col*2]
			rightH := colHeights[col*2+1]

			// For each of the 4 dot rows (0=top, 3=bottom of this cell)
			for dotRow := 0; dotRow < 4; dotRow++ {
				dotY := rowBaseDot + (3 - dotRow) // map dotRow 0 (top of cell) to highest dot
				if dotY < leftH {
					braille |= brailleDotBits[dotRow][0]
				}
				if dotY < rightH {
					braille |= brailleDotBits[dotRow][1]
				}
			}

			sb.WriteRune(rune(0x2800 + int(braille)))
		}

		// Apply gradient color based on row position (top = bright, bottom = dim)
		colorIdx := 0
		if len(gradient) > 1 && height > 1 {
			// row 0 = top (highest color index), row height-1 = bottom (lowest)
			colorIdx = (height - 1 - row) * (len(gradient) - 1) / (height - 1)
		}
		if colorIdx >= len(gradient) {
			colorIdx = len(gradient) - 1
		}
		style := lipgloss.NewStyle().Foreground(gradient[colorIdx])
		rows[row] = style.Render(sb.String())
	}

	return rows
}

// RenderBrailleSparkline renders a single-row braille sparkline.
// Each character covers 2 horizontal samples and 4 vertical levels,
// giving 2x horizontal resolution compared to block-char sparklines.
func RenderBrailleSparkline(pdf []float64, lo, hi float64, width int) string {
	if len(pdf) == 0 || width <= 0 {
		return strings.Repeat(" ", width)
	}

	// 2 horizontal samples per character
	dotsW := width * 2
	values := resamplePDF(pdf, dotsW)

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
	for col := 0; col < width; col++ {
		var braille byte
		leftVal := values[col*2]
		rightVal := values[col*2+1]

		// Map to 0-4 dot height (4 rows per braille cell)
		leftH := int(math.Round(leftVal / maxVal * 4))
		rightH := int(math.Round(rightVal / maxVal * 4))
		if leftH > 4 {
			leftH = 4
		}
		if rightH > 4 {
			rightH = 4
		}

		// Fill dots from bottom up
		for dotRow := 0; dotRow < 4; dotRow++ {
			dotY := 3 - dotRow // dotRow 0 = top of cell, dotY 3 = bottom
			if dotY < leftH {
				braille |= brailleDotBits[dotRow][0]
			}
			if dotY < rightH {
				braille |= brailleDotBits[dotRow][1]
			}
		}

		sb.WriteRune(rune(0x2800 + int(braille)))
	}

	return sb.String()
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
