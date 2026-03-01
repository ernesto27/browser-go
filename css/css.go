package css

import (
	"browser/dom"
	"image/color"
	"strconv"
	"strings"
)

const (
	DefaultFontSize       = 16.0
	DefaultViewportWidth  = 0.0
	DefaultViewportHeight = 0.0

	ListStyleNone       = "none"
	ListStyleDisc       = "disc"
	ListStyleCircle     = "circle"
	ListStyleSquare     = "square"
	ListStyleDecimal    = "decimal"
	ListStyleLowerAlpha = "lower-alpha"
	ListStyleLowerLatin = "lower-latin"
	ListStyleUpperAlpha = "upper-alpha"
	ListStyleUpperLatin = "upper-latin"
	ListStyleLowerRoman = "lower-roman"
	ListStyleUpperRoman = "upper-roman"

	ListMarkerNumeric    = "1"
	ListMarkerLowerAlpha = "a"
	ListMarkerUpperAlpha = "A"
	ListMarkerLowerRoman = "i"
	ListMarkerUpperRoman = "I"
)

type Style struct {
	Color            color.Color
	BackgroundColor  color.Color
	BackgroundImage  string
	FontSize         float64
	FontVariant      string
	LineHeight       float64
	Bold             bool
	Italic           bool
	MarginTop        float64
	MarginBottom     float64
	MarginLeft       float64
	MarginRight      float64
	MarginLeftAuto   bool
	MarginRightAuto  bool
	PaddingTop       float64
	PaddingBottom    float64
	PaddingLeft      float64
	PaddingRight     float64
	TextAlign        string
	VerticalAlign    string
	Display          string
	Float            string
	Position         string
	Top              float64
	Left             float64
	Right            float64
	Bottom           float64
	TextDecoration   string
	Opacity          float64
	Visibility       string
	Cursor           string
	TextTransform    string
	LetterSpacing    float64
	LetterSpacingSet bool
	WordSpacing      float64
	WordSpacingSet   bool
	Width            float64
	WidthPercent     float64 // percentage width (e.g., 25 means 25%)
	Height           float64
	MinWidth         float64
	MaxWidth         float64
	MinHeight        float64
	MaxHeight        float64
	FontFamily       []string

	// Border properties
	BorderTopWidth          float64
	BorderRightWidth        float64
	BorderBottomWidth       float64
	BorderLeftWidth         float64
	BorderTopColor          color.Color
	BorderRightColor        color.Color
	BorderBottomColor       color.Color
	BorderLeftColor         color.Color
	BorderTopStyle          string
	BorderRightStyle        string
	BorderBottomStyle       string
	BorderLeftStyle         string
	BorderRadius            float64
	BorderTopLeftRadius     float64
	BorderTopRightRadius    float64
	BorderBottomLeftRadius  float64
	BorderBottomRightRadius float64

	TopSet    bool
	LeftSet   bool
	RightSet  bool
	BottomSet bool

	ListStyleType string
}

func DefaultStyle() Style {
	return Style{
		FontSize: DefaultFontSize,
		Bold:     false,
		Italic:   false,
		Opacity:  1.0,
	}
}

// stripImportant checks if a CSS value ends with !important,
// returns the clean value and whether !important was present
func stripImportant(value string) (cleanValue string, isImportant bool) {
	lowerValue := strings.ToLower(value)
	if strings.HasSuffix(lowerValue, "!important") {
		return strings.TrimSpace(value[:len(value)-10]), true
	}
	return value, false
}

func ParseInlineStyle(styleAttr string) Style {
	style := DefaultStyle()
	importantProps := make(map[string]bool) // Track !important properties

	parts := strings.Split(styleAttr, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}

		property := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// Check for !important flag
		value, important := stripImportant(value)

		// Skip if property was set with !important and new value is not
		if importantProps[property] && !important {
			continue
		}

		applyDeclaration(&style, property, value)

		if important {
			importantProps[property] = true
		}
	}
	return style
}

// parseBorderShorthand parses "1px solid black" into width, style, color
func parseBorderShorthand(value string) (float64, string, color.Color) {
	parts := strings.Fields(value)
	var width float64
	var borderStyle string
	var borderColor color.Color
	for _, part := range parts {
		if w := ParseSize(part); w > 0 {
			width = w
		} else if part == "solid" || part == "dashed" || part == "dotted" || part == "none" {
			borderStyle = part
		} else if c := ParseColor(part); c != nil {
			borderColor = c
		}
	}
	return width, borderStyle, borderColor
}

func ParseSize(value string) float64 {
	return ParseSizeWithContext(value, DefaultFontSize, DefaultViewportWidth, DefaultViewportHeight)
}

func ParseSizeWithContext(value string, baseFontSize float64, viewportWidth, viewportHeight float64) float64 {
	value = strings.TrimSpace(strings.ToLower(value))

	if strings.HasSuffix(value, "vh") {
		num := strings.TrimSuffix(value, "vh")
		if percent, err := strconv.ParseFloat(num, 64); err == nil {
			return (percent / 100.0) * viewportHeight
		}
		return 0
	}

	if strings.HasSuffix(value, "vw") {
		num := strings.TrimSuffix(value, "vw")
		if percent, err := strconv.ParseFloat(num, 64); err == nil {
			return (percent / 100.0) * viewportWidth
		}
		return 0
	}

	// Handle em units
	if strings.HasSuffix(value, "em") {
		num := strings.TrimSuffix(value, "em")
		if multiplier, err := strconv.ParseFloat(num, 64); err == nil {
			return multiplier * baseFontSize
		}
		return 0
	}

	// Handle px units
	if strings.HasSuffix(value, "px") {
		num := strings.TrimSuffix(value, "px")
		if size, err := strconv.ParseFloat(num, 64); err == nil {
			return size
		}
	}

	// Plain number (treat as px)
	if size, err := strconv.ParseFloat(value, 64); err == nil {
		return size
	}

	return 0
}

// ParseColor converts color names or hex to color.Color
func ParseColor(value string) color.Color {
	value = strings.ToLower(value)

	// Named colors (CSS standard)
	colors := map[string]color.Color{
		// Basic colors
		"black":        color.Black,
		"white":        color.White,
		"red":          color.RGBA{255, 0, 0, 255},
		"green":        color.RGBA{0, 128, 0, 255},
		"blue":         color.RGBA{0, 0, 255, 255},
		"yellow":       color.RGBA{255, 255, 0, 255},
		"purple":       color.RGBA{128, 0, 128, 255},
		"mediumpurple": color.RGBA{47, 112, 216, 255},
		"orange":       color.RGBA{255, 165, 0, 255},
		"gray":         color.RGBA{128, 128, 128, 255},
		"grey":         color.RGBA{128, 128, 128, 255},
		"cyan":         color.RGBA{0, 255, 255, 255},
		"magenta":      color.RGBA{255, 0, 255, 255},
		"pink":         color.RGBA{255, 192, 203, 255},
		"brown":        color.RGBA{165, 42, 42, 255},

		// Light variants
		"lightgray":   color.RGBA{211, 211, 211, 255},
		"lightgrey":   color.RGBA{211, 211, 211, 255},
		"lightblue":   color.RGBA{173, 216, 230, 255},
		"lightgreen":  color.RGBA{144, 238, 144, 255},
		"lightyellow": color.RGBA{255, 255, 224, 255},
		"lightpink":   color.RGBA{255, 182, 193, 255},
		"lightcyan":   color.RGBA{224, 255, 255, 255},

		// Dark variants
		"darkgray":    color.RGBA{169, 169, 169, 255},
		"darkgrey":    color.RGBA{169, 169, 169, 255},
		"darkblue":    color.RGBA{0, 0, 139, 255},
		"darkgreen":   color.RGBA{0, 100, 0, 255},
		"darkred":     color.RGBA{139, 0, 0, 255},
		"darkcyan":    color.RGBA{0, 139, 139, 255},
		"darkmagenta": color.RGBA{139, 0, 139, 255},
		"darkorange":  color.RGBA{255, 140, 0, 255},

		// Other common colors
		"navy":       color.RGBA{0, 0, 128, 255},
		"teal":       color.RGBA{0, 128, 128, 255},
		"maroon":     color.RGBA{128, 0, 0, 255},
		"olive":      color.RGBA{128, 128, 0, 255},
		"silver":     color.RGBA{192, 192, 192, 255},
		"aqua":       color.RGBA{0, 255, 255, 255},
		"lime":       color.RGBA{0, 255, 0, 255},
		"fuchsia":    color.RGBA{255, 0, 255, 255},
		"gold":       color.RGBA{255, 215, 0, 255},
		"coral":      color.RGBA{255, 127, 80, 255},
		"salmon":     color.RGBA{250, 128, 114, 255},
		"tomato":     color.RGBA{255, 99, 71, 255},
		"crimson":    color.RGBA{220, 20, 60, 255},
		"indigo":     color.RGBA{75, 0, 130, 255},
		"violet":     color.RGBA{238, 130, 238, 255},
		"plum":       color.RGBA{221, 160, 221, 255},
		"khaki":      color.RGBA{240, 230, 140, 255},
		"beige":      color.RGBA{245, 245, 220, 255},
		"ivory":      color.RGBA{255, 255, 240, 255},
		"wheat":      color.RGBA{245, 222, 179, 255},
		"tan":        color.RGBA{210, 180, 140, 255},
		"chocolate":  color.RGBA{210, 105, 30, 255},
		"firebrick":  color.RGBA{178, 34, 34, 255},
		"skyblue":    color.RGBA{135, 206, 235, 255},
		"steelblue":  color.RGBA{70, 130, 180, 255},
		"slategray":  color.RGBA{112, 128, 144, 255},
		"slategrey":  color.RGBA{112, 128, 144, 255},
		"dimgray":    color.RGBA{105, 105, 105, 255},
		"dimgrey":    color.RGBA{105, 105, 105, 255},
		"whitesmoke": color.RGBA{245, 245, 245, 255},
		"snow":       color.RGBA{255, 250, 250, 255},
		"honeydew":   color.RGBA{240, 255, 240, 255},
		"mintcream":  color.RGBA{245, 255, 250, 255},
		"azure":      color.RGBA{240, 255, 255, 255},
		"aliceblue":  color.RGBA{240, 248, 255, 255},
		"lavender":   color.RGBA{230, 230, 250, 255},
		"linen":      color.RGBA{250, 240, 230, 255},
		"seashell":   color.RGBA{255, 245, 238, 255},

		// Extended CSS named colors
		"seagreen":        color.RGBA{46, 139, 87, 255},
		"mediumseagreen":  color.RGBA{60, 179, 113, 255},
		"limegreen":       color.RGBA{50, 205, 50, 255},
		"yellowgreen":     color.RGBA{154, 205, 50, 255},
		"olivedrab":       color.RGBA{107, 142, 35, 255},
		"goldenrod":       color.RGBA{218, 165, 32, 255},
		"darkgoldenrod":   color.RGBA{184, 134, 11, 255},
		"hotpink":         color.RGBA{255, 105, 180, 255},
		"deeppink":        color.RGBA{255, 20, 147, 255},
		"turquoise":       color.RGBA{64, 224, 208, 255},
		"mediumturquoise": color.RGBA{72, 209, 204, 255},
		"cadetblue":       color.RGBA{95, 158, 160, 255},
		"dodgerblue":      color.RGBA{30, 144, 255, 255},
		"royalblue":       color.RGBA{65, 105, 225, 255},
		"cornflowerblue":  color.RGBA{100, 149, 237, 255},
		"mediumblue":      color.RGBA{0, 0, 205, 255},
		"peru":            color.RGBA{205, 133, 63, 255},
		"sienna":          color.RGBA{160, 82, 45, 255},
		"saddlebrown":     color.RGBA{139, 69, 19, 255},
		"orchid":          color.RGBA{218, 112, 214, 255},
		"darkviolet":      color.RGBA{148, 0, 211, 255},
		"darkorchid":      color.RGBA{153, 50, 204, 255},
		"mediumorchid":    color.RGBA{186, 85, 211, 255},
		"palegreen":       color.RGBA{152, 251, 152, 255},
		"lightcoral":      color.RGBA{240, 128, 128, 255},
		"rosybrown":       color.RGBA{188, 143, 143, 255},
		"mistyrose":       color.RGBA{255, 228, 225, 255},
		"bisque":          color.RGBA{255, 228, 196, 255},
		"moccasin":        color.RGBA{255, 228, 181, 255},
		"peachpuff":       color.RGBA{255, 218, 185, 255},
		"darkkhaki":       color.RGBA{189, 183, 107, 255},
		"palegoldenrod":   color.RGBA{238, 232, 170, 255},
	}

	if c, ok := colors[value]; ok {
		return c
	}

	// Hex color: #RGB or #RRGGBB
	if strings.HasPrefix(value, "#") {
		hex := value[1:]
		if len(hex) == 3 {
			// #RGB -> #RRGGBB
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if len(hex) == 6 {
			r, _ := strconv.ParseUint(hex[0:2], 16, 8)
			g, _ := strconv.ParseUint(hex[2:4], 16, 8)
			b, _ := strconv.ParseUint(hex[4:6], 16, 8)
			return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
		}
	}

	return nil
}

type Selector struct {
	TagName     string
	ID          string
	Classes     []string
	PseudoClass string    // e.g. "link", "visited", "hover" — empty means none
	Ancestor    *Selector // non-nil for descendant selectors (e.g. "div p" → p.Ancestor = &div)
}

// Specificity represents CSS selector specificity as (A, B, C):
// A = ID selectors, B = class/pseudo-class selectors, C = element/type selectors.
type Specificity [3]int

// LessThan returns true if s has lower specificity than o.
func (s Specificity) LessThan(o Specificity) bool {
	for i := range s {
		if s[i] != o[i] {
			return s[i] < o[i]
		}
	}
	return false // equal
}

// selectorSpecificity computes the specificity of a selector, including its ancestor chain.
func selectorSpecificity(sel Selector) Specificity {
	sp := Specificity{}
	if sel.ID != "" {
		sp[0]++
	}
	sp[1] += len(sel.Classes)
	if sel.PseudoClass != "" {
		sp[1]++
	}
	if sel.TagName != "" {
		sp[2]++
	}
	if sel.Ancestor != nil {
		anc := selectorSpecificity(*sel.Ancestor)
		for i := range sp {
			sp[i] += anc[i]
		}
	}
	return sp
}

// MatchContext provides runtime state needed for pseudo-class matching.
type MatchContext struct {
	IsVisited  func(url string) bool    // returns true if url has been visited
	ResolveURL func(href string) string // resolves relative hrefs to absolute (optional)
}

type Declaration struct {
	Property  string
	Value     string
	Important bool
}

type Rule struct {
	Selectors    []Selector
	Declarations []Declaration
}

type Stylesheet struct {
	Rules []Rule
}

// MatchSelector checks if a selector matches a DOM node
func MatchSelector(sel Selector, tagName string, id string, classes []string) bool {
	// Check tag name
	if sel.TagName != "" && sel.TagName != tagName {
		return false
	}

	// Check ID
	if sel.ID != "" && sel.ID != id {
		return false
	}

	// Check classes (all selector classes must be present)
	for _, selClass := range sel.Classes {
		found := false
		for _, nodeClass := range classes {
			if selClass == nodeClass {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// MatchSelectorNode checks if a selector (including any descendant chain) matches a DOM node.
func MatchSelectorNode(sel Selector, node *dom.Node, ctx MatchContext) bool {
	if node.Type != dom.Element {
		return false
	}
	id := node.Attributes["id"]
	classes := strings.Fields(node.Attributes["class"])
	if !MatchSelector(sel, node.TagName, id, classes) {
		return false
	}
	// Check pseudo-class
	if sel.PseudoClass != "" {
		href := node.Attributes["href"]
		switch sel.PseudoClass {
		case "link":
			if href == "" {
				return false
			}
			resolvedHref := href
			if ctx.ResolveURL != nil {
				resolvedHref = ctx.ResolveURL(href)
			}
			if ctx.IsVisited != nil && ctx.IsVisited(resolvedHref) {
				return false
			}
		case "visited":
			if href == "" {
				return false
			}
			resolvedHref := href
			if ctx.ResolveURL != nil {
				resolvedHref = ctx.ResolveURL(href)
			}
			if ctx.IsVisited == nil || !ctx.IsVisited(resolvedHref) {
				return false
			}
		default:
			// :hover, :focus, :active, etc. — not yet supported
			return false
		}
	}
	if sel.Ancestor == nil {
		return true
	}
	// Walk up the DOM tree looking for a matching ancestor
	for p := node.Parent; p != nil; p = p.Parent {
		if p.Type != dom.Element {
			continue
		}
		if MatchSelectorNode(*sel.Ancestor, p, ctx) {
			return true
		}
	}
	return false
}

// ApplyStylesheet applies matching rules from stylesheet to a base style
func ApplyStylesheet(sheet Stylesheet, tagName string, id string, classes []string) Style {
	style := DefaultStyle()
	importantProps := make(map[string]bool)

	// Check each rule
	for _, rule := range sheet.Rules {
		// Check if any selector matches
		matches := false
		for _, sel := range rule.Selectors {
			if MatchSelector(sel, tagName, id, classes) {
				matches = true
				break
			}
		}

		if !matches {
			continue
		}

		// Apply declarations
		for _, decl := range rule.Declarations {
			if importantProps[decl.Property] && !decl.Important {
				continue
			}

			applyDeclaration(&style, decl.Property, decl.Value)

			if decl.Important {
				importantProps[decl.Property] = true
			}
		}
	}

	return style
}

// applyDeclaration applies a single CSS property to a style
func applyDeclaration(style *Style, property, value string) {
	switch property {
	case "color":
		if c := ParseColor(value); c != nil {
			style.Color = c
		}
	case "background-color":
		if c := ParseColor(value); c != nil {
			style.BackgroundColor = c
		}
	case "background-image":
		if strings.HasPrefix(value, "url(") && strings.HasSuffix(value, ")") {
			url := value[4 : len(value)-1]
			url = strings.Trim(url, `"'`)
			url = strings.TrimSpace(url)
			style.BackgroundImage = url
		} else if value == "none" {
			style.BackgroundImage = ""
		}
	case "background":
		bgColor, bgImage := parseBackgroundShorthand(value)
		if bgColor != nil {
			style.BackgroundColor = bgColor
		}

		if bgImage != "" {
			style.BackgroundImage = bgImage
		}

	case "font-size":
		if size := ParseSize(value); size > 0 {
			style.FontSize = size
		}
	case "line-height":
		style.LineHeight = parseLineHeight(value, style.FontSize)
	case "font-weight":
		style.Bold = (value == "bold")
	case "font-style":
		style.Italic = (value == "italic")
	case "font-family":
		style.FontFamily = ParseFontFamily(value)
	case "font-variant":
		style.FontVariant = value
	case "margin":
		m := ParseSize(value)
		style.MarginTop = m
		style.MarginBottom = m
		style.MarginLeft = m
		style.MarginRight = m
	case "margin-top":
		style.MarginTop = ParseSize(value)
	case "margin-bottom":
		style.MarginBottom = ParseSize(value)
	case "margin-left":
		style.MarginLeft = ParseSize(value)
	case "margin-right":
		style.MarginRight = ParseSize(value)
	case "padding":
		p := ParseSize(value)
		style.PaddingTop = p
		style.PaddingBottom = p
		style.PaddingLeft = p
		style.PaddingRight = p
	case "padding-top":
		style.PaddingTop = ParseSize(value)
	case "padding-bottom":
		style.PaddingBottom = ParseSize(value)
	case "padding-left":
		style.PaddingLeft = ParseSize(value)
	case "padding-right":
		style.PaddingRight = ParseSize(value)
	case "text-align":
		style.TextAlign = value
	case "vertical-align":
		switch value {
		case "top", "middle", "bottom", "baseline":
			style.VerticalAlign = value
		}
	case "display":
		style.Display = value
	case "float":
		style.Float = value
	case "position":
		style.Position = value
	case "top":
		style.Top = ParseSize(value)
		style.TopSet = true
	case "left":
		style.Left = ParseSize(value)
		style.LeftSet = true
	case "right":
		style.Right = ParseSize(value)
		style.RightSet = true
	case "bottom":
		style.Bottom = ParseSize(value)
		style.BottomSet = true

	case "text-decoration":
		style.TextDecoration = value
	case "text-transform":
		style.TextTransform = value
	case "letter-spacing":
		if ls, ok := parseSpacingWithContext(value, style.FontSize, DefaultViewportWidth, DefaultViewportHeight); ok {
			style.LetterSpacing = ls
			style.LetterSpacingSet = true
		}
	case "word-spacing":
		if ws, ok := parseSpacingWithContext(value, style.FontSize, DefaultViewportWidth, DefaultViewportHeight); ok {
			style.WordSpacing = ws
			style.WordSpacingSet = true
		}
	case "opacity":
		if op, err := strconv.ParseFloat(value, 64); err == nil {
			if op < 0 {
				op = 0
			} else if op > 1 {
				op = 1
			}
			style.Opacity = op
		}
	case "visibility":
		style.Visibility = value
	case "cursor":
		style.Cursor = value
	case "border":
		w, s, c := parseBorderShorthand(value)
		style.BorderTopWidth = w
		style.BorderRightWidth = w
		style.BorderBottomWidth = w
		style.BorderLeftWidth = w
		style.BorderTopStyle = s
		style.BorderRightStyle = s
		style.BorderBottomStyle = s
		style.BorderLeftStyle = s
		style.BorderTopColor = c
		style.BorderRightColor = c
		style.BorderBottomColor = c
		style.BorderLeftColor = c
	case "border-width":
		w := ParseSize(value)
		style.BorderTopWidth = w
		style.BorderRightWidth = w
		style.BorderBottomWidth = w
		style.BorderLeftWidth = w
	case "border-color":
		if c := ParseColor(value); c != nil {
			style.BorderTopColor = c
			style.BorderRightColor = c
			style.BorderBottomColor = c
			style.BorderLeftColor = c
		}
	case "border-style":
		style.BorderTopStyle = value
		style.BorderRightStyle = value
		style.BorderBottomStyle = value
		style.BorderLeftStyle = value
	case "border-top":
		w, s, c := parseBorderShorthand(value)
		style.BorderTopWidth = w
		style.BorderTopStyle = s
		style.BorderTopColor = c
	case "border-right":
		w, s, c := parseBorderShorthand(value)
		style.BorderRightWidth = w
		style.BorderRightStyle = s
		style.BorderRightColor = c
	case "border-bottom":
		w, s, c := parseBorderShorthand(value)
		style.BorderBottomWidth = w
		style.BorderBottomStyle = s
		style.BorderBottomColor = c
	case "border-left":
		w, s, c := parseBorderShorthand(value)
		style.BorderLeftWidth = w
		style.BorderLeftStyle = s
		style.BorderLeftColor = c
	case "border-top-left-radius":
		style.BorderTopLeftRadius = ParseSize(value)
	case "border-top-right-radius":
		style.BorderTopRightRadius = ParseSize(value)
	case "border-bottom-left-radius":
		style.BorderBottomLeftRadius = ParseSize(value)
	case "border-bottom-right-radius":
		style.BorderBottomRightRadius = ParseSize(value)

	case "list-style":
		if listType, ok := parseListStyleShorthand(value); ok {
			style.ListStyleType = listType
		}
	case "list-style-type":
		style.ListStyleType = value

	case "width":
		if strings.HasSuffix(strings.TrimSpace(value), "%") {
			num := strings.TrimSuffix(strings.TrimSpace(value), "%")
			if pct, err := strconv.ParseFloat(num, 64); err == nil && pct > 0 {
				style.WidthPercent = pct
			}
		} else if w := ParseSize(value); w > 0 {
			style.Width = w
		}
	case "height":
		if h := ParseSize(value); h > 0 {
			style.Height = h
		}

	case "min-width":
		if w := ParseSize(value); w > 0 {
			style.MinWidth = w
		}

	case "max-width":
		if w := ParseSize(value); w > 0 {
			style.MaxWidth = w
		}
	case "min-height":
		if h := ParseSize(value); h > 0 {
			style.MinHeight = h
		}
	case "max-height":
		if h := ParseSize(value); h > 0 {
			style.MaxHeight = h
		}
	case "border-radius":
		style.BorderRadius = ParseSize(value)
	}
}

// ParseFontFamily parses a CSS font-family value into a slice of font names
// Example: "Helvetica Neue", Arial, sans-serif → ["Helvetica Neue", "Arial", "sans-serif"]
func ParseFontFamily(value string) []string {
	var fonts []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(value); i++ {
		c := value[i]

		if !inQuote && (c == '"' || c == '\'') {
			// Start of quoted font name
			inQuote = true
			quoteChar = c
		} else if inQuote && c == quoteChar {
			// End of quoted font name
			inQuote = false
			quoteChar = 0
		} else if !inQuote && c == ',' {
			// Separator - save current font
			font := strings.TrimSpace(current.String())
			if font != "" {
				fonts = append(fonts, font)
			}
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}

	// Don't forget the last font
	font := strings.TrimSpace(current.String())
	if font != "" {
		fonts = append(fonts, font)
	}

	return fonts
}

// parseMarginValue returns the size and whether it's "auto"
func parseMarginValue(value string, fontSize, vw, vh float64) (float64, bool) {
	if strings.ToLower(value) == "auto" {
		return 0, true
	}
	return ParseSizeWithContext(value, fontSize, vw, vh), false
}

// parseSpacingWithContext parses spacing values for letter/word spacing.
// Supports: normal, px, em, vh/vw, and unitless numeric values.
func parseSpacingWithContext(value string, fontSize, viewportWidth, viewportHeight float64) (float64, bool) {
	v := strings.TrimSpace(strings.ToLower(value))
	if v == "" {
		return 0, false
	}
	if v == "normal" {
		return 0, true
	}
	num := v
	switch {
	case strings.HasSuffix(v, "px"):
		num = strings.TrimSuffix(v, "px")
	case strings.HasSuffix(v, "em"):
		num = strings.TrimSuffix(v, "em")
	case strings.HasSuffix(v, "vh"):
		num = strings.TrimSuffix(v, "vh")
	case strings.HasSuffix(v, "vw"):
		num = strings.TrimSuffix(v, "vw")
	}
	if _, err := strconv.ParseFloat(num, 64); err != nil {
		return 0, false
	}
	return ParseSizeWithContext(v, fontSize, viewportWidth, viewportHeight), true
}

func applyDeclarationWithContext(style *Style, property, value string, baseFontSize, viewportWidth, viewportHeight float64) {
	switch property {
	case "font-size":
		// font-size em is relative to PARENT's font-size
		if size := ParseSizeWithContext(value, baseFontSize, viewportWidth, viewportHeight); size > 0 {
			style.FontSize = size
		}
	case "margin":
		parts := strings.Fields(value)
		var top, right, bottom, left float64
		var rightAuto, leftAuto bool

		switch len(parts) {
		case 1:
			m, isAuto := parseMarginValue(parts[0], style.FontSize, viewportWidth, viewportHeight)
			top, right, bottom, left = m, m, m, m
			rightAuto, leftAuto = isAuto, isAuto
		case 2:
			top, _ = parseMarginValue(parts[0], style.FontSize, viewportWidth, viewportHeight)
			bottom = top
			right, rightAuto = parseMarginValue(parts[1], style.FontSize, viewportWidth, viewportHeight)
			left, leftAuto = right, rightAuto
		case 3:
			top, _ = parseMarginValue(parts[0], style.FontSize, viewportWidth, viewportHeight)
			right, rightAuto = parseMarginValue(parts[1], style.FontSize, viewportWidth, viewportHeight)
			bottom, _ = parseMarginValue(parts[2], style.FontSize, viewportWidth, viewportHeight)
			left, leftAuto = right, rightAuto
		case 4:
			top, _ = parseMarginValue(parts[0], style.FontSize, viewportWidth, viewportHeight)
			right, rightAuto = parseMarginValue(parts[1], style.FontSize, viewportWidth, viewportHeight)
			bottom, _ = parseMarginValue(parts[2], style.FontSize, viewportWidth, viewportHeight)
			left, leftAuto = parseMarginValue(parts[3], style.FontSize, viewportWidth, viewportHeight)
		}

		style.MarginTop = top
		style.MarginRight = right
		style.MarginBottom = bottom
		style.MarginLeft = left
		style.MarginRightAuto = rightAuto
		style.MarginLeftAuto = leftAuto
	case "margin-top":
		style.MarginTop = ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight)
	case "margin-bottom":
		style.MarginBottom = ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight)
	case "margin-left":
		if strings.ToLower(value) == "auto" {
			style.MarginLeftAuto = true
		} else {
			style.MarginLeft = ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight)
			style.MarginLeftAuto = false
		}
	case "margin-right":
		if strings.ToLower(value) == "auto" {
			style.MarginRightAuto = true
		} else {
			style.MarginRight = ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight)
			style.MarginRightAuto = false
		}
	case "padding":
		p := ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight)
		style.PaddingTop = p
		style.PaddingBottom = p
		style.PaddingLeft = p
		style.PaddingRight = p
	case "padding-top":
		style.PaddingTop = ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight)
	case "padding-bottom":
		style.PaddingBottom = ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight)
	case "padding-left":
		style.PaddingLeft = ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight)
	case "padding-right":
		style.PaddingRight = ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight)
	case "width":
		if strings.HasSuffix(strings.TrimSpace(value), "%") {
			num := strings.TrimSuffix(strings.TrimSpace(value), "%")
			if pct, err := strconv.ParseFloat(num, 64); err == nil && pct > 0 {
				style.WidthPercent = pct
			}
		} else if w := ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight); w > 0 {
			style.Width = w
		}
	case "height":
		if h := ParseSizeWithContext(value, style.FontSize, viewportWidth, viewportHeight); h > 0 {
			style.Height = h
		}
	case "line-height":
		style.LineHeight = parseLineHeight(value, style.FontSize)
	case "letter-spacing":
		if ls, ok := parseSpacingWithContext(value, style.FontSize, viewportWidth, viewportHeight); ok {
			style.LetterSpacing = ls
			style.LetterSpacingSet = true
		}
	case "word-spacing":
		if ws, ok := parseSpacingWithContext(value, style.FontSize, viewportWidth, viewportHeight); ok {
			style.WordSpacing = ws
			style.WordSpacingSet = true
		}
	default:
		// Fall back to original for non-size properties
		applyDeclaration(style, property, value)
	}
}

// ApplyStylesheetWithContext applies matching rules with parent font-size for em units
func ApplyStylesheetWithContext(sheet Stylesheet, node *dom.Node, parentFontSize, viewportWidth, viewportHeight float64, ctx MatchContext) Style {
	tagName := node.TagName
	style := DefaultStyle()
	importantProps := make(map[string]bool)
	specificities := make(map[string]Specificity) // winning specificity per property

	// Apply user-agent default styles based on tag
	applyUserAgentDefaults(&style, tagName, parentFontSize, node, ctx)

	// ruleSpecificity returns the highest specificity among the rule's matching selectors.
	ruleSpecificity := func(rule Rule) (Specificity, bool) {
		best := Specificity{}
		found := false
		for _, sel := range rule.Selectors {
			if MatchSelectorNode(sel, node, ctx) {
				sp := selectorSpecificity(sel)
				if !found || best.LessThan(sp) {
					best = sp
					found = true
				}
			}
		}
		return best, found
	}

	// First pass: find font-size (uses parent's font-size for em)
	for _, rule := range sheet.Rules {
		sp, matches := ruleSpecificity(rule)
		if !matches {
			continue
		}

		for _, decl := range rule.Declarations {
			if decl.Property == "font-size" {
				if importantProps["font-size"] && !decl.Important {
					continue
				}
				if !decl.Important && specificities["font-size"] != (Specificity{}) && sp.LessThan(specificities["font-size"]) {
					continue
				}

				if size := ParseSizeWithContext(decl.Value, parentFontSize, viewportWidth, viewportHeight); size > 0 {
					style.FontSize = size
				}

				if decl.Important {
					importantProps["font-size"] = true
				} else {
					specificities["font-size"] = sp
				}
			}
		}
	}

	// If no font-size was set, inherit from parent
	if style.FontSize == 0 {
		style.FontSize = parentFontSize
	}

	// Second pass: apply other properties (using computed font-size for em)
	for _, rule := range sheet.Rules {
		sp, matches := ruleSpecificity(rule)
		if !matches {
			continue
		}

		for _, decl := range rule.Declarations {
			if decl.Property != "font-size" {
				if importantProps[decl.Property] && !decl.Important {
					continue
				}
				if !decl.Important && specificities[decl.Property] != (Specificity{}) && sp.LessThan(specificities[decl.Property]) {
					continue
				}

				applyDeclarationWithContext(&style, decl.Property, decl.Value, style.FontSize, viewportWidth, viewportHeight)

				if decl.Important {
					importantProps[decl.Property] = true
				} else {
					specificities[decl.Property] = sp
				}
			}
		}
	}

	return style
}

// ParseInlineStyleWithContext parses inline style with font-size context for em units
func ParseInlineStyleWithContext(styleAttr string, parentFontSize, viewportWidth, viewportHeight float64) Style {
	style := DefaultStyle()
	importantProps := make(map[string]bool) // Track !important properties

	parts := strings.Split(styleAttr, ";")

	// First pass: find font-size
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}

		property := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// Check for !important flag
		value, important := stripImportant(value)

		if property == "font-size" {
			// Skip if already set with !important and new value is not
			if importantProps["font-size"] && !important {
				continue
			}

			if size := ParseSizeWithContext(value, parentFontSize, viewportWidth, viewportHeight); size > 0 {
				style.FontSize = size
			}

			if important {
				importantProps["font-size"] = true
			}
		}
	}

	// If no font-size was set, use parent's
	if style.FontSize == 0 {
		style.FontSize = parentFontSize
	}

	// Second pass: apply other properties
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}

		property := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// Check for !important flag
		value, important := stripImportant(value)

		if property != "font-size" {
			// Skip if already set with !important and new value is not
			if importantProps[property] && !important {
				continue
			}

			applyDeclarationWithContext(&style, property, value, style.FontSize, viewportWidth, viewportHeight)

			if important {
				importantProps[property] = true
			}
		}
	}

	return style
}

// applyUserAgentDefaults applies browser default styles for HTML elements
func applyUserAgentDefaults(style *Style, tagName string, fontSize float64, node *dom.Node, ctx MatchContext) {
	switch tagName {
	case "p", "dl":
		style.MarginTop = fontSize
		style.MarginBottom = fontSize
	case "h1":
		style.MarginTop = fontSize * 0.67
		style.MarginBottom = fontSize * 0.67
	case "h2":
		style.MarginTop = fontSize * 0.83
		style.MarginBottom = fontSize * 0.83
	case "h3":
		style.MarginTop = fontSize
		style.MarginBottom = fontSize
	case "h4":
		style.MarginTop = fontSize * 1.33
		style.MarginBottom = fontSize * 1.33
	case "h5":
		style.MarginTop = fontSize * 1.67
		style.MarginBottom = fontSize * 1.67
	case "h6":
		style.MarginTop = fontSize * 2.33
		style.MarginBottom = fontSize * 2.33
	case "ul", "ol":
		style.MarginTop = fontSize
		style.MarginBottom = fontSize
		style.PaddingLeft = 40 // Standard list indentation
	case "blockquote":
		style.MarginTop = fontSize
		style.MarginBottom = fontSize
		style.MarginLeft = 40
		style.MarginRight = 40
	case "hr":
		style.MarginTop = fontSize * 0.5
		style.MarginBottom = fontSize * 0.5
	case "a":
		// UA default link styling — overridable by user CSS rules via specificity cascade
		if node != nil {
			href := node.Attributes["href"]
			if href != "" {
				style.TextDecoration = "underline"
				resolvedHref := href
				if ctx.ResolveURL != nil {
					resolvedHref = ctx.ResolveURL(href)
				}
				if ctx.IsVisited != nil && ctx.IsVisited(resolvedHref) {
					style.Color = color.RGBA{R: 0x55, G: 0x1a, B: 0x8b, A: 0xff} // visited purple
				} else {
					style.Color = color.RGBA{R: 0x00, G: 0x00, B: 0xee, A: 0xff} // unvisited blue
				}
			}
		}
	}
}

// parseLineHeight handles: unitless (1.5), px (24px), normal
func parseLineHeight(value string, fontSize float64) float64 {
	value = strings.TrimSpace(strings.ToLower(value))

	// "normal" keyword = ~1.2 multiplier
	if value == "normal" {
		return fontSize * 1.2
	}

	// Pixel value (e.g., "24px")
	if strings.HasSuffix(value, "px") {
		num := strings.TrimSuffix(value, "px")
		if size, err := strconv.ParseFloat(num, 64); err == nil {
			return size
		}
		return 0
	}

	// Unitless multiplier (e.g., "1.5")
	if multiplier, err := strconv.ParseFloat(value, 64); err == nil {
		return multiplier * fontSize
	}

	return 0
}

func parseBackgroundShorthand(value string) (color.Color, string) {
	var bgColor color.Color
	var bgImage string

	parts := splitBackgroundValue(value)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, "url(") && strings.HasSuffix(part, ")") {
			url := part[4 : len(part)-1]
			url = strings.Trim(url, `"'`)
			bgImage = strings.TrimSpace(url)
			continue
		}

		if part == "none" {
			bgImage = ""
			continue
		}

		if part == "none" {
			bgImage = ""
			continue
		}

		if c := ParseColor(part); c != nil {
			bgColor = c
		}
	}

	return bgColor, bgImage
}

func parseListStyleShorthand(value string) (string, bool) {
	for _, token := range strings.Fields(value) {
		token = strings.ToLower(token)
		switch token {
		case ListStyleNone,
			ListStyleDisc,
			ListStyleCircle,
			ListStyleSquare,
			ListStyleDecimal,
			ListStyleLowerAlpha,
			ListStyleLowerLatin,
			ListStyleUpperAlpha,
			ListStyleUpperLatin,
			ListStyleLowerRoman,
			ListStyleUpperRoman:
			return token, true
		}
	}

	return "", false
}

func splitBackgroundValue(value string) []string {
	var parts []string
	var current strings.Builder
	parenDepth := 0

	for _, ch := range value {
		if ch == '(' {
			parenDepth++
		}
		if ch == ')' {
			parenDepth--
		}

		if ch == ' ' && parenDepth == 0 {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
