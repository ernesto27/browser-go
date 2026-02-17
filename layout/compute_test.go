package layout

import (
	"browser/css"
	"browser/dom"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestColWidth tests the getColWidth helper that reads width from a DOM node.
func TestColWidth(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]string
		tableWidth float64
		expected   float64
	}{
		{"pixel width attr", map[string]string{"width": "150"}, 600, 150},
		{"percent width attr", map[string]string{"width": "50%"}, 600, 300},
		{"CSS style width px", map[string]string{"style": "width: 200px"}, 600, 200},
		{"CSS style width percent", map[string]string{"style": "width: 25%"}, 600, 150},
		{"CSS style overrides attr", map[string]string{"width": "100", "style": "width: 300px"}, 600, 300},
		{"no width", map[string]string{}, 600, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &dom.Node{Type: dom.Element, TagName: dom.TagCol, Attributes: tt.attributes}
			result := getColWidth(node, tt.tableWidth)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExtractColWidths tests extracting column widths from a table DOM node.
func TestExtractColWidths(t *testing.T) {
	tests := []struct {
		name       string
		html       string
		tableWidth float64
		expected   []float64
	}{
		{
			name:       "single col with width",
			html:       `<table><colgroup><col width="100"><col width="200"></colgroup><tr><td>A</td><td>B</td></tr></table>`,
			tableWidth: 600,
			expected:   []float64{100, 200},
		},
		{
			name:       "col with span",
			html:       `<table><colgroup><col width="80"><col span="2" width="150"></colgroup><tr><td>A</td><td>B</td><td>C</td></tr></table>`,
			tableWidth: 600,
			expected:   []float64{80, 150, 150},
		},
		{
			name:       "colgroup with span no col children",
			html:       `<table><colgroup span="2" width="120"></colgroup><tr><td>A</td><td>B</td></tr></table>`,
			tableWidth: 600,
			expected:   []float64{120, 120},
		},
		{
			name:       "no colgroup",
			html:       `<table><tr><td>A</td></tr></table>`,
			tableWidth: 600,
			expected:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree(tt.html)
			tableNode := findBoxByTag(tree, "table")
			result := extractColWidths(tableNode.Node, tt.tableWidth)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestColWidthLayout tests that <col> widths drive the table column layout.
func TestColWidthLayout(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		containerWidth float64
		verify         func(t *testing.T, tree *LayoutBox)
	}{
		{
			name:           "col pixel width applied to cells",
			html:           `<table><colgroup><col width="100"><col width="250"><col width="150"></colgroup><tr><td>A</td><td>B</td><td>C</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				assert.Equal(t, 100.0, findCellByText(tree, "A").Rect.Width)
				assert.Equal(t, 250.0, findCellByText(tree, "B").Rect.Width)
				assert.Equal(t, 150.0, findCellByText(tree, "C").Rect.Width)
			},
		},
		{
			name:           "col span applies width to multiple columns",
			html:           `<table><colgroup><col width="80"><col span="2" width="200"></colgroup><tr><td>A</td><td>B</td><td>C</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				assert.Equal(t, 80.0, findCellByText(tree, "A").Rect.Width)
				assert.Equal(t, 200.0, findCellByText(tree, "B").Rect.Width)
				assert.Equal(t, 200.0, findCellByText(tree, "C").Rect.Width)
			},
		},
		{
			name:           "cell explicit width overrides col width",
			html:           `<table><colgroup><col width="80"><col width="80"><col width="80"></colgroup><tr><td>A</td><td style="width: 300px;">B</td><td>C</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				assert.Equal(t, 80.0, findCellByText(tree, "A").Rect.Width)
				assert.Equal(t, 300.0, findCellByText(tree, "B").Rect.Width)
				assert.Equal(t, 80.0, findCellByText(tree, "C").Rect.Width)
			},
		},
		{
			name:           "col width mixes with auto columns",
			html:           `<table><colgroup><col width="200"><col><col></colgroup><tr><td>A</td><td>B</td><td>C</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				assert.Equal(t, 200.0, findCellByText(tree, "A").Rect.Width)
				autoWidth := (table.Rect.Width - 200.0) / 2
				assert.Equal(t, autoWidth, findCellByText(tree, "B").Rect.Width)
				assert.Equal(t, autoWidth, findCellByText(tree, "C").Rect.Width)
			},
		},
		{
			name:           "col CSS style width",
			html:           `<table><colgroup><col style="width: 180px;"><col style="width: 120px;"></colgroup><tr><td>A</td><td>B</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				assert.Equal(t, 180.0, findCellByText(tree, "A").Rect.Width)
				assert.Equal(t, 120.0, findCellByText(tree, "B").Rect.Width)
			},
		},
		{
			name:           "colgroup span no col children",
			html:           `<table><colgroup span="2" width="150"></colgroup><tr><td>A</td><td>B</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				assert.Equal(t, 150.0, findCellByText(tree, "A").Rect.Width)
				assert.Equal(t, 150.0, findCellByText(tree, "B").Rect.Width)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree(tt.html)
			ComputeLayout(tree, tt.containerWidth)
			tt.verify(t, tree)
		})
	}
}

func TestGetLineHeight(t *testing.T) {
	tests := []struct {
		tag      string
		expected float64
	}{
		{"h1", 40.0},
		{"h2", 32.0},
		{"h3", 26.0},
		{"h4", 24.0},
		{"h5", 22.0},
		{"h6", 20.0},
		{"small", 18.0},
		{"p", 24.0},
		{"div", 24.0},
		{"span", 24.0},
		{"", 24.0},
		{"unknown", 24.0},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			result := getDefaultLineHeight(tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLineHeightFromStyle(t *testing.T) {
	tests := []struct {
		name        string
		lineHeight  float64
		tagName     string
		expected    float64
	}{
		{"style has line-height", 32.0, "p", 32.0},
		{"style has line-height overrides tag default", 50.0, "h1", 50.0},
		{"no line-height falls back to h1 default", 0, "h1", 40.0},
		{"no line-height falls back to h2 default", 0, "h2", 32.0},
		{"no line-height falls back to p default", 0, "p", 24.0},
		{"small line-height value", 12.0, "p", 12.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := css.Style{LineHeight: tt.lineHeight}
			result := getLineHeightFromStyle(style, tt.tagName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFontSize(t *testing.T) {
	tests := []struct {
		tag      string
		expected float64
	}{
		{"h1", 32.0},
		{"h2", 24.0},
		{"h3", 18.0},
		{"h4", 16.0},
		{"h5", 14.0},
		{"h6", 12.0},
		{"small", 12.0},
		{"p", 16.0},
		{"div", 16.0},
		{"span", 16.0},
		{"", 16.0},
		{"unknown", 16.0},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			result := getFontSize(tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetImageSize(t *testing.T) {
	tests := []struct {
		name           string
		attrs          map[string]string
		expectedWidth  float64
		expectedHeight float64
	}{
		{"nil node", nil, 200.0, 150.0},
		{"no attributes", map[string]string{}, 200.0, 150.0},
		{"width only", map[string]string{"width": "300"}, 300.0, 150.0},
		{"height only", map[string]string{"height": "200"}, 200.0, 200.0},
		{"both attributes", map[string]string{"width": "400", "height": "300"}, 400.0, 300.0},
		{"with px suffix", map[string]string{"width": "250px", "height": "180px"}, 250.0, 180.0},
		{"invalid width", map[string]string{"width": "abc"}, 200.0, 150.0},
		{"invalid height", map[string]string{"height": "xyz"}, 200.0, 150.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node *dom.Node
			if tt.attrs != nil {
				node = dom.NewElement("img", tt.attrs)
			}
			w, h := getImageSize(node)
			assert.Equal(t, tt.expectedWidth, w)
			assert.Equal(t, tt.expectedHeight, h)
		})
	}
}

func TestIsInsidePre(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *LayoutBox
		expected bool
	}{
		{
			name: "direct child of pre",
			setup: func() *LayoutBox {
				pre := createElementBox(BlockBox, "pre")
				text := createTextBox("code")
				addChild(pre, text)
				return text
			},
			expected: true,
		},
		{
			name: "grandchild of pre",
			setup: func() *LayoutBox {
				pre := createElementBox(BlockBox, "pre")
				span := createElementBox(InlineBox, "span")
				text := createTextBox("code")
				addChild(pre, span)
				addChild(span, text)
				return text
			},
			expected: true,
		},
		{
			name: "not inside pre",
			setup: func() *LayoutBox {
				div := createElementBox(BlockBox, "div")
				text := createTextBox("text")
				addChild(div, text)
				return text
			},
			expected: false,
		},
		{
			name: "no parent",
			setup: func() *LayoutBox {
				return createTextBox("orphan")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := tt.setup()
			result := isInsidePre(box)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetButtonText(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *LayoutBox
		expected string
	}{
		{
			name: "text child",
			setup: func() *LayoutBox {
				btn := createElementBox(ButtonBox, "button")
				text := createTextBox("Click Me")
				addChild(btn, text)
				return btn
			},
			expected: "Click Me",
		},
		{
			name: "value attribute",
			setup: func() *LayoutBox {
				attrs := map[string]string{"value": "Submit"}
				return &LayoutBox{
					Type: ButtonBox,
					Node: dom.NewElement("button", attrs),
				}
			},
			expected: "Submit",
		},
		{
			name: "default when no text or value",
			setup: func() *LayoutBox {
				return createElementBox(ButtonBox, "button")
			},
			expected: "Button",
		},
		{
			name: "text child takes priority over value",
			setup: func() *LayoutBox {
				attrs := map[string]string{"value": "Ignored"}
				btn := &LayoutBox{
					Type: ButtonBox,
					Node: dom.NewElement("button", attrs),
				}
				text := createTextBox("Visible")
				addChild(btn, text)
				return btn
			},
			expected: "Visible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := tt.setup()
			result := getButtonText(box)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyLineAlignment(t *testing.T) {
	tests := []struct {
		name       string
		boxes      []*LayoutBox
		innerX     float64
		innerWidth float64
		textAlign  string
		expectedX  []float64
	}{
		{
			name:       "empty boxes",
			boxes:      []*LayoutBox{},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "center",
			expectedX:  []float64{},
		},
		{
			name: "left alignment no change",
			boxes: []*LayoutBox{
				{Rect: Rect{X: 0, Width: 50}},
			},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "left",
			expectedX:  []float64{0},
		},
		{
			name: "center alignment",
			boxes: []*LayoutBox{
				{Rect: Rect{X: 0, Width: 50}},
			},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "center",
			expectedX:  []float64{25}, // (100-50)/2 = 25
		},
		{
			name: "right alignment",
			boxes: []*LayoutBox{
				{Rect: Rect{X: 0, Width: 50}},
			},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "right",
			expectedX:  []float64{50}, // 100-50 = 50
		},
		{
			name: "multiple boxes centered",
			boxes: []*LayoutBox{
				{Rect: Rect{X: 0, Width: 30}},
				{Rect: Rect{X: 30, Width: 20}},
			},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "center",
			expectedX:  []float64{25, 55}, // offset = (100-50)/2 = 25
		},
		{
			name: "empty text align",
			boxes: []*LayoutBox{
				{Rect: Rect{X: 0, Width: 50}},
			},
			innerX:     0,
			innerWidth: 100,
			textAlign:  "",
			expectedX:  []float64{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyLineAlignment(tt.boxes, tt.innerX, tt.innerWidth, tt.textAlign)
			for i, box := range tt.boxes {
				assert.Equal(t, tt.expectedX[i], box.Rect.X)
			}
		})
	}
}

func TestComputeLayout(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		containerWidth float64
		verify         func(t *testing.T, tree *LayoutBox)
	}{
		{
			name:           "div inside body offset by margin",
			html:           "<div></div>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				div := findBoxByTag(tree, "div")
				// div is inside body which has 8px margin
				assert.Equal(t, 8.0, div.Rect.X)
				assert.Equal(t, 8.0, div.Rect.Y)
			},
		},
		{
			name:           "div width reduced by body margin",
			html:           "<div></div>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				div := findBoxByTag(tree, "div")
				// 800 - 8*2 = 784
				assert.Equal(t, 784.0, div.Rect.Width)
			},
		},
		{
			name:           "body has 8px margin",
			html:           "<body><p>Text</p></body>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				body := findBoxByTag(tree, "body")
				assert.Equal(t, 8.0, body.Margin.Top)
				assert.Equal(t, 8.0, body.Margin.Right)
				assert.Equal(t, 8.0, body.Margin.Bottom)
				assert.Equal(t, 8.0, body.Margin.Left)
			},
		},
		{
			name:           "p has vertical margin",
			html:           "<p>Text</p>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				p := findBoxByTag(tree, "p")
				// Browser default: 1em margin (16px)
				assert.Equal(t, 16.0, p.Margin.Top)
				assert.Equal(t, 16.0, p.Margin.Bottom)
			},
		},
		{
			name:           "h1 has vertical margin",
			html:           "<h1>Title</h1>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				h1 := findBoxByTag(tree, "h1")
				// Browser default: 0.67em margin (16px * 0.67 = 10.72)
				assert.Equal(t, 10.72, h1.Margin.Top)
				assert.Equal(t, 10.72, h1.Margin.Bottom)
			},
		},
		{
			name:           "hr has fixed height",
			html:           "<div><hr></div>",
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				hr := findBoxByTag(tree, "hr")
				assert.Equal(t, 2.0, hr.Rect.Height)
			},
		},
		{
			name:           "explicit CSS width respected",
			html:           `<div style="width: 400px"></div>`,
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				div := findBoxByTag(tree, "div")
				assert.Equal(t, 400.0, div.Rect.Width)
			},
		},
		{
			name:           "min-width respected",
			html:           `<div style="min-width: 500px"></div>`,
			containerWidth: 300,
			verify: func(t *testing.T, tree *LayoutBox) {
				div := findBoxByTag(tree, "div")
				assert.Equal(t, 500.0, div.Rect.Width)
			},
		},
		{
			name:           "explicit CSS height respected",
			html:           `<div style="height: 100px"></div>`,
			containerWidth: 800,
			verify: func(t *testing.T, tree *LayoutBox) {
				div := findBoxByTag(tree, "div")
				assert.Equal(t, 100.0, div.Rect.Height)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree(tt.html)
			ComputeLayout(tree, tt.containerWidth)
			tt.verify(t, tree)
		})
	}
}

func TestGetCellRowSpan(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]string
		expected int
	}{
		{"no attribute", nil, 1},
		{"rowspan=1", map[string]string{"rowspan": "1"}, 1},
		{"rowspan=2", map[string]string{"rowspan": "2"}, 2},
		{"rowspan=5", map[string]string{"rowspan": "5"}, 5},
		{"rowspan=0 defaults to 1", map[string]string{"rowspan": "0"}, 1},
		{"negative defaults to 1", map[string]string{"rowspan": "-1"}, 1},
		{"invalid defaults to 1", map[string]string{"rowspan": "abc"}, 1},
		{"empty defaults to 1", map[string]string{"rowspan": ""}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := &LayoutBox{
				Type: TableCellBox,
				Node: dom.NewElement("td", tt.attrs),
			}
			assert.Equal(t, tt.expected, getCellRowSpan(cell))
		})
	}
}

func TestRowspanLayout(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		containerWidth float64
		verify         func(t *testing.T, tree *LayoutBox)
	}{
		{
			name:           "basic rowspan=2 cell spans two rows",
			html:           `<table><tr><td rowspan="2">A</td><td>B</td><td>C</td></tr><tr><td>D</td><td>E</td></tr></table>`,
			containerWidth: 300,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				rows := findTableRows(table)
				assert.Equal(t, 2, len(rows))

				colWidth := table.Rect.Width / 3 // 3 logical columns

				cellA := findCellByText(table, "A")
				cellB := findCellByText(table, "B")
				cellD := findCellByText(table, "D")
				cellE := findCellByText(table, "E")

				// A at col 0, B at col 1
				assert.Equal(t, table.Rect.X, cellA.Rect.X)
				assert.Equal(t, table.Rect.X+colWidth, cellB.Rect.X)

				// Row 1: col 0 occupied by A → D at col 1, E at col 2
				assert.Equal(t, table.Rect.X+colWidth, cellD.Rect.X)
				assert.Equal(t, table.Rect.X+2*colWidth, cellE.Rect.X)

				// A spans both rows
				assert.Equal(t, rows[0].Rect.Height+rows[1].Rect.Height, cellA.Rect.Height)
				assert.Equal(t, rows[0].Rect.Y, cellA.Rect.Y)
			},
		},
		{
			name:           "rowspan=3 spanning all rows",
			html:           `<table><tr><td>A</td><td rowspan="3">Side</td><td>C1</td></tr><tr><td>B</td><td>C2</td></tr><tr><td>C</td><td>C3</td></tr></table>`,
			containerWidth: 300,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				rows := findTableRows(table)
				assert.Equal(t, 3, len(rows))

				colWidth := table.Rect.Width / 3

				cellSide := findCellByText(table, "Side")

				// Side spans all 3 rows
				totalHeight := rows[0].Rect.Height + rows[1].Rect.Height + rows[2].Rect.Height
				assert.Equal(t, totalHeight, cellSide.Rect.Height)
				assert.Equal(t, rows[0].Rect.Y, cellSide.Rect.Y)

				// Side is at col 1
				assert.Equal(t, table.Rect.X+colWidth, cellSide.Rect.X)
			},
		},
		{
			name:           "rowspan + colspan combined",
			html:           `<table><tr><td rowspan="2" colspan="2">Big</td><td>C1</td></tr><tr><td>C2</td></tr><tr><td>A3</td><td>B3</td><td>C3</td></tr></table>`,
			containerWidth: 300,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				rows := findTableRows(table)
				assert.Equal(t, 3, len(rows))

				colWidth := table.Rect.Width / 3

				cellBig := findCellByText(table, "Big")
				cellC1 := findCellByText(table, "C1")
				cellC2 := findCellByText(table, "C2")
				cellA3 := findCellByText(table, "A3")

				// Big takes 2 columns wide
				assert.Equal(t, 2*colWidth, cellBig.Rect.Width)

				// C1 and C2 are at col 2
				assert.Equal(t, table.Rect.X+2*colWidth, cellC1.Rect.X)
				assert.Equal(t, table.Rect.X+2*colWidth, cellC2.Rect.X)

				// Row 2: normal — A3 at col 0
				assert.Equal(t, table.Rect.X, cellA3.Rect.X)

				// Big spans 2 rows vertically
				assert.Equal(t, rows[0].Rect.Height+rows[1].Rect.Height, cellBig.Rect.Height)
			},
		},
		{
			name:           "rowspan exceeding available rows is clamped",
			html:           `<table><tr><td rowspan="10">A</td><td>B</td></tr><tr><td>C</td></tr></table>`,
			containerWidth: 200,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				rows := findTableRows(table)

				cellA := findCellByText(table, "A")

				// A claims 10 rows but only 2 exist — clamped
				assert.Equal(t, rows[0].Rect.Height+rows[1].Rect.Height, cellA.Rect.Height)
				assert.Equal(t, rows[0].Rect.Height+rows[1].Rect.Height, table.Rect.Height)
			},
		},
		{
			name:           "rowspan=1 behaves like no rowspan",
			html:           `<table><tr><td rowspan="1">A</td><td>B</td></tr><tr><td>C</td><td>D</td></tr></table>`,
			containerWidth: 200,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				rows := findTableRows(table)

				cellA := findCellByText(table, "A")
				cellC := findCellByText(table, "C")

				// A only spans row 0
				assert.Equal(t, rows[0].Rect.Height, cellA.Rect.Height)

				// C is at col 0 (not pushed)
				assert.Equal(t, table.Rect.X, cellC.Rect.X)
			},
		},
		{
			name:           "row with all columns occupied by rowspan",
			html:           `<table><tr><td rowspan="2">A</td><td rowspan="2">B</td></tr><tr></tr><tr><td>C</td><td>D</td></tr></table>`,
			containerWidth: 200,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				rows := findTableRows(table)

				colWidth := table.Rect.Width / 2

				cellA := findCellByText(table, "A")
				cellB := findCellByText(table, "B")

				// A and B each span rows 0+1
				assert.Equal(t, rows[0].Rect.Height+rows[1].Rect.Height, cellA.Rect.Height)
				assert.Equal(t, rows[0].Rect.Height+rows[1].Rect.Height, cellB.Rect.Height)

				// Row 2: C at col 0, D at col 1
				cellC := findCellByText(table, "C")
				cellD := findCellByText(table, "D")
				assert.Equal(t, table.Rect.X, cellC.Rect.X)
				assert.Equal(t, table.Rect.X+colWidth, cellD.Rect.X)
			},
		},
		{
			name:           "rowspan with extra columns from rowspan push",
			html:           `<table><tr><td rowspan="2">A</td><td>B</td></tr><tr><td>C</td><td>D</td></tr></table>`,
			containerWidth: 300,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")

				// Row 0: A(rs=2) + B → 2 direct cells
				// Row 1: skip col 0, C at col 1, D at col 2 → 3 columns!
				colWidth := table.Rect.Width / 3

				cellA := findCellByText(table, "A")
				cellC := findCellByText(table, "C")
				cellD := findCellByText(table, "D")

				assert.Equal(t, table.Rect.X, cellA.Rect.X)
				assert.Equal(t, table.Rect.X+colWidth, cellC.Rect.X)
				assert.Equal(t, table.Rect.X+2*colWidth, cellD.Rect.X)
				assert.Equal(t, colWidth, cellC.Rect.Width)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree(tt.html)
			ComputeLayout(tree, tt.containerWidth)
			tt.verify(t, tree)
		})
	}
}

func TestTableCellCSSWidth(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		containerWidth float64
		verify         func(t *testing.T, tree *LayoutBox)
	}{
		{
			name:           "explicit pixel width on td",
			html:           `<table><tr><td style="width: 200px;">A</td><td>B</td><td>C</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				cellA := findCellByText(tree, "A")
				cellB := findCellByText(tree, "B")
				cellC := findCellByText(tree, "C")
				assert.Equal(t, 200.0, cellA.Rect.Width)
				autoWidth := (table.Rect.Width - 200.0) / 2
				assert.Equal(t, autoWidth, cellB.Rect.Width)
				assert.Equal(t, autoWidth, cellC.Rect.Width)
			},
		},
		{
			name:           "all cells with explicit width",
			html:           `<table><tr><td style="width: 100px;">A</td><td style="width: 200px;">B</td><td style="width: 150px;">C</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				cellA := findCellByText(tree, "A")
				cellB := findCellByText(tree, "B")
				cellC := findCellByText(tree, "C")
				assert.Equal(t, 100.0, cellA.Rect.Width)
				assert.Equal(t, 200.0, cellB.Rect.Width)
				assert.Equal(t, 150.0, cellC.Rect.Width)
			},
		},
		{
			name:           "width set in second row affects column",
			html:           `<table><tr><td>A</td><td>B</td></tr><tr><td style="width: 250px;">C</td><td>D</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				cellA := findCellByText(tree, "A")
				cellC := findCellByText(tree, "C")
				assert.Equal(t, 250.0, cellA.Rect.Width)
				assert.Equal(t, 250.0, cellC.Rect.Width)
			},
		},
		{
			name:           "percentage width on td",
			html:           `<table><tr><td style="width: 25%;">A</td><td style="width: 50%;">B</td><td style="width: 25%;">C</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				tw := table.Rect.Width
				cellA := findCellByText(tree, "A")
				cellB := findCellByText(tree, "B")
				cellC := findCellByText(tree, "C")
				assert.Equal(t, tw*0.25, cellA.Rect.Width)
				assert.Equal(t, tw*0.50, cellB.Rect.Width)
				assert.Equal(t, tw*0.25, cellC.Rect.Width)
			},
		},
		{
			name:           "width with colspan uses summed column widths",
			html:           `<table><tr><td style="width: 120px;">A</td><td style="width: 120px;">B</td><td style="width: 120px;">C</td></tr><tr><td colspan="2">D</td><td>E</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				cellA := findCellByText(tree, "A")
				cellD := findCellByText(tree, "D")
				cellE := findCellByText(tree, "E")
				assert.Equal(t, 120.0, cellA.Rect.Width)
				assert.Equal(t, 240.0, cellD.Rect.Width)
				assert.Equal(t, 120.0, cellE.Rect.Width)
			},
		},
		{
			name:           "no explicit widths falls back to equal distribution",
			html:           `<table><tr><td>A</td><td>B</td><td>C</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				eqWidth := table.Rect.Width / 3
				cellA := findCellByText(tree, "A")
				cellB := findCellByText(tree, "B")
				cellC := findCellByText(tree, "C")
				assert.Equal(t, eqWidth, cellA.Rect.Width)
				assert.Equal(t, eqWidth, cellB.Rect.Width)
				assert.Equal(t, eqWidth, cellC.Rect.Width)
			},
		},
		{
			name:           "table width attribute percent",
			html:           `<table width="85%"><tr><td>A</td><td>B</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				body := findBoxByTag(tree, "body")
				availableWidth := 600.0 - body.Margin.Left - body.Margin.Right
				assert.InDelta(t, availableWidth*0.85, table.Rect.Width, 0.1)
			},
		},
		{
			name:           "table width attribute pixels",
			html:           `<table width="400"><tr><td>A</td><td>B</td></tr></table>`,
			containerWidth: 600,
			verify: func(t *testing.T, tree *LayoutBox) {
				table := findBoxByTag(tree, "table")
				assert.InDelta(t, 400.0, table.Rect.Width, 0.1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree(tt.html)
			ComputeLayout(tree, tt.containerWidth)
			tt.verify(t, tree)
		})
	}
}

func TestTableCellTextWrapping(t *testing.T) {
	// MeasureText estimation: fontSize(16) * 0.5 * len(text) = 8 * len(text) pixels.
	// cellPadding = 8, so inner width = cellWidth - 16.
	// "Hello World" in a 100px cell: inner = 84px.
	//   "Hello" = 40px fits; "Hello World" = 88px > 84px → wraps to ["Hello", "World"].
	tests := []struct {
		name   string
		html   string
		verify func(t *testing.T, tree *LayoutBox)
	}{
		{
			name: "short text does not wrap",
			html: `<table><tr><td style="width: 300px;">Hi</td></tr></table>`,
			verify: func(t *testing.T, tree *LayoutBox) {
				cell := findCellByText(tree, "Hi")
				assert.NotNil(t, cell)
				textBox := findTextBoxInSubtree(cell, "Hi")
				assert.NotNil(t, textBox)
				assert.True(t, len(textBox.WrappedLines) <= 1, "short text should not have multiple wrapped lines")
				assert.Equal(t, 24.0, textBox.Rect.Height)
			},
		},
		{
			name: "long text wraps in narrow cell",
			html: `<table><tr><td style="width: 100px;">Hello World</td></tr></table>`,
			verify: func(t *testing.T, tree *LayoutBox) {
				cell := findCellByText(tree, "Hello World")
				assert.NotNil(t, cell)
				textBox := findTextBoxInSubtree(cell, "Hello World")
				assert.NotNil(t, textBox)
				assert.Greater(t, len(textBox.WrappedLines), 1, "text should wrap to multiple lines")
				assert.Equal(t, []string{"Hello", "World"}, textBox.WrappedLines)
			},
		},
		{
			name: "wrapped text increases cell height",
			html: `<table><tr><td style="width: 100px;">Hello World</td></tr></table>`,
			verify: func(t *testing.T, tree *LayoutBox) {
				cell := findCellByText(tree, "Hello World")
				assert.NotNil(t, cell)
				lineHeight := 24.0
				cellPadding := 8.0
				// 2 wrapped lines + top and bottom padding
				assert.Equal(t, 2*lineHeight+2*cellPadding, cell.Rect.Height)
			},
		},
		{
			name: "wrapped text box height spans all lines",
			html: `<table><tr><td style="width: 100px;">Hello World</td></tr></table>`,
			verify: func(t *testing.T, tree *LayoutBox) {
				cell := findCellByText(tree, "Hello World")
				assert.NotNil(t, cell)
				textBox := findTextBoxInSubtree(cell, "Hello World")
				assert.NotNil(t, textBox)
				lineHeight := 24.0
				assert.Equal(t, 2*lineHeight, textBox.Rect.Height)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree(tt.html)
			ComputeLayout(tree, 600)
			tt.verify(t, tree)
		})
	}
}

func TestTableCellVerticalAlign(t *testing.T) {
	// "Hello World" in a 100px cell wraps to 2 lines:
	//   content height = 2 * 24 = 48px
	//   cell height    = 48 + 16 = 64px  (+ 2 * cellPadding)
	//   innerHeight    = 64 - 16 = 48px
	// Short cell has 1 line = 24px content.
	// Empty space = 48 - 24 = 24px.
	const cellPadding = 8.0
	const lineHeight = 24.0
	const rowH = 2*lineHeight + 2*cellPadding // 64

	tests := []struct {
		name   string
		html   string
		wantDY float64
	}{
		{
			name:   "top - no shift",
			html:   `<table><tr><td style="width:100px;">Hello World</td><td style="vertical-align:top;">X</td></tr></table>`,
			wantDY: 0,
		},
		{
			name:   "middle - centered",
			html:   `<table><tr><td style="width:100px;">Hello World</td><td style="vertical-align:middle;">X</td></tr></table>`,
			wantDY: (rowH - 2*cellPadding - lineHeight) / 2, // 12
		},
		{
			name:   "bottom - full shift",
			html:   `<table><tr><td style="width:100px;">Hello World</td><td style="vertical-align:bottom;">X</td></tr></table>`,
			wantDY: rowH - 2*cellPadding - lineHeight, // 24
		},
		{
			name:   "valign middle attribute",
			html:   `<table><tr><td style="width:100px;">Hello World</td><td valign="middle">X</td></tr></table>`,
			wantDY: (rowH - 2*cellPadding - lineHeight) / 2,
		},
		{
			name:   "valign bottom attribute",
			html:   `<table><tr><td style="width:100px;">Hello World</td><td valign="bottom">X</td></tr></table>`,
			wantDY: rowH - 2*cellPadding - lineHeight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree(tt.html)
			ComputeLayout(tree, 600)

			cell := findCellByText(tree, "X")
			assert.NotNil(t, cell)
			textBox := findTextBoxInSubtree(cell, "X")
			assert.NotNil(t, textBox)

			expectedY := cell.Rect.Y + cellPadding + tt.wantDY
			assert.Equal(t, expectedY, textBox.Rect.Y)
		})
	}
}
