package layout

import (
	"browser/css"
	"browser/dom"
	"browser/utils"
	"fmt"
	"strconv"
	"strings"
)

const (
	DefaultMargin      = 8.0
	DefaultImageWidth  = 200.0
	DefaultImageHeight = 150.0
)

// getDefaultLineHeight returns default line heights for different elements
func getDefaultLineHeight(tagName string) float64 {
	switch tagName {
	case dom.TagH1:
		return 40.0
	case dom.TagH2:
		return 32.0
	case dom.TagH3:
		return 26.0
	case dom.TagH4:
		return 24.0
	case dom.TagH5:
		return 22.0
	case dom.TagH6:
		return 20.0
	case dom.TagSmall:
		return 18.0
	default:
		return 24.0
	}
}

// Font sizes for text measurement (should match render/paint.go)
func getFontSize(tagName string) float64 {
	switch tagName {
	case dom.TagH1:
		return 32.0
	case dom.TagH2:
		return 24.0
	case dom.TagH3:
		return 18.0
	case dom.TagH4:
		return 16.0
	case dom.TagH5:
		return 14.0
	case dom.TagH6:
		return 12.0
	case dom.TagSmall:
		return 12.0
	default:
		return 16.0
	}
}

func ComputeLayout(root *LayoutBox, containerWidth float64) {
	computeBlockLayout(root, containerWidth, 0, 0, "")
}

func computeBlockLayout(box *LayoutBox, containerWidth float64, startX, startY float64, parentTag string) {
	// Separate positioned children from normal flow
	var positionedChildren []*LayoutBox
	var floatedChildren []*LayoutBox
	var normalChildren []*LayoutBox

	for _, child := range box.Children {
		if child.Position == "absolute" {
			positionedChildren = append(positionedChildren, child)
		} else if child.Float == "left" || child.Float == "right" {
			floatedChildren = append(floatedChildren, child)
		} else {
			normalChildren = append(normalChildren, child)
		}
	}
	box.Children = normalChildren

	box.Rect.X = startX
	box.Rect.Y = startY
	box.Rect.Width = containerWidth

	if box.Style.Width > 0 {
		box.Rect.Width = box.Style.Width
	}

	if box.Style.MinWidth > 0 && box.Rect.Width < box.Style.MinWidth {
		box.Rect.Width = box.Style.MinWidth
	}

	if box.Style.MaxWidth > 0 && box.Rect.Width > box.Style.MaxWidth {
		box.Rect.Width = box.Style.MaxWidth
	}

	innerX := startX
	innerWidth := box.Rect.Width

	// Get current tag name
	currentTag := ""
	if box.Node != nil {
		currentTag = box.Node.TagName
	}

	// Body gets margin
	if currentTag == dom.TagBody {
		box.Margin = EdgeSizes{
			Top: DefaultMargin, Right: DefaultMargin,
			Bottom: DefaultMargin, Left: DefaultMargin,
		}
		innerX = startX + DefaultMargin
		innerWidth = containerWidth - (DefaultMargin * 2)
	}

	// Lists get indentation
	if currentTag == dom.TagUL || currentTag == dom.TagOL || currentTag == dom.TagMenu {
		innerX = startX + 20
		innerWidth = containerWidth - 20
	}

	if currentTag == dom.TagBlockquote {
		innerX = startX + 30
		innerWidth = containerWidth - 30
	}

	if currentTag == dom.TagDD {
		innerX = startX + 40
		innerWidth = containerWidth - 40
	}

	// Fieldset default styling
	if box.Type == FieldsetBox {
		box.Padding = EdgeSizes{Top: 10, Right: 10, Bottom: 10, Left: 10}
		box.Style.BorderTopWidth = 1
		box.Style.BorderRightWidth = 1
		box.Style.BorderBottomWidth = 1
		box.Style.BorderLeftWidth = 1
	}

	// Default margins for block elements
	switch currentTag {
	case dom.TagP:
		box.Margin.Top = 12
		box.Margin.Bottom = 12
	case dom.TagH1:
		box.Margin.Top = 6
		box.Margin.Bottom = 6
	case dom.TagH2:
		box.Margin.Top = 5
		box.Margin.Bottom = 5
	case dom.TagH3:
		box.Margin.Top = 4
		box.Margin.Bottom = 4
	case dom.TagH4, dom.TagH5, dom.TagH6:
		box.Margin.Top = 4
		box.Margin.Bottom = 4
	case dom.TagUL, dom.TagOL, dom.TagMenu:
		box.Margin.Top = 4
		box.Margin.Bottom = 4
	case dom.TagFigure:
		box.Margin.Top = 16
		box.Margin.Bottom = 16
		innerX = startX + 40
		innerWidth = containerWidth - 40
	}

	// Apply CSS margins from inline style (override defaults)
	if box.Style.MarginTop > 0 {
		box.Margin.Top = box.Style.MarginTop
	}
	if box.Style.MarginBottom > 0 {
		box.Margin.Bottom = box.Style.MarginBottom
	}

	// Handle auto margins for horizontal centering
	if box.Style.MarginLeftAuto && box.Style.MarginRightAuto && box.Style.Width > 0 {
		// Both auto = center horizontally
		leftover := containerWidth - box.Rect.Width
		if leftover > 0 {
			autoMargin := leftover / 2
			box.Rect.X = startX + autoMargin
			innerX = box.Rect.X
			box.Margin.Left = autoMargin
			box.Margin.Right = autoMargin
		}
	} else {
		if box.Style.MarginLeft > 0 {
			innerX += box.Style.MarginLeft
			innerWidth -= box.Style.MarginLeft
		}
		if box.Style.MarginRight > 0 {
			innerWidth -= box.Style.MarginRight
		}
	}

	// Apply CSS padding from inline style
	if box.Style.PaddingTop > 0 {
		box.Padding.Top = box.Style.PaddingTop
	}
	if box.Style.PaddingBottom > 0 {
		box.Padding.Bottom = box.Style.PaddingBottom
	}
	if box.Style.PaddingLeft > 0 {
		box.Padding.Left = box.Style.PaddingLeft
		innerX += box.Style.PaddingLeft
		innerWidth -= box.Style.PaddingLeft
	}
	if box.Style.PaddingRight > 0 {
		box.Padding.Right = box.Style.PaddingRight
		innerWidth -= box.Style.PaddingRight
	}

	// Apply border widths to inner content area
	if box.Style.BorderLeftWidth > 0 {
		innerX += box.Style.BorderLeftWidth
		innerWidth -= box.Style.BorderLeftWidth
	}
	if box.Style.BorderRightWidth > 0 {
		innerWidth -= box.Style.BorderRightWidth
	}

	yOffset := startY + box.Margin.Top + box.Padding.Top + box.Style.BorderTopWidth

	// Line state for inline flow
	currentX := innerX
	lineStartY := yOffset
	lineHeight := 0.0
	var lineBoxes []*LayoutBox

	// Handle legend for fieldset
	var legendBox *LayoutBox
	if box.Type == FieldsetBox {
		for i, child := range box.Children {
			if child.Type == LegendBox {
				legendBox = child
				// Remove legend from normal children flow
				box.Children = append(box.Children[:i], box.Children[i+1:]...)
				break
			}
		}

		// Position legend on the border
		if legendBox != nil {
			legendText := GetLegendText(legendBox)
			legendWidth := MeasureText(legendText, 16) + 16 // 8px padding each side
			legendHeight := 20.0

			legendBox.Rect.X = innerX + 12                              // 12px from left edge
			legendBox.Rect.Y = startY + box.Margin.Top - legendHeight/2 // Centered on border
			legendBox.Rect.Width = legendWidth
			legendBox.Rect.Height = legendHeight

			// Layout legend's children (the text)
			textX := legendBox.Rect.X + 4
			for _, child := range legendBox.Children {
				if child.Type == TextBox {
					child.Rect.X = textX
					child.Rect.Y = legendBox.Rect.Y
					child.Rect.Width = MeasureText(child.Text, 16)
					child.Rect.Height = legendHeight
				}
			}

			// Add legend back to children so paint.go can find it
			box.Children = append([]*LayoutBox{legendBox}, box.Children...)
		}
	}

	for _, child := range box.Children {
		// Skip LegendBox - already positioned above
		if child.Type == LegendBox {
			continue
		}

		var childWidth, childHeight float64

		switch child.Type {
		case TextBox:
			fontSize := getFontSize(parentTag)
			// Check if inside a <pre> element
			if isInsidePre(child) {
				// Handle multi-line preformatted text
				lines := strings.Split(child.Text, "\n")
				lineHeight := fontSize * 1.5 // Match render/paint.go line height

				// Find the widest line
				maxWidth := 0.0
				for _, line := range lines {
					w := MeasureText(line, fontSize)
					if w > maxWidth {
						maxWidth = w
					}
				}

				childWidth = maxWidth
				childHeight = float64(len(lines)) * lineHeight
			} else {
				// Wrap text to fit container width
				child.WrappedLines = WrapText(child.Text, fontSize, innerWidth)

				lineHeight := getLineHeightFromStyle(box.Style, parentTag)
				numLines := len(child.WrappedLines)
				if numLines == 0 {
					numLines = 1
				}

				// Width is the widest wrapped line
				maxLineWidth := 0.0
				for _, line := range child.WrappedLines {
					w := MeasureText(line, fontSize)
					if w > maxLineWidth {
						maxLineWidth = w
					}
				}
				childWidth = maxLineWidth
				childHeight = float64(numLines) * lineHeight
			}

		case InlineBox:
			// Compute inline box size from its content
			childWidth, childHeight = computeInlineSize(child, parentTag)

		case ImageBox:
			childWidth, childHeight = getImageSize(child.Node)
			childWidth += 4 // Add small right margin between images
		case InputBox:
			childWidth = 200.0
			childHeight = 28.0
		case RadioBox:
			childWidth = 20.0
			childHeight = 20.0
		case CheckboxBox:
			childWidth = 20.0
			childHeight = 20.0
		case ButtonBox:
			buttonText := getButtonText(child)
			fontSize := getFontSize(parentTag)
			childWidth = MeasureText(buttonText, fontSize) + 24.0
			childHeight = 32.0
		case TextareaBox:
			childWidth = 300.0
			childHeight = 80.0
		case SelectBox:
			childWidth = 200.0
			childHeight = 28.0
		case FileInputBox:
			childWidth = 250.0
			childHeight = 32.0

		case HRBox:
			// Block element - flush line first
			applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
			lineBoxes = nil
			if lineHeight > 0 {
				yOffset = lineStartY + lineHeight
			}
			child.Rect.X = innerX
			child.Rect.Y = yOffset + 8
			child.Rect.Width = innerWidth
			child.Rect.Height = 2
			yOffset += 18
			// Reset line state
			currentX = innerX
			lineStartY = yOffset
			lineHeight = 0
			continue

		case BRBox:
			// Line break - flush current line
			applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
			lineBoxes = nil
			if lineHeight > 0 {
				yOffset = lineStartY + lineHeight
			} else {
				yOffset += getLineHeightFromStyle(box.Style, parentTag)
			}
			child.Rect.X = currentX
			child.Rect.Y = yOffset
			child.Rect.Width = 0
			child.Rect.Height = 0
			currentX = innerX
			lineStartY = yOffset
			lineHeight = 0
			continue

		case TableBox:
			applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
			lineBoxes = nil
			computeTableLayout(child, innerWidth, innerX, yOffset)
			yOffset += child.Rect.Height
			// Reset line state
			currentX = innerX
			lineStartY = yOffset
			lineHeight = 0
			continue

		default:
			// Block element - flush line first
			applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
			lineBoxes = nil
			if lineHeight > 0 {
				yOffset = lineStartY + lineHeight
				lineStartY = yOffset
				lineHeight = 0
			}
			currentX = innerX

			childTag := ""
			if child.Node != nil {
				childTag = child.Node.TagName
			}
			computeBlockLayout(child, innerWidth, innerX, yOffset, childTag)
			yOffset += child.Rect.Height
			lineStartY = yOffset
			continue
		}

		// Inline element - check if we need to wrap
		if currentX+childWidth > innerX+innerWidth && currentX > innerX {
			// Wrap to new line - apply alignment first
			applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
			lineBoxes = nil
			yOffset = lineStartY + lineHeight
			currentX = innerX
			lineStartY = yOffset
			lineHeight = 0
		}

		// Position inline element
		child.Rect.X = currentX
		child.Rect.Y = lineStartY
		child.Rect.Width = childWidth
		child.Rect.Height = childHeight

		// For InlineBox, position its children within it
		if child.Type == InlineBox {
			layoutInlineChildren(child, parentTag)
		}

		// Track this element for alignment
		lineBoxes = append(lineBoxes, child)

		// Advance horizontal position
		currentX += childWidth
		if childHeight > lineHeight {
			lineHeight = childHeight
		}
	}

	// Final line
	applyLineAlignment(lineBoxes, innerX, innerWidth, box.Style.TextAlign)
	if lineHeight > 0 {
		yOffset = lineStartY + lineHeight
	}

	if box.Style.Height > 0 {
		box.Rect.Height = box.Style.Height
	} else {
		box.Rect.Height = yOffset - startY + box.Margin.Bottom + box.Padding.Bottom + box.Style.BorderBottomWidth
	}

	if box.Style.MinHeight > 0 && box.Rect.Height < box.Style.MinHeight {
		box.Rect.Height = box.Style.MinHeight
	}

	if box.Style.MaxHeight > 0 && box.Rect.Height > box.Style.MaxHeight {
		box.Rect.Height = box.Style.MaxHeight
	}

	// Position absolute children
	for _, child := range positionedChildren {
		childWidth := child.Style.Width
		if childWidth <= 0 {
			childWidth = containerWidth
		}

		// First, compute layout to determine child dimensions
		computeBlockLayout(child, childWidth, 0, 0, "")

		// Calculate X position
		childX := startX
		if child.Left > 0 {
			childX = startX + child.Left
		} else if child.Right > 0 {
			childX = startX + box.Rect.Width - child.Right - child.Rect.Width
		}

		// Calculate Y position
		childY := startY
		if child.Top > 0 {
			childY = startY + child.Top
		} else if child.Bottom > 0 {
			childY = startY + box.Rect.Height - child.Bottom - child.Rect.Height
		}

		// Apply final position by offsetting the entire subtree
		offsetBox(child, childX, childY)

		box.Children = append(box.Children, child)
	}

	// Position floated children (inside padding area)
	leftFloatX := innerX
	rightFloatX := innerX + innerWidth
	floatY := startY + box.Padding.Top + box.Style.BorderTopWidth

	for _, child := range floatedChildren {
		childWidth := child.Style.Width
		if childWidth <= 0 {
			childWidth = 100 // Default width for floats without explicit width
		}

		// Compute layout to determine dimensions
		computeBlockLayout(child, childWidth, 0, 0, "")

		switch child.Float {
		case "left":
			offsetBox(child, leftFloatX, floatY)
			leftFloatX += child.Rect.Width
		case "right":
			offsetBox(child, rightFloatX-child.Rect.Width, floatY)
			rightFloatX -= child.Rect.Width
		}

		box.Children = append(box.Children, child)
	}

}

// offsetBox moves a box and all its children by (dx, dy)
func offsetBox(box *LayoutBox, dx, dy float64) {
	box.Rect.X += dx
	box.Rect.Y += dy
	for _, child := range box.Children {
		offsetBox(child, dx, dy)
	}
}

// applyLineAlignment repositions inline elements based on text-align
func applyLineAlignment(lineBoxes []*LayoutBox, innerX, innerWidth float64, textAlign string) {
	if len(lineBoxes) == 0 || textAlign == "" || textAlign == "left" {
		return
	}

	// Calculate actual line width used
	lineWidth := 0.0
	for _, b := range lineBoxes {
		lineWidth += b.Rect.Width
	}

	// Calculate offset based on textAlign
	var offset float64
	switch textAlign {
	case "center":
		offset = (innerWidth - lineWidth) / 2
	case "right":
		offset = innerWidth - lineWidth
	}

	// Apply offset to all boxes
	for _, b := range lineBoxes {
		b.Rect.X += offset
	}
}

// computeInlineSize calculates the total size of an inline box from its children
func computeInlineSize(box *LayoutBox, parentTag string) (float64, float64) {
	var totalWidth float64
	var maxHeight float64

	// Use the inline element's tag if it affects font size (e.g., <small>)
	tagForSize := parentTag
	if box.Node != nil && box.Node.TagName == dom.TagSmall {
		tagForSize = dom.TagSmall
	}

	for _, child := range box.Children {
		var w, h float64
		switch child.Type {
		case TextBox:
			fontSize := getFontSize(tagForSize)
			text := css.ApplyTextTransform(child.Text, box.Style.TextTransform)

			// Check if inside a <pre> element for multi-line handling
			if isInsidePre(box) && strings.Contains(child.Text, "\n") {
				w, h = measurePreformattedText(child.Text, fontSize)
			} else {
				w = MeasureText(text, fontSize)
				h = getLineHeightFromStyle(box.Style, tagForSize)
			}
		case InlineBox:
			w, h = computeInlineSize(child, parentTag)
		case ImageBox:
			w, h = getImageSize(child.Node)
		case CheckboxBox, RadioBox:
			w = 20.0
			h = 20.0
		}
		totalWidth += w
		if h > maxHeight {
			maxHeight = h
		}
	}

	return totalWidth, maxHeight
}

// layoutInlineChildren positions children within an inline box
func layoutInlineChildren(box *LayoutBox, parentTag string) {
	// Use the inline element's tag if it affects font size (e.g., <small>)
	tagForSize := parentTag
	if box.Node != nil && box.Node.TagName == dom.TagSmall {
		tagForSize = dom.TagSmall
	}

	// Calculate vertical offset for baseline alignment
	parentLineHeight := getDefaultLineHeight(parentTag)
	childLineHeight := getLineHeightFromStyle(box.Style, tagForSize)
	baselineOffset := (parentLineHeight - childLineHeight) / 2

	offsetX := 0.0
	for _, child := range box.Children {
		switch child.Type {
		case TextBox:
			fontSize := getFontSize(tagForSize)
			text := css.ApplyTextTransform(child.Text, box.Style.TextTransform)

			var w, h float64
			// Check if inside a <pre> element for multi-line handling
			if isInsidePre(box) && strings.Contains(child.Text, "\n") {
				w, h = measurePreformattedText(child.Text, fontSize)
			} else {
				w = MeasureText(text, fontSize)
				h = getLineHeightFromStyle(box.Style, tagForSize)
			}

			child.Rect.X = box.Rect.X + offsetX
			child.Rect.Y = box.Rect.Y + baselineOffset
			child.Rect.Width = w
			child.Rect.Height = h
			offsetX += w
		case InlineBox:
			w, h := computeInlineSize(child, parentTag)
			child.Rect.X = box.Rect.X + offsetX
			child.Rect.Y = box.Rect.Y + baselineOffset
			child.Rect.Width = w
			child.Rect.Height = h
			layoutInlineChildren(child, parentTag)
			offsetX += w
		case ImageBox:
			w, h := getImageSize(child.Node)
			child.Rect.X = box.Rect.X + offsetX
			child.Rect.Y = box.Rect.Y
			child.Rect.Width = w
			child.Rect.Height = h
			offsetX += w
		case CheckboxBox, RadioBox:
			child.Rect.X = box.Rect.X + offsetX
			child.Rect.Y = box.Rect.Y
			child.Rect.Width = 20.0
			child.Rect.Height = 20.0
			offsetX += 20.0
		}
	}
}

// getCellColSpan returns the colspan value for a table cell, defaulting to 1.
func getCellColSpan(cell *LayoutBox) int {
	if cell.Node == nil || cell.Node.Attributes == nil {
		return 1
	}
	if val, ok := cell.Node.Attributes["colspan"]; ok {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			return n
		}
	}
	return 1
}

func getCellRowSpan(cell *LayoutBox) int {
	if cell.Node == nil || cell.Node.Attributes == nil {
		return 1
	}

	if val, ok := cell.Node.Attributes["rowspan"]; ok {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			return n
		}
	}
	return 1
}

// computeTableLayout handles table, row, and cell positioning
func computeTableLayout(table *LayoutBox, containerWidth float64, startX, startY float64) {
	tableWidth := containerWidth
	if table.Style.Width > 0 {
		tableWidth = table.Style.Width
	} else if table.Style.WidthPercent > 0 {
		tableWidth = containerWidth * table.Style.WidthPercent / 100.0
	} else if table.Node != nil {
		if w, ok := table.Node.Attributes["width"]; ok {
			if parsed := utils.ParseHTMLSizeAttribute(w, containerWidth); parsed > 0 {
				tableWidth = parsed
			}
		}
	}

	table.Rect.X = startX
	table.Rect.Y = startY
	table.Rect.Width = tableWidth

	// Track tbody/thead/tfoot wrappers to set their dimensions later
	var wrappers []*LayoutBox

	// Collect all rows (may be direct children or inside tbody/thead/tfoot)
	var rows []*LayoutBox
	for _, child := range table.Children {
		switch child.Type {
		case TableRowBox:
			rows = append(rows, child)
		case TableBox:
			// This is tbody/thead/tfoot - get rows from inside
			wrappers = append(wrappers, child)
			for _, grandchild := range child.Children {
				if grandchild.Type == TableRowBox {
					rows = append(rows, grandchild)
				}
			}
		}
	}

	// Count max logical columns (respecting both colspan and rowspan).
	// A cell with rowspan>1 occupies grid positions in future rows,
	// which may push those rows' cells into higher column indices.
	numCols := 0
	{
		occupied := make(map[int]map[int]bool)
		for rowIdx, row := range rows {
			colIdx := 0
			for _, cell := range row.Children {
				if cell.Type != TableCellBox {
					continue
				}

				for occupied[rowIdx] != nil && occupied[rowIdx][colIdx] {
					colIdx++
				}

				cs := getCellColSpan(cell)
				rs := getCellRowSpan(cell)

				if rs > 1 {
					for r := rowIdx + 1; r < rowIdx+rs && r < len(rows); r++ {
						if occupied[r] == nil {
							occupied[r] = make(map[int]bool)
						}

						for c := colIdx; c < colIdx+cs; c++ {
							occupied[r][c] = true
						}
					}
				}
				colIdx += cs
			}

			if colIdx > numCols {
				numCols = colIdx
			}
		}
	}

	if numCols == 0 {
		table.Rect.Height = 0
		return
	}

	cellPadding := 8.0

	// Determine per-column widths: scan all cells for explicit CSS width values.
	// For each logical column, use the maximum explicit width found across rows.
	colWidths := make([]float64, numCols)
	{
		occupied := make(map[int]map[int]bool)
		for rowIdx, row := range rows {
			colIdx := 0
			for _, cell := range row.Children {
				if cell.Type != TableCellBox {
					continue
				}
				for occupied[rowIdx] != nil && occupied[rowIdx][colIdx] {
					colIdx++
				}
				cs := getCellColSpan(cell)
				rs := getCellRowSpan(cell)
				if rs > 1 {
					for r := rowIdx + 1; r < rowIdx+rs && r < len(rows); r++ {
						if occupied[r] == nil {
							occupied[r] = make(map[int]bool)
						}
						for c := colIdx; c < colIdx+cs; c++ {
							occupied[r][c] = true
						}
					}
				}
				// Only use width from non-spanning cells for column sizing
				if cs == 1 && colIdx < numCols {
					w := cell.Style.Width
					if w == 0 && cell.Style.WidthPercent > 0 {
						w = containerWidth * cell.Style.WidthPercent / 100.0
					}
					if w > colWidths[colIdx] {
						colWidths[colIdx] = w
					}
				}
				colIdx += cs
			}
		}
	}

	// Distribute remaining space equally among auto (width=0) columns
	usedWidth := 0.0
	autoCount := 0
	for _, w := range colWidths {
		if w > 0 {
			usedWidth += w
		} else {
			autoCount++
		}
	}
	autoWidth := 0.0
	if autoCount > 0 {
		remaining := containerWidth - usedWidth
		if remaining < 0 {
			remaining = 0
		}
		autoWidth = remaining / float64(autoCount)
	}
	for i, w := range colWidths {
		if w == 0 {
			colWidths[i] = autoWidth
		}
	}

	// Precompute cumulative X offsets per column
	colXOffsets := make([]float64, numCols)
	colXOffsets[0] = 0
	for i := 1; i < numCols; i++ {
		colXOffsets[i] = colXOffsets[i-1] + colWidths[i-1]
	}

	yOffset := startY

	// Handle caption first (renders above the table rows, centered)
	for _, child := range table.Children {
		if child.Type == TableCaptionBox {
			child.Rect.X = startX
			child.Rect.Y = yOffset
			child.Rect.Width = containerWidth

			// Center the caption text
			captionHeight := 24.0
			for _, textChild := range child.Children {
				if textChild.Type == TextBox {
					fontSize := 16.0
					textWidth := MeasureText(textChild.Text, fontSize)
					textChild.Rect.X = startX + (containerWidth-textWidth)/2 // centered
					textChild.Rect.Y = yOffset
					textChild.Rect.Width = textWidth
					textChild.Rect.Height = 24.0
				}
			}
			child.Rect.Height = captionHeight
			yOffset += captionHeight + 4
		}
	}

	// Layout each row (grid-aware for rowspan support)
	type rowspanEntry struct {
		cell     *LayoutBox
		startRow int
		rowspan  int
	}
	var rowspanCells []rowspanEntry
	gridOccupied := make(map[int]map[int]bool)
	rowHeights := make([]float64, len(rows))
	cellContentH := make(map[*LayoutBox]float64)

	for rowIdx, row := range rows {
		row.Rect.X = startX
		row.Rect.Y = yOffset
		row.Rect.Width = containerWidth

		rowHeight := 24.0 // minimum row height
		colIdx := 0

		for _, cell := range row.Children {
			if cell.Type != TableCellBox {
				continue
			}

			// Skip columns reserved by rowspan cells from above
			for gridOccupied[rowIdx] != nil && gridOccupied[rowIdx][colIdx] {
				colIdx++
			}

			cs := getCellColSpan(cell)
			rs := getCellRowSpan(cell)

			// Sum widths of spanned columns
			cellWidth := 0.0
			for c := colIdx; c < colIdx+cs && c < numCols; c++ {
				cellWidth += colWidths[c]
			}
			xPos := startX
			if colIdx < numCols {
				xPos += colXOffsets[colIdx]
			}

			cell.Rect.X = xPos
			cell.Rect.Y = yOffset
			cell.Rect.Width = cellWidth

			// Compute cell content height
			cellHeight := computeCellContent(cell, cellWidth-cellPadding*2, xPos+cellPadding, yOffset+cellPadding)
			cell.Rect.Height = cellHeight + cellPadding*2
			cellContentH[cell] = cellHeight

			// Only rowspan=1 cells count toward this row's height
			if rs == 1 {
				if cell.Rect.Height > rowHeight {
					rowHeight = cell.Rect.Height
				}
			}

			// Reserve grid positions for rowspan > 1
			if rs > 1 {
				for r := rowIdx + 1; r < rowIdx+rs && r < len(rows); r++ {
					if gridOccupied[r] == nil {
						gridOccupied[r] = make(map[int]bool)
					}
					for c := colIdx; c < colIdx+cs; c++ {
						gridOccupied[r][c] = true
					}
				}
				rowspanCells = append(rowspanCells, rowspanEntry{
					cell:     cell,
					startRow: rowIdx,
					rowspan:  rs,
				})
			}

			colIdx += cs
		}

		// Set all rowspan=1 cells to same height (tallest in row)
		for _, cell := range row.Children {
			if cell.Type == TableCellBox && getCellRowSpan(cell) == 1 {
				cell.Rect.Height = rowHeight
			}
		}

		for _, cell := range row.Children {
			if cell.Type != TableCellBox || getCellRowSpan(cell) != 1 {
				continue
			}
			va := getCellVerticalAlign(cell)
			if va == "top" || va == "baseline" || va == "" {
				continue
			}
			contentHeight := cellContentH[cell]
			innerHeight := cell.Rect.Height - cellPadding*2
			var dy float64
			switch va {
			case "middle":
				dy = (innerHeight - contentHeight) / 2
			case "bottom":
				dy = innerHeight - contentHeight
			}
			if dy > 0 {
				for _, child := range cell.Children {
					shiftBoxTree(child, dy)
				}
			}
		}

		row.Rect.Height = rowHeight
		rowHeights[rowIdx] = rowHeight
		yOffset += rowHeight
	}

	// Resolve rowspan cell heights.
	// If a rowspan cell's content is taller than the combined spanned rows,
	// distribute the extra height to the last spanned row.
	needsReposition := false
	for _, rs := range rowspanCells {
		endRow := rs.startRow + rs.rowspan
		if endRow > len(rows) {
			endRow = len(rows)
		}
		totalHeight := 0.0
		for r := rs.startRow; r < endRow; r++ {
			totalHeight += rowHeights[r]
		}
		if rs.cell.Rect.Height > totalHeight {
			extra := rs.cell.Rect.Height - totalHeight
			rowHeights[endRow-1] += extra
			needsReposition = true
		}
	}

	// Reposition rows and cells if row heights changed due to rowspan overflow
	if needsReposition {
		yOffset = startY
		for _, child := range table.Children {
			if child.Type == TableCaptionBox {
				yOffset += child.Rect.Height + 4
			}
		}
		for rowIdx, row := range rows {
			row.Rect.Y = yOffset
			row.Rect.Height = rowHeights[rowIdx]
			for _, cell := range row.Children {
				if cell.Type == TableCellBox {
					cell.Rect.Y = yOffset
					if getCellRowSpan(cell) == 1 {
						cell.Rect.Height = rowHeights[rowIdx]
					}
					// Re-layout cell content at new position
					computeCellContent(cell, cell.Rect.Width-cellPadding*2, cell.Rect.X+cellPadding, yOffset+cellPadding)
				}
			}
			yOffset += rowHeights[rowIdx]
		}
	}

	// Set final heights for rowspan cells (sum of all spanned rows)
	for _, rs := range rowspanCells {
		endRow := rs.startRow + rs.rowspan
		if endRow > len(rows) {
			endRow = len(rows)
		}
		totalHeight := 0.0
		for r := rs.startRow; r < endRow; r++ {
			totalHeight += rowHeights[r]
		}
		rs.cell.Rect.Height = totalHeight
		rs.cell.Rect.Y = rows[rs.startRow].Rect.Y
	}

	table.Rect.Height = yOffset - startY

	// Set dimensions on tbody/thead/tfoot wrappers so hit testing works
	for _, wrapper := range wrappers {
		wrapper.Rect.X = startX
		wrapper.Rect.Y = startY
		wrapper.Rect.Width = containerWidth
		wrapper.Rect.Height = table.Rect.Height
	}
}

// computeCellContent layouts the content inside a table cell
func computeCellContent(cell *LayoutBox, width float64, startX, startY float64) float64 {
	currentX := startX
	currentY := startY
	lineHeight := 24.0
	maxY := startY

	var layoutInline func(box *LayoutBox)
	layoutInline = func(box *LayoutBox) {
		switch box.Type {
		case TextBox:
			fontSize := 16.0
			lines := WrapText(box.Text, fontSize, width)
			box.Rect.X = currentX
			box.Rect.Y = currentY
			if len(lines) > 1 {
				box.WrappedLines = lines
				box.Rect.Width = width
				totalHeight := float64(len(lines)) * lineHeight
				box.Rect.Height = totalHeight
				currentY += totalHeight
				currentX = startX
				if currentY > maxY {
					maxY = currentY
				}
			} else {
				textWidth := MeasureText(box.Text, fontSize)
				box.Rect.Width = textWidth
				box.Rect.Height = lineHeight
				currentX += textWidth
				if currentY+lineHeight > maxY {
					maxY = currentY + lineHeight
				}
			}

		case InlineBox:
			box.Rect.X = currentX
			box.Rect.Y = currentY
			childStartX := currentX
			for _, child := range box.Children {
				layoutInline(child)
			}
			box.Rect.Width = currentX - childStartX
			box.Rect.Height = lineHeight
			if currentY+lineHeight > maxY {
				maxY = currentY + lineHeight
			}

		case BRBox:
			currentY += lineHeight
			currentX = startX
			if currentY > maxY {
				maxY = currentY
			}

		case BlockBox:
			if currentX > startX {
				currentY += lineHeight
				currentX = startX
			}
			box.Rect.X = startX
			box.Rect.Y = currentY
			for _, child := range box.Children {
				layoutInline(child)
			}
			currentY += lineHeight
			currentX = startX
			if currentY > maxY {
				maxY = currentY
			}

		case TableBox:
			// Nested table inside cell - layout it recursively
			if currentX > startX {
				currentY += lineHeight
				currentX = startX
			}
			computeTableLayout(box, width, startX, currentY)
			currentY += box.Rect.Height
			currentX = startX
			if currentY > maxY {
				maxY = currentY
			}

		case ImageBox:
			imgW, imgH := getImageSize(box.Node)
			box.Rect.X = currentX
			box.Rect.Y = currentY
			box.Rect.Width = imgW
			box.Rect.Height = imgH
			currentX += imgW
			if currentY+imgH > maxY {
				maxY = currentY + imgH
			}

		default:
			for _, child := range box.Children {
				layoutInline(child)
			}
		}
	}

	for _, child := range cell.Children {
		layoutInline(child)
	}

	return maxY - startY
}

// getImageSize reads width/height attributes or returns defaults
func getImageSize(node *dom.Node) (float64, float64) {
	if node == nil {
		return DefaultImageWidth, DefaultImageHeight
	}

	width := DefaultImageWidth
	height := DefaultImageHeight

	if w, ok := node.Attributes["width"]; ok {
		if parsed := utils.ParseHTMLSizeAttribute(w, 0); parsed > 0 {
			width = parsed
		}
	}

	if h, ok := node.Attributes["height"]; ok {
		if parsed := utils.ParseHTMLSizeAttribute(h, 0); parsed > 0 {
			height = parsed
		}
	}

	return width, height
}

func (box *LayoutBox) Print(indent int) {
	prefix := strings.Repeat("  ", indent)

	typeName := "Block"
	switch box.Type {
	case InlineBox:
		typeName = "Inline"
	case TextBox:
		typeName = "Text"
	case ImageBox:
		typeName = "Image"
	}

	if box.Type == TextBox {
		fmt.Printf("%s[%s] \"%s\" (%.0f,%.0f) %.0fx%.0f\n",
			prefix, typeName, truncate(box.Text, 20),
			box.Rect.X, box.Rect.Y,
			box.Rect.Width, box.Rect.Height)
	} else {
		tag := ""
		if box.Node != nil && box.Node.TagName != "" {
			tag = "<" + box.Node.TagName + "> "
		}
		fmt.Printf("%s[%s] %s(%.0f,%.0f) %.0fx%.0f\n",
			prefix, typeName, tag,
			box.Rect.X, box.Rect.Y,
			box.Rect.Width, box.Rect.Height)
	}

	for _, child := range box.Children {
		child.Print(indent + 1)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func isInsidePre(box *LayoutBox) bool {
	for p := box.Parent; p != nil; p = p.Parent {
		if p.Node != nil && p.Node.TagName == dom.TagPre {
			return true
		}
	}
	return false
}

// measurePreformattedText calculates width and height for multi-line text inside <pre>
func measurePreformattedText(text string, fontSize float64) (width, height float64) {
	// Expand tabs to spaces for proper alignment
	text = dom.ExpandTabs(text, 8)
	lines := strings.Split(text, "\n")
	lineHeight := fontSize * 1.5

	// Find widest line
	maxWidth := 0.0
	for _, line := range lines {
		lw := MeasureText(line, fontSize)
		if lw > maxWidth {
			maxWidth = lw
		}
	}

	return maxWidth, float64(len(lines)) * lineHeight
}

// getButtonText extracts text content from a button element
func getButtonText(box *LayoutBox) string {
	for _, child := range box.Children {
		if child.Type == TextBox {
			return child.Text
		}
	}
	if box.Node != nil {
		if val, ok := box.Node.Attributes["value"]; ok {
			return val
		}
	}
	return "Button"
}

// GetLegendText extracts text content from a legend element
func GetLegendText(box *LayoutBox) string {
	for _, child := range box.Children {
		if child.Type == TextBox {
			return child.Text
		}
	}
	return ""
}

func getLineHeightFromStyle(style css.Style, tagName string) float64 {
	if style.LineHeight > 0 {
		return style.LineHeight
	}
	return getDefaultLineHeight(tagName)
}

func getCellVerticalAlign(cell *LayoutBox) string {
	if cell.Style.VerticalAlign != "" {
		return cell.Style.VerticalAlign
	}
	return "top"
}

func shiftBoxTree(box *LayoutBox, dy float64) {
	box.Rect.Y += dy
	for _, child := range box.Children {
		shiftBoxTree(child, dy)
	}
}
