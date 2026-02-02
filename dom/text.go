package dom

import "strings"

var blockElements = map[string]bool{
	"html": true, "body": true, "div": true,
	"p": true, "h1": true, "h2": true, "h3": true,
	"h4": true, "h5": true, "h6": true,
	"ul": true, "ol": true, "li": true,
	"table": true, "tr": true,
	"header": true, "footer": true, "main": true,
	"section": true, "article": true, "nav": true,
	"blockquote": true, "pre": true, "hr": true,
}

var skipElements = map[string]bool{
	"script": true, "style": true, "head": true,
	"meta": true, "link": true, "noscript": true,
}

func (n *Node) InnerText() string {
	var sb strings.Builder
	n.innerTextRecursive(&sb, false)
	return strings.TrimSpace(sb.String())
}

func (n *Node) SetInnerText(text string) {
	n.Children = []*Node{}

	if text != "" {
		texNode := NewText(text)
		n.AppendChild(texNode)
	}
}

func (n *Node) innerTextRecursive(sb *strings.Builder, prevBlock bool) bool {
	if n.Type == Element && skipElements[n.TagName] {
		return prevBlock
	}

	isBlock := n.Type == Element && blockElements[n.TagName]

	if isBlock && !prevBlock {
		sb.WriteString("\n")
	}

	if n.Type == Text {
		sb.WriteString(n.Text)
		sb.WriteString(" ")
		return false
	}

	lastWasBlock := isBlock
	for _, child := range n.Children {
		lastWasBlock = child.innerTextRecursive(sb, lastWasBlock)
	}

	if isBlock {
		sb.WriteString("\n")
		return true
	}

	return lastWasBlock
}

// ExpandTabs replaces tab characters with spaces to align to tab stops.
// Default tab width is 8 characters.
func ExpandTabs(text string, tabWidth int) string {
	if !strings.Contains(text, "\t") {
		return text
	}

	if tabWidth <= 0 {
		tabWidth = 8
	}

	var result strings.Builder
	col := 0

	for _, ch := range text {
		switch ch {
		case '\t':
			// Calculate spaces needed to reach next tab stop
			spaces := tabWidth - (col % tabWidth)
			for i := 0; i < spaces; i++ {
				result.WriteRune(' ')
			}
			col += spaces
		case '\n':
			result.WriteRune(ch)
			col = 0
		default:
			result.WriteRune(ch)
			col++
		}
	}

	return result.String()
}
