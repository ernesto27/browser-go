package render

import (
	"browser/css"
	"browser/dom"
	"browser/layout"
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Colors are defined in colors.go

// LinkStyler provides functions for styling links
type LinkStyler struct {
	IsVisited  func(url string) bool
	ResolveURL func(href string) string
}

// Font sizes
var (
	SizeH1     float32 = 32
	SizeH2     float32 = 24
	SizeH3     float32 = 18
	SizeH4     float32 = 16
	SizeH5     float32 = 14
	SizeH6     float32 = 12
	SizeNormal float32 = 16
	SizeSmall  float32 = 12
)

// Text decoration constants
const (
	TextDecorationNone            = ""
	TextDecorationUnderline       = "underline"
	TextDecorationDottedUnderline = "dotted-underline"
	TextDecorationLineThrough     = "line-through"
)

// TextStyle holds inherited text styling
type TextStyle struct {
	Color          color.Color
	Size           float32
	Bold           bool
	Italic         bool
	Monospace      bool
	FontFamily     []string
	TextDecoration string
	TextTransform  string

	Opacity       float64
	Visibility    string
	LetterSpacing float64
	WordSpacing   float64
	LineHeight    float64
}

type DrawInput struct {
	layout.Rect
	Placeholder string
	Value       string
	InputType   string // text, password, email, number, etc.
	IsFocused   bool
	IsDisabled  bool
	IsReadonly  bool
	IsValid     bool // For validation feedback (email, etc.)
}

type DrawButton struct {
	layout.Rect
	Text       string
	IsDisabled bool
}

type DrawTextarea struct {
	layout.Rect
	Placeholder string
	Value       string
	IsFocused   bool
	IsDisabled  bool
	IsReadonly  bool
}

type DrawSelect struct {
	layout.Rect
	Options       []string // List of option texts
	SelectedValue string   // Currently selected value
	IsOpen        bool     // Is dropdown open?
	IsDisabled    bool
	IsReadonly    bool
}

type DrawRadio struct {
	layout.Rect
	IsChecked  bool
	IsDisabled bool
}

type DrawCheckbox struct {
	layout.Rect
	IsChecked  bool
	IsDisabled bool
	IsReadonly bool
}

// InputState holds all interactive form state for rendering
type InputState struct {
	InputValues     map[*dom.Node]string // Text input values
	FocusedNode     *dom.Node            // Currently focused input/textarea
	OpenSelectNode  *dom.Node            // Which select dropdown is open
	RadioValues     map[string]*dom.Node // Selected radio per group (key: name attr)
	CheckboxValues  map[*dom.Node]bool   // Checked state per check
	FileInputValues map[*dom.Node]string // Selected filename per file input
	InvalidNodes    map[*dom.Node]bool   // Nodes with invalid input

	SelectionStart *SelectionPoint
	SelectionEnd   *SelectionPoint
}

// isTextSelected checks if a text box is within the current selection range
func isTextSelected(box *layout.LayoutBox, state InputState) bool {
	if state.SelectionStart == nil || state.SelectionEnd == nil {
		return false
	}

	// Calculate selection bounds (handle reverse selection - dragging up)
	minY := state.SelectionStart.Y
	maxY := state.SelectionEnd.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}

	minX := state.SelectionStart.X
	maxX := state.SelectionEnd.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}

	boxTop := box.Rect.Y
	boxBottom := box.Rect.Y + box.Rect.Height
	boxLeft := box.Rect.X
	boxRight := box.Rect.X + box.Rect.Width

	// Check if box overlaps selection area
	verticalOverlap := boxBottom >= minY && boxTop <= maxY
	horizontalOverlap := boxRight >= minX && boxLeft <= maxX

	// For single-line selection, require both overlaps
	// For multi-line, just vertical overlap is enough for middle lines
	if maxY-minY < box.Rect.Height {
		// Single line selection - need both overlaps
		return verticalOverlap && horizontalOverlap
	}

	// Multi-line selection
	return verticalOverlap
}

// DefaultStyle returns the default text style
func DefaultStyle() TextStyle {
	return TextStyle{
		Color:      ColorBlack,
		Size:       SizeNormal,
		Bold:       false,
		Italic:     false,
		Opacity:    1.0,
		LineHeight: float64(SizeNormal) * 1.2,
	}
}

// applyOpacity returns a color with opacity applied to alpha channel
func applyOpacity(c color.Color, opacity float64) color.Color {
	if opacity >= 1.0 {
		return c
	}
	r, g, b, a := c.RGBA()
	newAlpha := uint8(float64(a>>8) * opacity)
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), newAlpha}
}

type DisplayCommand any

type DrawRect struct {
	layout.Rect
	Color             color.Color
	CornerRadius      float64
	TopLeftRadius     float64
	TopRightRadius    float64
	BottomRightRadius float64
	BottomLeftRadius  float64
}

type DrawText struct {
	Text            string
	X, Y            float64
	Width           float64
	LetterSpacing   float64
	WordSpacing     float64
	Color           color.Color
	Size            float32
	Bold            bool
	Italic          bool
	Monospace       bool
	Underline       bool
	DottedUnderline bool
	Strikethrough   bool
	TextTransform   string
}

type DrawImage struct {
	layout.Rect
	URL            string
	AltText        string
	ReferrerPolicy string
	Node           *dom.Node
}

type DrawHR struct {
	layout.Rect
}

type DrawFileInput struct {
	layout.Rect
	Filename   string
	IsDisabled bool
}

type DrawFieldset struct {
	layout.Rect
	LegendX      float64
	LegendY      float64
	LegendWidth  float64
	LegendHeight float64
	LegendText   string
	HasLegend    bool
}

type paintLayer int

const (
	paintAll paintLayer = iota
	paintNormalOnly
	paintFixedOnly
)

func BuildDisplayList(root *layout.LayoutBox, state InputState, linkStyler LinkStyler) []DisplayCommand {
	normal, fixed := BuildDisplayLayers(root, state, linkStyler)
	return append(normal, fixed...)
}

func BuildDisplayLayers(root *layout.LayoutBox, state InputState, linkStyler LinkStyler) ([]DisplayCommand, []DisplayCommand) {
	var normalCommands []DisplayCommand
	var fixedCommands []DisplayCommand
	var commands []DisplayCommand

	// Calculate actual content height from layout tree
	contentHeight := root.Rect.Y + root.Rect.Height
	if contentHeight < 600 {
		contentHeight = 600 // Minimum height
	}

	commands = append(commands, DrawRect{
		Rect:  layout.Rect{X: 0, Y: 0, Width: 3000, Height: contentHeight}, // Wide enough for most screens
		Color: color.White,
	})

	normalCommands = append(normalCommands, commands...)
	paintLayoutBox(root, &normalCommands, DefaultStyle(), state, linkStyler, paintNormalOnly, false)
	paintLayoutBox(root, &fixedCommands, DefaultStyle(), state, linkStyler, paintFixedOnly, false)

	return normalCommands, fixedCommands
}

func paintLayoutBox(box *layout.LayoutBox, commands *[]DisplayCommand, style TextStyle, state InputState, linkStyler LinkStyler, layer paintLayer, ancestorFixed bool) {
	currentStyle := style
	isFixed := ancestorFixed || box.Position == "fixed"

	if layer == paintNormalOnly && isFixed {
		return
	}
	if layer == paintFixedOnly && !isFixed {
		// Skip drawing this non-fixed box, but still traverse children to find fixed descendants.
		if box.Type != layout.ButtonBox && box.Type != layout.SelectBox {
			for _, child := range box.Children {
				if child.Type == layout.LegendBox {
					continue
				}
				paintLayoutBox(child, commands, currentStyle, state, linkStyler, layer, isFixed)
			}
		}
		return
	}

	// Apply inline styles from CSS
	if box.Style.Color != nil {
		currentStyle.Color = box.Style.Color
	}
	if box.Style.FontSize > 0 {
		currentStyle.Size = float32(box.Style.FontSize)
		if box.Style.LineHeight == 0 {
			currentStyle.LineHeight = box.Style.FontSize * 1.2
		}
	}
	if box.Style.Bold {
		currentStyle.Bold = true
	}
	if box.Style.Italic {
		currentStyle.Italic = true
	}

	if len(box.Style.FontFamily) > 0 {
		currentStyle.FontFamily = box.Style.FontFamily
	}

	if box.Style.TextDecoration != "" {
		currentStyle.TextDecoration = box.Style.TextDecoration
	}
	if box.Style.TextTransform != "" {
		currentStyle.TextTransform = box.Style.TextTransform
	}
	if box.Style.LetterSpacingSet {
		currentStyle.LetterSpacing = box.Style.LetterSpacing
	}
	if box.Style.WordSpacingSet {
		currentStyle.WordSpacing = box.Style.WordSpacing
	}

	if box.Style.LineHeight > 0 {
		currentStyle.LineHeight = box.Style.LineHeight
	}

	if box.Style.Opacity > 0 {
		currentStyle.Opacity = box.Style.Opacity
	}
	if box.Style.Visibility != "" {
		currentStyle.Visibility = box.Style.Visibility
	}

	isHidden := currentStyle.Visibility == "hidden"

	// Draw background if set
	if box.Style.BackgroundColor != nil && !isHidden {
		tl := box.Style.BorderTopLeftRadius
		tr := box.Style.BorderTopRightRadius
		br := box.Style.BorderBottomRightRadius
		bl := box.Style.BorderBottomLeftRadius
		if tl == 0 && tr == 0 && br == 0 && bl == 0 {
			tl = box.Style.BorderRadius
			tr = box.Style.BorderRadius
			br = box.Style.BorderRadius
			bl = box.Style.BorderRadius
		}
		*commands = append(*commands, DrawRect{
			Rect:              box.Rect,
			Color:             applyOpacity(box.Style.BackgroundColor, currentStyle.Opacity),
			CornerRadius:      box.Style.BorderRadius,
			TopLeftRadius:     tl,
			TopRightRadius:    tr,
			BottomRightRadius: br,
			BottomLeftRadius:  bl,
		})
	}

	if box.Style.BackgroundImage != "" && !isHidden {
		*commands = append(*commands, DrawImage{
			Rect: box.Rect,
			URL:  box.Style.BackgroundImage,
		})
	}

	// Draw borders if set
	if !isHidden {
		if box.Style.BorderTopWidth > 0 && box.Style.BorderTopStyle != "none" && box.Style.BorderTopColor != nil {
			*commands = append(*commands, DrawRect{
				Rect:  layout.Rect{X: box.Rect.X, Y: box.Rect.Y, Width: box.Rect.Width, Height: box.Style.BorderTopWidth},
				Color: applyOpacity(box.Style.BorderTopColor, currentStyle.Opacity),
			})
		}
		if box.Style.BorderBottomWidth > 0 && box.Style.BorderBottomStyle != "none" && box.Style.BorderBottomColor != nil {
			*commands = append(*commands, DrawRect{
				Rect:  layout.Rect{X: box.Rect.X, Y: box.Rect.Y + box.Rect.Height - box.Style.BorderBottomWidth, Width: box.Rect.Width, Height: box.Style.BorderBottomWidth},
				Color: applyOpacity(box.Style.BorderBottomColor, currentStyle.Opacity),
			})
		}
		if box.Style.BorderLeftWidth > 0 && box.Style.BorderLeftStyle != "none" && box.Style.BorderLeftColor != nil {
			*commands = append(*commands, DrawRect{
				Rect:  layout.Rect{X: box.Rect.X, Y: box.Rect.Y, Width: box.Style.BorderLeftWidth, Height: box.Rect.Height},
				Color: applyOpacity(box.Style.BorderLeftColor, currentStyle.Opacity),
			})
		}
		if box.Style.BorderRightWidth > 0 && box.Style.BorderRightStyle != "none" && box.Style.BorderRightColor != nil {
			*commands = append(*commands, DrawRect{
				Rect:  layout.Rect{X: box.Rect.X + box.Rect.Width - box.Style.BorderRightWidth, Y: box.Rect.Y, Width: box.Style.BorderRightWidth, Height: box.Rect.Height},
				Color: applyOpacity(box.Style.BorderRightColor, currentStyle.Opacity),
			})
		}
	}

	// Apply tag-based styles
	if box.Node != nil {
		switch box.Node.TagName {
		case dom.TagH1:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH1
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagH2:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH2
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagH3:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH3
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagH4:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH4
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagH5:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH5
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagH6:
			if box.Style.FontSize == 0 {
				currentStyle.Size = SizeH6
			}
			if !box.Style.Bold {
				currentStyle.Bold = true
			}
		case dom.TagA:
			// Link color and text-decoration are now handled via CSS cascade
			// (UA defaults in applyUserAgentDefaults, overridable by user CSS rules)
		case dom.TagStrong, dom.TagB:
			currentStyle.Bold = true
		case dom.TagEm, dom.TagI, dom.TagCite, dom.TagDnf:
			currentStyle.Italic = true
		case dom.TagAbbr:
			currentStyle.TextDecoration = TextDecorationDottedUnderline
		case dom.TagSmall:
			currentStyle.Size = SizeSmall
		case dom.TagU:
			currentStyle.TextDecoration = TextDecorationUnderline
		case dom.TagDel:
			currentStyle.TextDecoration = TextDecorationLineThrough
		case dom.TagS:
			currentStyle.TextDecoration = TextDecorationLineThrough
		case dom.TagIns:
			currentStyle.TextDecoration = TextDecorationUnderline
		case dom.TagPre:
			currentStyle.Monospace = true
			if box.Style.BackgroundColor == nil && !isHidden {
				*commands = append(*commands, DrawRect{
					Rect:  box.Rect,
					Color: color.RGBA{245, 245, 245, 255},
				})
			}
		case dom.TagTH:
			currentStyle.Bold = true
		case dom.TagMark:
			currentStyle.Color = color.RGBA{0, 0, 0, 255}
			if !isHidden {
				*commands = append(*commands, DrawRect{
					Rect:  box.Rect,
					Color: color.RGBA{255, 255, 0, 255},
				})
			}

		}
	}

	// Draw text
	if box.Type == layout.TextBox && box.Text != "" && !isHidden {
		if isTextSelected(box, state) {
			*commands = append(*commands, DrawRect{
				Rect:  box.Rect,
				Color: color.RGBA{0, 120, 215, 128}, // Selection blue
			})

		}

		text := css.ApplyTextTransform(box.Text, currentStyle.TextTransform)

		if isListItem, isOrdered, index, listType := getListInfo(box); isListItem {
			if isOrdered {
				text = formatListMarker(index, listType) + " " + text
			} else {
				text = "• " + text
			}
		}

		if currentStyle.Monospace && strings.Contains(text, "\n") {
			// Expand tabs to spaces for proper alignment
			text = dom.ExpandTabs(text, 8)
			lines := strings.Split(text, "\n")
			lineHeight := float64(currentStyle.Size) * 1.5
			y := box.Rect.Y
			for _, line := range lines {
				*commands = append(*commands, DrawText{
					Text: line, X: box.Rect.X, Y: y, Width: box.Rect.Width,
					LetterSpacing:   currentStyle.LetterSpacing,
					WordSpacing:     currentStyle.WordSpacing,
					Size:            currentStyle.Size,
					Color:           applyOpacity(currentStyle.Color, currentStyle.Opacity),
					Bold:            currentStyle.Bold,
					Italic:          currentStyle.Italic,
					Monospace:       currentStyle.Monospace || fontStackHasMonospace(currentStyle.FontFamily),
					Underline:       currentStyle.TextDecoration == TextDecorationUnderline,
					DottedUnderline: currentStyle.TextDecoration == TextDecorationDottedUnderline,
					Strikethrough:   currentStyle.TextDecoration == TextDecorationLineThrough,
				})
				y += lineHeight
			}
		} else if len(box.WrappedLines) > 1 {
			// Render wrapped lines
			lineHeight := currentStyle.LineHeight
			y := box.Rect.Y
			for _, line := range box.WrappedLines {
				transformedLine := css.ApplyTextTransform(line, currentStyle.TextTransform)
				*commands = append(*commands, DrawText{
					Text: transformedLine, X: box.Rect.X, Y: y, Width: box.Rect.Width,
					LetterSpacing:   currentStyle.LetterSpacing,
					WordSpacing:     currentStyle.WordSpacing,
					Size:            currentStyle.Size,
					Color:           applyOpacity(currentStyle.Color, currentStyle.Opacity),
					Bold:            currentStyle.Bold,
					Italic:          currentStyle.Italic,
					Monospace:       currentStyle.Monospace || fontStackHasMonospace(currentStyle.FontFamily),
					Underline:       currentStyle.TextDecoration == TextDecorationUnderline,
					DottedUnderline: currentStyle.TextDecoration == TextDecorationDottedUnderline,
					Strikethrough:   currentStyle.TextDecoration == TextDecorationLineThrough,
				})
				y += lineHeight
			}
		} else {
			*commands = append(*commands, DrawText{
				Text: text, X: box.Rect.X, Y: box.Rect.Y, Width: box.Rect.Width,
				LetterSpacing:   currentStyle.LetterSpacing,
				WordSpacing:     currentStyle.WordSpacing,
				Size:            currentStyle.Size,
				Color:           applyOpacity(currentStyle.Color, currentStyle.Opacity),
				Bold:            currentStyle.Bold,
				Italic:          currentStyle.Italic,
				Monospace:       currentStyle.Monospace || fontStackHasMonospace(currentStyle.FontFamily),
				Underline:       currentStyle.TextDecoration == TextDecorationUnderline,
				DottedUnderline: currentStyle.TextDecoration == TextDecorationDottedUnderline,
				Strikethrough:   currentStyle.TextDecoration == TextDecorationLineThrough,
			})
		}
	}

	// Draw image
	if box.Type == layout.ImageBox && box.Node != nil && !isHidden {
		if src := box.Node.Attributes["src"]; src != "" {
			*commands = append(*commands, DrawImage{
				Rect:           box.Rect,
				URL:            src,
				AltText:        box.Node.Attributes["alt"],
				ReferrerPolicy: box.Node.Attributes["referrerpolicy"],
				Node:           box.Node,
			})
		}
	}

	if box.Type == layout.HRBox && !isHidden {
		*commands = append(*commands, DrawHR{
			Rect: box.Rect,
		})
	}

	// Input with state - use DOM node for lookup (stable across reflow)
	if box.Type == layout.InputBox && box.Node != nil && !isHidden {
		value := state.InputValues[box.Node]
		isFocused := (box.Node == state.FocusedNode)

		placeholder := box.Node.Attributes["placeholder"]
		if placeholder == "" {
			placeholder = box.Node.Attributes["value"]
		}

		inputType := strings.ToLower(box.Node.Attributes["type"])
		if inputType == "" {
			inputType = "text"
		}

		_, isDisabled := box.Node.Attributes["disabled"]
		_, isReadonly := box.Node.Attributes["readonly"]

		if isDisabled {
			isFocused = false
		}

		// Validate based on input type
		isValid := true
		if inputType == "email" && value != "" {
			isValid = isValidEmail(value)
		}

		if state.InvalidNodes != nil && state.InvalidNodes[box.Node] {
			isValid = false
		}

		*commands = append(*commands, DrawInput{
			Rect:        box.Rect,
			Placeholder: placeholder,
			Value:       value,
			InputType:   inputType,
			IsFocused:   isFocused,
			IsDisabled:  isDisabled,
			IsReadonly:  isReadonly,
			IsValid:     isValid,
		})
	}

	if box.Type == layout.ButtonBox && !isHidden {
		*commands = append(*commands, DrawButton{
			Rect: box.Rect,
			Text: getButtonTextFromBox(box),
		})
	}

	if box.Type == layout.TextareaBox && box.Node != nil && !isHidden {
		value := state.InputValues[box.Node]
		isFocused := (box.Node == state.FocusedNode)

		_, isDisabled := box.Node.Attributes["disabled"]
		_, isReadonly := box.Node.Attributes["readonly"]

		if isDisabled {
			isFocused = false
		}

		*commands = append(*commands, DrawTextarea{
			Rect:        box.Rect,
			Placeholder: box.Node.Attributes["placeholder"],
			Value:       value,
			IsFocused:   isFocused,
			IsDisabled:  isDisabled,
			IsReadonly:  isReadonly,
		})
	}

	if box.Type == layout.SelectBox && box.Node != nil && !isHidden {
		// Get options from <option> children
		var options []string
		fmt.Printf("Select box found, children: %d\n", len(box.Node.Children))
		for _, child := range box.Node.Children {
			fmt.Printf("  Child: TagName=%s, Type=%d\n", child.TagName, child.Type)
			if child.TagName == "option" {
				for _, textNode := range child.Children {
					fmt.Printf("    TextNode: Type=%d, Text=%q\n", textNode.Type, textNode.Text)
					if textNode.Type == dom.Text {
						options = append(options, textNode.Text)
						break
					}
				}
			}
		}

		selectedValue := state.InputValues[box.Node]
		_, isDisabled := box.Node.Attributes["disabled"]

		// Disabled selects cannot be open
		isOpen := (box.Node == state.OpenSelectNode) && !isDisabled
		fmt.Printf("Select: options=%v, isOpen=%v, openSelectNode=%p, box.Node=%p\n", options, isOpen, state.OpenSelectNode, box.Node)

		*commands = append(*commands, DrawSelect{
			Rect:          box.Rect,
			Options:       options,
			SelectedValue: selectedValue,
			IsOpen:        isOpen,
			IsDisabled:    isDisabled,
		})
	}

	// Radio button
	if box.Type == layout.RadioBox && box.Node != nil && !isHidden {
		name := box.Node.Attributes["name"]
		isChecked := false
		if name != "" && state.RadioValues != nil {
			isChecked = (state.RadioValues[name] == box.Node)
		}
		// Fallback to HTML checked attribute if no runtime state
		if !isChecked {
			_, isChecked = box.Node.Attributes["checked"]
		}

		_, isDisabled := box.Node.Attributes["disabled"]

		*commands = append(*commands, DrawRadio{
			Rect:       box.Rect,
			IsChecked:  isChecked,
			IsDisabled: isDisabled,
		})
	}

	if box.Type == layout.CheckboxBox && box.Node != nil && !isHidden {
		isChecked := false
		if state.CheckboxValues != nil {
			isChecked = state.CheckboxValues[box.Node]
		}

		_, isDisabled := box.Node.Attributes["disabled"]
		*commands = append(*commands, DrawCheckbox{
			Rect:       box.Rect,
			IsChecked:  isChecked,
			IsDisabled: isDisabled,
		})
	}

	if box.Type == layout.FieldsetBox {
		var legendX, legendY, legendWidth, legendHeight float64
		var legendText string
		var hasLegend bool

		for _, child := range box.Children {
			if child.Type == layout.LegendBox {
				hasLegend = true
				legendX = child.Rect.X
				legendY = child.Rect.Y
				legendWidth = child.Rect.Width
				legendHeight = child.Rect.Height
				legendText = layout.GetLegendText(child)
				break
			}
		}

		*commands = append(*commands, DrawFieldset{
			Rect:         box.Rect,
			LegendX:      legendX,
			LegendY:      legendY,
			LegendWidth:  legendWidth,
			LegendHeight: legendHeight,
			LegendText:   legendText,
			HasLegend:    hasLegend,
		})
	}

	if box.Type == layout.FileInputBox && box.Node != nil && !isHidden {
		filename := state.FileInputValues[box.Node]
		_, isDisabled := box.Node.Attributes["disabled"]

		*commands = append(*commands, DrawFileInput{
			Rect:       box.Rect,
			Filename:   filename,
			IsDisabled: isDisabled,
		})
	}

	// Draw table cell border (only when table has border attribute > 0)
	if box.Type == layout.TableCellBox && box.TableBorder > 0 {
		borderColor := color.Gray{Y: 180}
		bw := float64(box.TableBorder)
		*commands = append(*commands, DrawRect{Rect: layout.Rect{X: box.Rect.X, Y: box.Rect.Y, Width: box.Rect.Width, Height: bw}, Color: borderColor})
		*commands = append(*commands, DrawRect{Rect: layout.Rect{X: box.Rect.X, Y: box.Rect.Y + box.Rect.Height - bw, Width: box.Rect.Width, Height: bw}, Color: borderColor})
		*commands = append(*commands, DrawRect{Rect: layout.Rect{X: box.Rect.X, Y: box.Rect.Y, Width: bw, Height: box.Rect.Height}, Color: borderColor})
		*commands = append(*commands, DrawRect{Rect: layout.Rect{X: box.Rect.X + box.Rect.Width - bw, Y: box.Rect.Y, Width: bw, Height: box.Rect.Height}, Color: borderColor})
	}

	// Paint children with input state
	// Skip children for elements that render their own content
	if box.Type != layout.ButtonBox && box.Type != layout.SelectBox {
		for _, child := range box.Children {
			// Skip LegendBox - DrawFieldset already renders the legend text
			if child.Type == layout.LegendBox {
				continue
			}
			paintLayoutBox(child, commands, currentStyle, state, linkStyler, layer, isFixed)
		}
	}
}

// getListInfo returns (isListItem, isOrdered, itemIndex)
func getListInfo(box *layout.LayoutBox) (bool, bool, int, string) {
	// Check if parent is <li>
	if box.Parent == nil || box.Parent.Node == nil {
		return false, false, 0, ""
	}
	if box.Parent.Node.TagName != dom.TagLI {
		return false, false, 0, ""
	}

	li := box.Parent

	// Check if grandparent is <ul> or <ol>
	if li.Parent == nil || li.Parent.Node == nil {
		return false, false, 0, ""
	}

	listTag := li.Parent.Node.TagName
	if listTag != dom.TagUL && listTag != dom.TagOL && listTag != dom.TagMenu {
		return false, false, 0, ""
	}

	isOrdered := listTag == dom.TagOL

	listType := "1"
	if typeAttr, ok := li.Parent.Node.Attributes["type"]; ok {
		listType = typeAttr
	}

	_, isReversed := li.Parent.Node.Attributes["reversed"]

	// Count total items first (needed for reversed default start)
	totalItems := 0
	for _, sibling := range li.Parent.Children {
		if sibling.Node != nil && sibling.Node.TagName == dom.TagLI {
			totalItems++
		}
	}

	// Determine starting ordinal per WHATWG spec
	ordinal := 1
	if startAttr, ok := li.Parent.Node.Attributes["start"]; ok {
		if parsed, err := strconv.Atoi(startAttr); err == nil {
			ordinal = parsed
		}
	} else if isReversed {
		ordinal = totalItems
	}

	// Calculate ordinal for current li, respecting value attrs and reversed
	currentOrdinal := ordinal
	for _, sibling := range li.Parent.Children {
		if sibling.Node != nil && sibling.Node.TagName == dom.TagLI {
			// Check if this li has a value attribute
			if valAttr, ok := sibling.Node.Attributes["value"]; ok {
				if parsed, err := strconv.Atoi(valAttr); err == nil {
					ordinal = parsed
				}
			}

			if sibling == li {
				currentOrdinal = ordinal
				break
			}

			// Move to next ordinal
			if isReversed {
				ordinal--
			} else {
				ordinal++
			}
		}
	}

	return true, isOrdered, currentOrdinal, listType
}

func formatListMarker(index int, listType string) string {
	switch listType {
	case "a":
		return string(rune('a'+index-1)) + "."
	case "A":
		return string(rune('A'+index-1)) + "."
	case "i":
		return toRomanLower(index) + "."
	case "I":
		return toRomanUpper(index) + "."
	default:
		return fmt.Sprintf("%d.", index)
	}
}

func toRomanLower(n int) string {
	return strings.ToLower(toRomanUpper(n))
}

func toRomanUpper(n int) string {
	if n <= 0 || n > 3999 {
		return fmt.Sprintf("%d", n)
	}
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var result strings.Builder
	for i, v := range vals {
		for n >= v {
			result.WriteString(syms[i])
			n -= v
		}
	}
	return result.String()
}

func getButtonTextFromBox(box *layout.LayoutBox) string {
	for _, child := range box.Children {
		if child.Type == layout.TextBox {
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

// isValidEmail checks if the value is a valid email format
func isValidEmail(value string) bool {
	// Simple validation: must have @ with text before and after
	atIndex := strings.Index(value, "@")
	if atIndex < 1 {
		return false
	}
	// Must have something after @
	afterAt := value[atIndex+1:]
	if len(afterAt) < 1 {
		return false
	}
	// Must have a dot after @ (for domain)
	dotIndex := strings.Index(afterAt, ".")
	if dotIndex < 1 || dotIndex >= len(afterAt)-1 {
		return false
	}
	return true
}

// fontStackHasMonospace checks if any font in the stack is a monospace font
func fontStackHasMonospace(fonts []string) bool {
	for _, font := range fonts {
		f := strings.ToLower(font)
		if f == "monospace" || f == "courier" || f == "courier new" ||
			f == "consolas" || f == "monaco" || f == "menlo" {
			return true
		}
	}
	return false
}
