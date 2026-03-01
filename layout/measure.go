package layout

import (
	"strings"
	"unicode/utf8"
)

// MeasureTextFunc is a function that measures text width given text, fontSize, bold, italic
type MeasureTextFunc func(text string, fontSize float64, bold bool, italic bool) float64

// TextMeasurer is the function used to measure text width.
// Set this to use accurate font measurements (e.g., from Fyne).
// If nil, falls back to estimation.
var TextMeasurer MeasureTextFunc

// MeasureText returns the width of text.
// Uses TextMeasurer if set, otherwise estimates.
func MeasureText(text string, fontSize float64) float64 {
	if TextMeasurer != nil {
		return TextMeasurer(text, fontSize, false, false)
	}
	// Fallback: rough estimation
	if len(text) == 0 {
		return 0
	}
	avgCharWidth := fontSize * 0.5
	return float64(len(text)) * avgCharWidth
}

// MeasureTextWithSpacing returns text width including CSS letter-spacing.
func MeasureTextWithSpacing(text string, fontSize, letterSpacing float64) float64 {
	width := MeasureText(text, fontSize)
	if letterSpacing == 0 {
		return width
	}
	runeCount := utf8.RuneCountInString(text)
	if runeCount <= 1 {
		return width
	}
	width += letterSpacing * float64(runeCount-1)
	if width < 0 {
		return 0
	}
	return width
}

// MeasureTextWithWordSpacing returns text width including CSS word-spacing.
// Word spacing applies only between word boundaries (space/tab runs).
func MeasureTextWithWordSpacing(text string, fontSize, wordSpacing float64) float64 {
	width := MeasureText(text, fontSize)
	if wordSpacing == 0 {
		return width
	}
	wordGaps := countWordGaps(text)
	if wordGaps == 0 {
		return width
	}
	width += wordSpacing * float64(wordGaps)
	if width < 0 {
		return 0
	}
	return width
}

// WrapText breaks text into lines that fit within maxWidth.
// Returns slice of lines. Words are not broken mid-word.
func WrapText(text string, fontSize float64, maxWidth float64) []string {
	return WrapTextWithSpacing(text, fontSize, maxWidth, 0, 0)
}

// WrapTextWithSpacing breaks text into lines that fit maxWidth using letter-spacing and word-spacing.
func WrapTextWithSpacing(text string, fontSize, maxWidth, letterSpacing, wordSpacing float64) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	startsWithSpace := strings.HasPrefix(text, " ") || strings.HasPrefix(text, "\t")
	endsWithSpace := strings.HasSuffix(text, " ") || strings.HasSuffix(text, "\t")

	words := strings.Fields(text)
	if len(words) == 0 {
		// Whitespace-only text nodes should still render a space.
		if strings.TrimSpace(text) == "" && strings.ContainsAny(text, " \t") {
			return []string{" "}
		}
		return []string{}
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// Try adding word to current line
		testLine := currentLine.String()
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		lineWidth := MeasureTextWithSpacingAndWordSpacing(testLine, fontSize, letterSpacing, wordSpacing)

		if lineWidth <= maxWidth || currentLine.Len() == 0 {
			// Word fits, or it's the first word (must include even if too long)
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		} else {
			// Word doesn't fit, start new line
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}

	// Don't forget the last line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	if len(lines) > 0 {
		if startsWithSpace {
			lines[0] = " " + lines[0]
		}
		if endsWithSpace {
			lines[len(lines)-1] = lines[len(lines)-1] + " "
		}
	}

	return lines
}

func MeasureTextWithSpacingAndWordSpacing(text string, fontSize, letterSpacing, wordSpacing float64) float64 {
	width := MeasureTextWithSpacing(text, fontSize, letterSpacing)
	if wordSpacing == 0 {
		return width
	}
	width += wordSpacing * float64(countWordGaps(text))
	if width < 0 {
		return 0
	}
	return width
}

func countWordGaps(text string) int {
	if text == "" {
		return 0
	}
	inSpace := false
	wordGaps := 0
	for _, r := range text {
		switch r {
		case ' ', '\t':
			if !inSpace {
				inSpace = true
				wordGaps++
			}
		default:
			inSpace = false
		}
	}
	return wordGaps
}
