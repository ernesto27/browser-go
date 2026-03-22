package render

import (
	"browser/css"
	"browser/dom"
	"browser/layout"
	"image/color"
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
			name:          "display list-item div in ul",
			html:          `<ul><div style="display: list-item;">A</div><div style="display: list-item;">B</div></ul>`,
			targetIndex:   1,
			wantIsItem:    true,
			wantIsOrdered: false,
			wantOrdinal:   2,
			wantListType:  "1",
		},
		{
			name:          "display list-item in ol with upper-roman",
			html:          `<ol style="list-style-type: upper-roman;"><div style="display: list-item;">A</div><div style="display: list-item;">B</div></ol>`,
			targetIndex:   1,
			wantIsItem:    true,
			wantIsOrdered: true,
			wantOrdinal:   2,
			wantListType:  "I",
		},
		{
			name:          "display list-item with square parent",
			html:          `<ul style="list-style-type: square;"><div style="display: list-item;">X</div></ul>`,
			targetIndex:   0,
			wantIsItem:    true,
			wantIsOrdered: false,
			wantOrdinal:   1,
			wantListType:  "square",
		},
		{
			name:          "standalone display list-item no list parent",
			html:          `<div style="display: list-item;">Solo</div>`,
			targetIndex:   0,
			wantIsItem:    true,
			wantIsOrdered: false,
			wantOrdinal:   1,
			wantListType:  "disc",
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
		{"decimal 1", 1, css.ListMarkerNumeric, "1."},
		{"decimal 10", 10, css.ListMarkerNumeric, "10."},
		{"lowercase a", 1, css.ListMarkerLowerAlpha, "a."},
		{"lowercase c", 3, css.ListMarkerLowerAlpha, "c."},
		{"uppercase A", 1, css.ListMarkerUpperAlpha, "A."},
		{"uppercase C", 3, css.ListMarkerUpperAlpha, "C."},
		{"roman lowercase i", 1, css.ListMarkerLowerRoman, "i."},
		{"roman lowercase iv", 4, css.ListMarkerLowerRoman, "iv."},
		{"roman uppercase I", 1, css.ListMarkerUpperRoman, "I."},
		{"roman uppercase IV", 4, css.ListMarkerUpperRoman, "IV."},
		{"roman uppercase X", 10, css.ListMarkerUpperRoman, "X."},
		{"disc", 1, css.ListStyleDisc, "•"},
		{"circle", 1, css.ListStyleCircle, "◦"},
		{"square", 1, css.ListStyleSquare, "■"},
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

func TestBuildDisplayListCarriesWordSpacing(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		cssText  string
		expected float64
	}{
		{
			name:     "inline word-spacing applied",
			html:     `<p style="word-spacing: 5px;">A B</p>`,
			expected: 5,
		},
		{
			name:     "inherited word-spacing applied",
			html:     `<div style="word-spacing: 3px;"><span>A B</span></div>`,
			expected: 3,
		},
		{
			name:     "stylesheet word-spacing applied",
			html:     `<p class="wide">A B</p>`,
			cssText:  `.wide { word-spacing: 2px; }`,
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
				if drawText.Text == "A B" {
					assert.Equal(t, tt.expected, drawText.WordSpacing)
					found = true
				}
			}
			assert.True(t, found, "expected to find DrawText command for A B")
		})
	}
}

func TestBuildDisplayListJustifyWordSpacing(t *testing.T) {
	// Use a narrow container to force wrapping, so justify spacing kicks in
	html := `<p style="text-align: justify;">word word word word word word word word word word word word end</p>`
	doc := dom.Parse(strings.NewReader(html))
	sheet := css.Parse("")
	viewport := layout.Viewport{Width: 200, Height: 600}
	layoutRoot := layout.BuildLayoutTree(doc, sheet, viewport, css.MatchContext{})
	layout.ComputeLayout(layoutRoot, 200)

	commands := BuildDisplayList(layoutRoot, InputState{}, LinkStyler{})

	// Collect all DrawText commands
	var textCmds []DrawText
	for _, cmd := range commands {
		if dt, ok := cmd.(DrawText); ok {
			textCmds = append(textCmds, dt)
		}
	}

	if len(textCmds) > 1 {
		// Non-last lines should have positive WordSpacing from justification
		for i := 0; i < len(textCmds)-1; i++ {
			assert.Greater(t, textCmds[i].WordSpacing, 0.0,
				"non-last DrawText line %d should have positive WordSpacing from justify", i)
		}
		// Last line should have zero WordSpacing (not justified)
		assert.Equal(t, 0.0, textCmds[len(textCmds)-1].WordSpacing,
			"last DrawText line should have zero WordSpacing")
	}
}

// helper: find all DrawRect commands matching a color
func findRectsByColor(commands []DisplayCommand, c color.Color) []DrawRect {
	var rects []DrawRect
	for _, cmd := range commands {
		if dr, ok := cmd.(DrawRect); ok && dr.Color == c {
			rects = append(rects, dr)
		}
	}
	return rects
}

// helper: build layout tree and compute layout from HTML+CSS
func buildLayout(html, cssText string, width float64) *layout.LayoutBox {
	doc := dom.Parse(strings.NewReader(html))
	sheet := css.Parse(cssText)
	viewport := layout.Viewport{Width: width, Height: 600}
	root := layout.BuildLayoutTree(doc, sheet, viewport, css.MatchContext{})
	layout.ComputeLayout(root, width)
	return root
}

func TestHorizontalScrollbar(t *testing.T) {
	tests := []struct {
		name            string
		html            string
		css             string
		expectTrack     bool
		expectThumbFull bool // thumb fills entire track (no overflow)
	}{
		{
			name:            "scroll always shows scrollbar",
			html:            `<div class="box">short</div>`,
			css:             `.box { width: 200px; overflow-x: scroll; }`,
			expectTrack:     true,
			expectThumbFull: true,
		},
		{
			name:            "auto with overflow shows scrollbar",
			html:            `<div class="box"><div style="width:500px;">wide</div></div>`,
			css:             `.box { width: 200px; overflow-x: auto; }`,
			expectTrack:     true,
			expectThumbFull: false,
		},
		{
			name:        "auto without overflow no scrollbar",
			html:        `<div class="box">short</div>`,
			css:         `.box { width: 200px; overflow-x: auto; }`,
			expectTrack: false,
		},
		{
			name:        "hidden no scrollbar",
			html:        `<div class="box"><div style="width:500px;">wide</div></div>`,
			css:         `.box { width: 200px; overflow-x: hidden; }`,
			expectTrack: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := buildLayout(tt.html, tt.css, 800)
			commands := BuildDisplayList(root, InputState{}, LinkStyler{})

			tracks := findRectsByColor(commands, ColorScrollbarTrack)
			thumbs := findRectsByColor(commands, ColorScrollbarThumb)

			if tt.expectTrack {
				assert.NotEmpty(t, tracks, "expected scrollbar track")
				assert.NotEmpty(t, thumbs, "expected scrollbar thumb")

				track := tracks[0]
				thumb := thumbs[0]
				assert.Equal(t, ScrollbarHeight, track.Rect.Height, "track height")
				assert.True(t, thumb.Rect.Width > 0, "thumb should have positive width")

				if tt.expectThumbFull {
					expectedThumbW := track.Rect.Width - 2*scrollbarThumbPadding
					assert.InDelta(t, expectedThumbW, thumb.Rect.Width, 1, "thumb should fill track when no overflow")
				} else {
					assert.True(t, thumb.Rect.Width < track.Rect.Width, "thumb should be smaller than track when overflowing")
				}
			} else {
				assert.Empty(t, tracks, "expected no scrollbar track")
				assert.Empty(t, thumbs, "expected no scrollbar thumb")
			}
		})
	}
}

func TestHorizontalScrollbarThumbMinWidth(t *testing.T) {
	// Very wide content should clamp thumb to minimum width
	html := `<div class="box"><div style="width:10000px;">x</div></div>`
	cssText := `.box { width: 200px; overflow-x: scroll; }`
	root := buildLayout(html, cssText, 800)
	commands := BuildDisplayList(root, InputState{}, LinkStyler{})

	thumbs := findRectsByColor(commands, ColorScrollbarThumb)
	assert.NotEmpty(t, thumbs)
	assert.True(t, thumbs[0].Rect.Width >= ScrollbarThumbMinWidth, "thumb width should be at least minimum")
}

func TestHorizontalScrollbarThumbPosition(t *testing.T) {
	// When scroll offset is set, thumb should move right
	html := `<div class="box">` + strings.Repeat("word ", 100) + `</div>`
	cssText := `.box { width: 200px; overflow-x: scroll; white-space: nowrap; }`
	root := buildLayout(html, cssText, 800)

	// Find the overflow container's DOM node
	var overflowNode *dom.Node
	var findNode func(box *layout.LayoutBox)
	findNode = func(box *layout.LayoutBox) {
		if box.Style.OverflowX == "scroll" && box.Node != nil {
			overflowNode = box.Node
			return
		}
		for _, child := range box.Children {
			findNode(child)
		}
	}
	findNode(root)
	assert.NotNil(t, overflowNode, "should find overflow node")

	// No scroll: thumb at left
	cmds0 := BuildDisplayList(root, InputState{}, LinkStyler{})
	thumbs0 := findRectsByColor(cmds0, ColorScrollbarThumb)
	assert.NotEmpty(t, thumbs0)
	thumbX0 := thumbs0[0].Rect.X

	// With scroll offset: thumb moves right
	state := InputState{
		ScrollOffsets: map[*dom.Node]float64{overflowNode: 100},
	}
	cmds1 := BuildDisplayList(root, state, LinkStyler{})
	thumbs1 := findRectsByColor(cmds1, ColorScrollbarThumb)
	assert.NotEmpty(t, thumbs1)
	assert.True(t, thumbs1[0].Rect.X > thumbX0, "thumb should move right when scrolled")
}

func TestScrollOffsetShiftsContent(t *testing.T) {
	html := `<div class="box">` + strings.Repeat("content ", 50) + `</div>`
	cssText := `.box { width: 200px; overflow-x: scroll; white-space: nowrap; }`
	root := buildLayout(html, cssText, 800)

	// Find the text DrawText in both cases
	findTextCmd := func(commands []DisplayCommand, text string) *DrawText {
		for _, cmd := range commands {
			if dt, ok := cmd.(DrawText); ok && strings.Contains(dt.Text, text) {
				return &dt
			}
		}
		return nil
	}

	// No scroll
	cmds0 := BuildDisplayList(root, InputState{}, LinkStyler{})
	dt0 := findTextCmd(cmds0, "content")
	assert.NotNil(t, dt0, "should find text command without scroll")

	// Find overflow node
	var overflowNode *dom.Node
	var findNode func(box *layout.LayoutBox)
	findNode = func(box *layout.LayoutBox) {
		if box.Style.OverflowX == "scroll" && box.Node != nil {
			overflowNode = box.Node
			return
		}
		for _, child := range box.Children {
			findNode(child)
		}
	}
	findNode(root)

	// With scroll offset: text gets left-clipped (ClipLeftOffset set)
	state := InputState{
		ScrollOffsets: map[*dom.Node]float64{overflowNode: 50},
	}
	cmds1 := BuildDisplayList(root, state, LinkStyler{})
	dt1 := findTextCmd(cmds1, "content")
	assert.NotNil(t, dt1, "should find text command with scroll")
	assert.True(t, dt1.ClipLeftOffset > 0, "scrolled text should have ClipLeftOffset > 0")
}

func TestScrollClipLeft(t *testing.T) {
	html := `<div class="box">` + strings.Repeat("ABCDEFGHIJKLMNOP ", 30) + `</div>`
	cssText := `.box { width: 200px; overflow-x: scroll; white-space: nowrap; }`
	root := buildLayout(html, cssText, 800)

	var overflowNode *dom.Node
	var findNode func(box *layout.LayoutBox)
	findNode = func(box *layout.LayoutBox) {
		if box.Style.OverflowX == "scroll" && box.Node != nil {
			overflowNode = box.Node
			return
		}
		for _, child := range box.Children {
			findNode(child)
		}
	}
	findNode(root)
	assert.NotNil(t, overflowNode)

	// With large scroll offset, text should have ClipLeftOffset set
	state := InputState{
		ScrollOffsets: map[*dom.Node]float64{overflowNode: 100},
	}
	cmds := BuildDisplayList(root, state, LinkStyler{})
	for _, cmd := range cmds {
		if dt, ok := cmd.(DrawText); ok && strings.Contains(dt.Text, "ABCDEFG") {
			assert.True(t, dt.ClipLeftOffset > 0, "text scrolled past left edge should have ClipLeftOffset > 0")
			return
		}
	}
}

func TestComputeClipStart(t *testing.T) {
	tests := []struct {
		name       string
		overflow   string
		boxType    layout.BoxType
		pos        float64
		border     float64
		current    float64
		expected   float64
	}{
		{"visible no clip", "visible", layout.BlockBox, 100, 2, 0, 0},
		{"empty no clip", "", layout.BlockBox, 100, 2, 0, 0},
		{"hidden sets clip", "hidden", layout.BlockBox, 100, 2, 0, 102},
		{"scroll sets clip", "scroll", layout.BlockBox, 100, 2, 0, 102},
		{"auto sets clip", "auto", layout.BlockBox, 100, 2, 0, 102},
		{"tighter clip wins", "hidden", layout.BlockBox, 150, 2, 102, 152},
		{"wider clip keeps current", "hidden", layout.BlockBox, 50, 2, 102, 102},
		{"inline box ignored", "hidden", layout.InlineBox, 100, 2, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeClipStart(tt.overflow, tt.boxType, tt.pos, tt.border, tt.current)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaxChildRight(t *testing.T) {
	box := &layout.LayoutBox{
		Children: []*layout.LayoutBox{
			{Rect: layout.Rect{X: 10, Width: 100}},
			{Rect: layout.Rect{X: 50, Width: 200}},
			{Rect: layout.Rect{X: 0, Width: 50}},
		},
	}
	assert.Equal(t, 250.0, maxChildRight(box))
}

func TestMaxChildRightEmpty(t *testing.T) {
	box := &layout.LayoutBox{}
	assert.Equal(t, 0.0, maxChildRight(box))
}

func TestNeedsHorizontalScrollbar(t *testing.T) {
	tests := []struct {
		name     string
		box      *layout.LayoutBox
		style    TextStyle
		expected bool
	}{
		{
			name: "scroll always true",
			box: &layout.LayoutBox{
				Type:  layout.BlockBox,
				Rect:  layout.Rect{X: 0, Width: 200},
				Style: css.Style{OverflowX: "scroll"},
			},
			style:    TextStyle{OverflowX: "scroll"},
			expected: true,
		},
		{
			name: "auto with overflow",
			box: &layout.LayoutBox{
				Type:    layout.BlockBox,
				Rect:    layout.Rect{X: 0, Width: 200},
				Padding: layout.EdgeSizes{Right: 0},
				Style:   css.Style{OverflowX: "auto"},
				Children: []*layout.LayoutBox{
					{Rect: layout.Rect{X: 0, Width: 300}},
				},
			},
			style:    TextStyle{OverflowX: "auto"},
			expected: true,
		},
		{
			name: "auto without overflow",
			box: &layout.LayoutBox{
				Type:    layout.BlockBox,
				Rect:    layout.Rect{X: 0, Width: 200},
				Padding: layout.EdgeSizes{Right: 0},
				Style:   css.Style{OverflowX: "auto"},
				Children: []*layout.LayoutBox{
					{Rect: layout.Rect{X: 0, Width: 100}},
				},
			},
			style:    TextStyle{OverflowX: "auto"},
			expected: false,
		},
		{
			name: "hidden never",
			box: &layout.LayoutBox{
				Type:  layout.BlockBox,
				Rect:  layout.Rect{X: 0, Width: 200},
				Style: css.Style{OverflowX: "hidden"},
			},
			style:    TextStyle{OverflowX: "hidden"},
			expected: false,
		},
		{
			name: "inline box never",
			box: &layout.LayoutBox{
				Type:  layout.InlineBox,
				Rect:  layout.Rect{X: 0, Width: 200},
				Style: css.Style{OverflowX: "scroll"},
			},
			style:    TextStyle{OverflowX: "scroll"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, needsHorizontalScrollbar(tt.box, tt.style))
		})
	}
}

func TestScrolledRect(t *testing.T) {
	r := layout.Rect{X: 100, Y: 50, Width: 200, Height: 100}
	shifted := scrolledRect(r, 30)
	assert.Equal(t, 70.0, shifted.X)
	assert.Equal(t, 50.0, shifted.Y)
	assert.Equal(t, 200.0, shifted.Width)
	assert.Equal(t, 100.0, shifted.Height)
}

func TestScrolledRectY(t *testing.T) {
	r := layout.Rect{X: 100, Y: 50, Width: 200, Height: 100}
	shifted := scrolledRectY(r, 20)
	assert.Equal(t, 100.0, shifted.X)
	assert.Equal(t, 30.0, shifted.Y)
	assert.Equal(t, 200.0, shifted.Width)
	assert.Equal(t, 100.0, shifted.Height)
}

func TestVerticalScrollbar(t *testing.T) {
	tests := []struct {
		name            string
		html            string
		css             string
		expectTrack     bool
		expectThumbFull bool
	}{
		{
			name:            "scroll always shows scrollbar",
			html:            `<div class="box">short</div>`,
			css:             `.box { height: 200px; overflow-y: scroll; }`,
			expectTrack:     true,
			expectThumbFull: true,
		},
		{
			name:            "auto with overflow shows scrollbar",
			html:            `<div class="box"><div style="height:500px;">tall</div></div>`,
			css:             `.box { height: 200px; overflow-y: auto; }`,
			expectTrack:     true,
			expectThumbFull: false,
		},
		{
			name:        "auto without overflow no scrollbar",
			html:        `<div class="box">short</div>`,
			css:         `.box { height: 200px; overflow-y: auto; }`,
			expectTrack: false,
		},
		{
			name:        "hidden no scrollbar",
			html:        `<div class="box"><div style="height:500px;">tall</div></div>`,
			css:         `.box { height: 200px; overflow-y: hidden; }`,
			expectTrack: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := buildLayout(tt.html, tt.css, 800)
			commands := BuildDisplayList(root, InputState{}, LinkStyler{})

			tracks := findRectsByColor(commands, ColorScrollbarTrack)
			thumbs := findRectsByColor(commands, ColorScrollbarThumb)

			if tt.expectTrack {
				assert.NotEmpty(t, tracks, "expected scrollbar track")
				assert.NotEmpty(t, thumbs, "expected scrollbar thumb")

				track := tracks[0]
				thumb := thumbs[0]
				assert.Equal(t, ScrollbarWidth, track.Rect.Width, "track width")
				assert.True(t, thumb.Rect.Height > 0, "thumb should have positive height")

				if tt.expectThumbFull {
					expectedThumbH := track.Rect.Height - 2*scrollbarThumbPadding
					assert.InDelta(t, expectedThumbH, thumb.Rect.Height, 1, "thumb should fill track when no overflow")
				} else {
					assert.True(t, thumb.Rect.Height < track.Rect.Height, "thumb should be smaller than track when overflowing")
				}
			} else {
				assert.Empty(t, tracks, "expected no scrollbar track")
				assert.Empty(t, thumbs, "expected no scrollbar thumb")
			}
		})
	}
}

func TestVerticalScrollbarThumbMinHeight(t *testing.T) {
	// Very tall content should clamp thumb to minimum height
	html := `<div class="box"><div style="height:10000px;">x</div></div>`
	cssText := `.box { height: 200px; overflow-y: scroll; }`
	root := buildLayout(html, cssText, 800)
	commands := BuildDisplayList(root, InputState{}, LinkStyler{})

	thumbs := findRectsByColor(commands, ColorScrollbarThumb)
	assert.NotEmpty(t, thumbs)
	assert.True(t, thumbs[0].Rect.Height >= ScrollbarThumbMinHeight, "thumb height should be at least minimum")
}

func TestVerticalScrollbarThumbPosition(t *testing.T) {
	// When scroll offset is set, thumb should move down
	html := `<div class="box"><div style="height:1000px;">tall content</div></div>`
	cssText := `.box { height: 200px; overflow-y: scroll; }`
	root := buildLayout(html, cssText, 800)

	// Find the overflow container's DOM node
	var overflowNode *dom.Node
	var findNode func(box *layout.LayoutBox)
	findNode = func(box *layout.LayoutBox) {
		if box.Style.OverflowY == "scroll" && box.Node != nil {
			overflowNode = box.Node
			return
		}
		for _, child := range box.Children {
			findNode(child)
		}
	}
	findNode(root)
	assert.NotNil(t, overflowNode, "should find overflow node")

	// No scroll: thumb at top
	cmds0 := BuildDisplayList(root, InputState{}, LinkStyler{})
	thumbs0 := findRectsByColor(cmds0, ColorScrollbarThumb)
	assert.NotEmpty(t, thumbs0)
	thumbY0 := thumbs0[0].Rect.Y

	// With scroll offset: thumb moves down
	state := InputState{
		ScrollOffsetsY: map[*dom.Node]float64{overflowNode: 100},
	}
	cmds1 := BuildDisplayList(root, state, LinkStyler{})
	thumbs1 := findRectsByColor(cmds1, ColorScrollbarThumb)
	assert.NotEmpty(t, thumbs1)
	assert.True(t, thumbs1[0].Rect.Y > thumbY0, "thumb should move down when scrolled")
}

func TestVerticalScrollRevealsHiddenChildren(t *testing.T) {
	// Regression: children below ClipBottom must become visible when scrolled into view.
	// A 200px-tall container with 5 children of 100px each (total 500px content).
	html := `<div class="box">` +
		`<div class="item">Item 1</div>` +
		`<div class="item">Item 2</div>` +
		`<div class="item">Item 3</div>` +
		`<div class="item">Item 4</div>` +
		`<div class="item">Item 5</div>` +
		`</div>`
	cssText := `.box { height: 200px; overflow-y: scroll; } .item { height: 100px; }`
	root := buildLayout(html, cssText, 800)

	findTextCmd := func(commands []DisplayCommand, text string) *DrawText {
		for _, cmd := range commands {
			if dt, ok := cmd.(DrawText); ok && strings.Contains(dt.Text, text) {
				return &dt
			}
		}
		return nil
	}

	// Find overflow container node
	var overflowNode *dom.Node
	var findNode func(box *layout.LayoutBox)
	findNode = func(box *layout.LayoutBox) {
		if box.Style.OverflowY == "scroll" && box.Node != nil {
			overflowNode = box.Node
			return
		}
		for _, child := range box.Children {
			findNode(child)
		}
	}
	findNode(root)
	assert.NotNil(t, overflowNode)

	// Without scroll: items 1 and 2 visible, items 3-5 clipped
	cmds0 := BuildDisplayList(root, InputState{}, LinkStyler{})
	assert.NotNil(t, findTextCmd(cmds0, "Item 1"), "Item 1 should be visible without scroll")
	assert.NotNil(t, findTextCmd(cmds0, "Item 2"), "Item 2 should be visible without scroll")

	// With scroll offset 200: items 3 and 4 should now be visible
	state := InputState{
		ScrollOffsetsY: map[*dom.Node]float64{overflowNode: 200},
	}
	cmds1 := BuildDisplayList(root, state, LinkStyler{})
	assert.NotNil(t, findTextCmd(cmds1, "Item 3"), "Item 3 should be visible when scrolled down")
	assert.NotNil(t, findTextCmd(cmds1, "Item 4"), "Item 4 should be visible when scrolled down")
}

func TestNeedsVerticalScrollbar(t *testing.T) {
	tests := []struct {
		name     string
		box      *layout.LayoutBox
		style    TextStyle
		expected bool
	}{
		{
			name: "scroll always true",
			box: &layout.LayoutBox{
				Type:  layout.BlockBox,
				Rect:  layout.Rect{Y: 0, Height: 200},
				Style: css.Style{OverflowY: "scroll"},
			},
			style:    TextStyle{OverflowY: "scroll"},
			expected: true,
		},
		{
			name: "auto with overflow",
			box: &layout.LayoutBox{
				Type:    layout.BlockBox,
				Rect:    layout.Rect{Y: 0, Height: 200},
				Padding: layout.EdgeSizes{Bottom: 0},
				Style:   css.Style{OverflowY: "auto"},
				Children: []*layout.LayoutBox{
					{Rect: layout.Rect{Y: 0, Height: 300}},
				},
			},
			style:    TextStyle{OverflowY: "auto"},
			expected: true,
		},
		{
			name: "auto without overflow",
			box: &layout.LayoutBox{
				Type:    layout.BlockBox,
				Rect:    layout.Rect{Y: 0, Height: 200},
				Padding: layout.EdgeSizes{Bottom: 0},
				Style:   css.Style{OverflowY: "auto"},
				Children: []*layout.LayoutBox{
					{Rect: layout.Rect{Y: 0, Height: 100}},
				},
			},
			style:    TextStyle{OverflowY: "auto"},
			expected: false,
		},
		{
			name: "hidden never",
			box: &layout.LayoutBox{
				Type:  layout.BlockBox,
				Rect:  layout.Rect{Y: 0, Height: 200},
				Style: css.Style{OverflowY: "hidden"},
			},
			style:    TextStyle{OverflowY: "hidden"},
			expected: false,
		},
		{
			name: "inline box never",
			box: &layout.LayoutBox{
				Type:  layout.InlineBox,
				Rect:  layout.Rect{Y: 0, Height: 200},
				Style: css.Style{OverflowY: "scroll"},
			},
			style:    TextStyle{OverflowY: "scroll"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, needsVerticalScrollbar(tt.box, tt.style))
		})
	}
}

func TestMaxChildBottom(t *testing.T) {
	box := &layout.LayoutBox{
		Children: []*layout.LayoutBox{
			{Rect: layout.Rect{Y: 10, Height: 100}},
			{Rect: layout.Rect{Y: 50, Height: 200}},
			{Rect: layout.Rect{Y: 0, Height: 50}},
		},
	}
	assert.Equal(t, 250.0, maxChildBottom(box))
}

func TestMaxChildBottomEmpty(t *testing.T) {
	box := &layout.LayoutBox{}
	assert.Equal(t, 0.0, maxChildBottom(box))
}

func TestBothScrollbars(t *testing.T) {
	// Both overflow-x and overflow-y scroll should render both scrollbars plus corner
	html := `<div class="box"><div style="width:500px;height:500px;">content</div></div>`
	cssText := `.box { width: 200px; height: 200px; overflow: scroll; }`
	root := buildLayout(html, cssText, 800)
	commands := BuildDisplayList(root, InputState{}, LinkStyler{})

	tracks := findRectsByColor(commands, ColorScrollbarTrack)
	thumbs := findRectsByColor(commands, ColorScrollbarThumb)

	// Should have at least 3 tracks: horizontal, vertical, and corner
	assert.True(t, len(tracks) >= 3, "expected at least 3 track rects (h-track, v-track, corner), got %d", len(tracks))
	assert.True(t, len(thumbs) >= 2, "expected at least 2 thumbs (h-thumb, v-thumb), got %d", len(thumbs))

	// Verify we have both horizontal (height=12) and vertical (width=12) tracks
	hasHorizontalTrack := false
	hasVerticalTrack := false
	for _, track := range tracks {
		if track.Rect.Height == ScrollbarHeight && track.Rect.Width > ScrollbarWidth {
			hasHorizontalTrack = true
		}
		if track.Rect.Width == ScrollbarWidth && track.Rect.Height > ScrollbarHeight {
			hasVerticalTrack = true
		}
	}
	assert.True(t, hasHorizontalTrack, "expected horizontal scrollbar track")
	assert.True(t, hasVerticalTrack, "expected vertical scrollbar track")
}

func TestFirstLineStyleAppliedToWrappedLines(t *testing.T) {
	// Use a narrow container so text wraps into multiple lines
	html := `<div style="width: 100px;"><p class="fl">word1 word2 word3 word4 word5 word6 word7 word8</p></div>`
	cssText := `.fl::first-line { color: red; }`
	red := color.RGBA{255, 0, 0, 255}

	doc := dom.Parse(strings.NewReader(html))
	sheet := css.Parse(cssText)
	viewport := layout.Viewport{Width: 800, Height: 600}
	layoutRoot := layout.BuildLayoutTree(doc, sheet, viewport, css.MatchContext{})
	layout.ComputeLayout(layoutRoot, 800)

	commands := BuildDisplayList(layoutRoot, InputState{}, LinkStyler{})

	// Collect all DrawText commands
	var drawTexts []DrawText
	for _, cmd := range commands {
		if dt, ok := cmd.(DrawText); ok && dt.Text != "" {
			drawTexts = append(drawTexts, dt)
		}
	}

	assert.True(t, len(drawTexts) >= 2, "expected at least 2 wrapped lines, got %d", len(drawTexts))

	// First text line should be red
	r1, g1, b1, _ := drawTexts[0].Color.RGBA()
	rr, gr, br, _ := red.RGBA()
	assert.Equal(t, rr, r1, "first line red component")
	assert.Equal(t, gr, g1, "first line green component")
	assert.Equal(t, br, b1, "first line blue component")

	// Second text line should NOT be red
	r2, g2, b2, _ := drawTexts[1].Color.RGBA()
	isRed := r2 == rr && g2 == gr && b2 == br
	assert.False(t, isRed, "second line should not be red")
}

func TestFirstLineBackgroundEmitsDrawRect(t *testing.T) {
	html := `<div style="width: 100px;"><p class="fl">word1 word2 word3 word4 word5 word6 word7 word8</p></div>`
	cssText := `.fl::first-line { background-color: yellow; }`
	yellow := color.RGBA{255, 255, 0, 255}

	doc := dom.Parse(strings.NewReader(html))
	sheet := css.Parse(cssText)
	viewport := layout.Viewport{Width: 800, Height: 600}
	layoutRoot := layout.BuildLayoutTree(doc, sheet, viewport, css.MatchContext{})
	layout.ComputeLayout(layoutRoot, 800)

	commands := BuildDisplayList(layoutRoot, InputState{}, LinkStyler{})

	// Find the first DrawText and look for a yellow DrawRect before it
	firstTextIdx := -1
	for i, cmd := range commands {
		if dt, ok := cmd.(DrawText); ok && dt.Text != "" {
			firstTextIdx = i
			break
		}
	}
	assert.True(t, firstTextIdx > 0, "expected DrawText command")

	// There should be a yellow DrawRect just before the first DrawText
	foundYellowRect := false
	for i := 0; i < firstTextIdx; i++ {
		if dr, ok := commands[i].(DrawRect); ok {
			yr, yg, yb, _ := yellow.RGBA()
			cr, cg, cb, _ := dr.Color.RGBA()
			if cr == yr && cg == yg && cb == yb {
				foundYellowRect = true
				break
			}
		}
	}
	assert.True(t, foundYellowRect, "expected yellow DrawRect for first-line background")
}

func TestFirstLineSingleLineText(t *testing.T) {
	html := `<p class="fl">Short text</p>`
	cssText := `.fl::first-line { color: red; }`
	red := color.RGBA{255, 0, 0, 255}

	doc := dom.Parse(strings.NewReader(html))
	sheet := css.Parse(cssText)
	viewport := layout.Viewport{Width: 800, Height: 600}
	layoutRoot := layout.BuildLayoutTree(doc, sheet, viewport, css.MatchContext{})
	layout.ComputeLayout(layoutRoot, 800)

	commands := BuildDisplayList(layoutRoot, InputState{}, LinkStyler{})

	found := false
	for _, cmd := range commands {
		if dt, ok := cmd.(DrawText); ok && dt.Text == "Short text" {
			r, g, b, _ := dt.Color.RGBA()
			rr, gr, br, _ := red.RGBA()
			assert.Equal(t, rr, r, "single-line red component")
			assert.Equal(t, gr, g, "single-line green component")
			assert.Equal(t, br, b, "single-line blue component")
			found = true
		}
	}
	assert.True(t, found, "expected DrawText for 'Short text'")
}
