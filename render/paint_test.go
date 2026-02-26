package render

import (
	"browser/css"
	"browser/dom"
	"browser/layout"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetListInfo(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		targetIndex   int // which li's text node to test (0-based)
		wantIsItem    bool
		wantIsOrdered bool
		wantOrdinal   int
		wantListType  string
	}{
		{
			name:          "basic ol",
			html:          "<ol><li>A</li><li>B</li><li>C</li></ol>",
			targetIndex:   1,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   2,
			wantListType:  "1",
		},
		{
			name:          "basic ul",
			html:          "<ul><li>A</li><li>B</li></ul>",
			targetIndex:   0,
			wantIsItem:    true,
			wantIsOrdered: false,
			wantOrdinal:   1,
			wantListType:  "1",
		},
		{
			name:          "ol with start",
			html:          `<ol start="5"><li>A</li><li>B</li><li>C</li></ol>`,
			targetIndex:   2,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   7,
			wantListType:  "1",
		},
		{
			name:          "ol reversed",
			html:          `<ol reversed><li>A</li><li>B</li><li>C</li></ol>`,
			targetIndex:   0,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   3,
			wantListType:  "1",
		},
		{
			name:          "ol reversed second item",
			html:          `<ol reversed><li>A</li><li>B</li><li>C</li></ol>`,
			targetIndex:   1,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   2,
			wantListType:  "1",
		},
		{
			name:          "li with value attribute",
			html:          `<ol><li>A</li><li value="5">B</li><li>C</li></ol>`,
			targetIndex:   1,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   5,
			wantListType:  "1",
		},
		{
			name:          "li value continues after",
			html:          `<ol><li>A</li><li value="5">B</li><li>C</li></ol>`,
			targetIndex:   2,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   6,
			wantListType:  "1",
		},
		{
			name:          "li value with reversed",
			html:          `<ol reversed start="10"><li>A</li><li value="5">B</li><li>C</li></ol>`,
			targetIndex:   0,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   10,
			wantListType:  "1",
		},
		{
			name:          "li value with reversed second",
			html:          `<ol reversed start="10"><li>A</li><li value="5">B</li><li>C</li></ol>`,
			targetIndex:   1,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   5,
			wantListType:  "1",
		},
		{
			name:          "li value with reversed third",
			html:          `<ol reversed start="10"><li>A</li><li value="5">B</li><li>C</li></ol>`,
			targetIndex:   2,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   4,
			wantListType:  "1",
		},
		{
			name:          "negative value",
			html:          `<ol><li value="-2">A</li><li>B</li><li>C</li></ol>`,
			targetIndex:   0,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   -2,
			wantListType:  "1",
		},
		{
			name:          "negative value continues",
			html:          `<ol><li value="-2">A</li><li>B</li><li>C</li></ol>`,
			targetIndex:   2,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   0,
			wantListType:  "1",
		},
		{
			name:          "ol type a",
			html:          `<ol type="a"><li>A</li><li>B</li></ol>`,
			targetIndex:   1,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   2,
			wantListType:  "a",
		},
		{
			name:          "menu element",
			html:          `<menu><li>Copy</li><li>Paste</li></menu>`,
			targetIndex:   0,
			wantIsItem:    true,
			wantIsOrdered: false,
			wantOrdinal:   1,
			wantListType:  "1",
		},
		{
			name:          "value jumps mid-list",
			html:          `<ol><li>A</li><li>B</li><li value="10">C</li><li>D</li><li value="5">E</li><li>F</li></ol>`,
			targetIndex:   5,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   6,
			wantListType:  "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse HTML and build layout tree
			doc := dom.Parse(strings.NewReader(tt.html))
			viewport := layout.Viewport{Width: 800, Height: 600}
			layoutRoot := layout.BuildLayoutTree(doc, css.Stylesheet{}, viewport, css.MatchContext{})
			layout.ComputeLayout(layoutRoot, 800)

			// Find the target li's text box
			textBox := findTextBoxByIndex(layoutRoot, tt.targetIndex)
			if textBox == nil {
				t.Fatalf("could not find text box at index %d", tt.targetIndex)
			}

			isItem, isOrdered, ordinal, listType := getListInfo(textBox)

			assert.Equal(t, tt.wantIsItem, isItem, "isItem")
			assert.Equal(t, tt.wantIsOrdered, isOrdered, "isOrdered")
			assert.Equal(t, tt.wantOrdinal, ordinal, "ordinal")
			assert.Equal(t, tt.wantListType, listType, "listType")
		})
	}
}

// findTextBoxByIndex finds the nth TextBox in the layout tree (0-indexed)
func findTextBoxByIndex(box *layout.LayoutBox, index int) *layout.LayoutBox {
	count := 0
	return findTextBoxHelper(box, index, &count)
}

func findTextBoxHelper(box *layout.LayoutBox, target int, count *int) *layout.LayoutBox {
	if box.Type == layout.TextBox && box.Text != "" {
		if *count == target {
			return box
		}
		*count++
	}
	for _, child := range box.Children {
		if result := findTextBoxHelper(child, target, count); result != nil {
			return result
		}
	}
	return nil
}

func TestFormatListMarker(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		listType string
		expected string
	}{
		{"decimal 1", 1, "1", "1."},
		{"decimal 10", 10, "1", "10."},
		{"lowercase a", 1, "a", "a."},
		{"lowercase c", 3, "a", "c."},
		{"uppercase A", 1, "A", "A."},
		{"uppercase C", 3, "A", "C."},
		{"roman lowercase i", 1, "i", "i."},
		{"roman lowercase iv", 4, "i", "iv."},
		{"roman uppercase I", 1, "I", "I."},
		{"roman uppercase IV", 4, "I", "IV."},
		{"roman uppercase X", 10, "I", "X."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatListMarker(tt.index, tt.listType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToRomanUpper(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "I"},
		{2, "II"},
		{3, "III"},
		{4, "IV"},
		{5, "V"},
		{6, "VI"},
		{9, "IX"},
		{10, "X"},
		{14, "XIV"},
		{19, "XIX"},
		{40, "XL"},
		{50, "L"},
		{90, "XC"},
		{100, "C"},
		{400, "CD"},
		{500, "D"},
		{900, "CM"},
		{1000, "M"},
		{1994, "MCMXCIV"},
		{2024, "MMXXIV"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := toRomanUpper(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildDisplayListCarriesLetterSpacing(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		cssText  string
		expected float64
	}{
		{
			name:     "inline letter-spacing applied",
			html:     `<p style="letter-spacing: 4px;">AB</p>`,
			expected: 4,
		},
		{
			name:     "inherited letter-spacing applied",
			html:     `<div style="letter-spacing: 3px;"><span>AB</span></div>`,
			expected: 3,
		},
		{
			name:     "stylesheet letter-spacing applied",
			html:     `<p class="wide">AB</p>`,
			cssText:  `.wide { letter-spacing: 2px; }`,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := dom.Parse(strings.NewReader(tt.html))
			sheet := css.Parse(tt.cssText)
			viewport := layout.Viewport{Width: 800, Height: 600}
			layoutRoot := layout.BuildLayoutTree(doc, sheet, viewport, css.MatchContext{})
			layout.ComputeLayout(layoutRoot, 800)

			commands := BuildDisplayList(layoutRoot, InputState{}, LinkStyler{})
			found := false
			for _, cmd := range commands {
				drawText, ok := cmd.(DrawText)
				if !ok {
					continue
				}
				if drawText.Text == "AB" {
					assert.Equal(t, tt.expected, drawText.LetterSpacing)
					found = true
				}
			}
			assert.True(t, found, "expected to find DrawText command for AB")
		})
	}
}
