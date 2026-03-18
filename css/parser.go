package css

import (
	"strings"
	"unicode"
)

type Parser struct {
	input string
	pos   int
}

func Parse(input string) Stylesheet {
	p := &Parser{input: input, pos: 0}
	return p.parseStylesheet()
}

func (p *Parser) parseStylesheet() Stylesheet {
	var imports []string
	var rules []Rule
	seenRule := false
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		// Handle at-rules (@import, @charset, etc.)
		if p.input[p.pos] == '@' {
			p.pos++ // skip '@'
			importURL := p.parseAtImport()
			if importURL != "" && !seenRule {
				imports = append(imports, importURL)
			}
			continue
		}
		seenRule = true
		rule := p.parseRule()
		rules = append(rules, rule)
	}
	return Stylesheet{Imports: imports, Rules: rules}
}

func (p *Parser) parseRule() Rule {
	selectors := p.parseSelectors()
	declarations := p.parseDeclarations()
	return Rule{Selectors: selectors, Declarations: declarations}
}

func (p *Parser) parseSelectors() []Selector {
	var selectors []Selector
	for {
		sel, valid := p.parseCompoundSelector()
		if valid {
			selectors = append(selectors, sel)
		}
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] == '{' {
			break
		}
		if p.input[p.pos] == ',' {
			p.pos++ // skip comma
		} else {
			// Skip unknown characters (like :link, :hover, etc.)
			p.pos++
		}
	}
	return selectors
}

// parseCompoundSelector parses a selector that may include descendant combinators (spaces).
// e.g. "span.pagetop b" → Selector{TagName:"b", Ancestor:&Selector{TagName:"span", Classes:["pagetop"]}}
func (p *Parser) parseCompoundSelector() (Selector, bool) {
	var parts []Selector
	for {
		part := p.parseSimpleSelector()
		if part.TagName == "" && part.ID == "" && len(part.Classes) == 0 && part.PseudoClass == "" {
			break
		}
		parts = append(parts, part)

		// Peek ahead: whitespace followed by another simple selector = descendant combinator
		savedPos := p.pos
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] == '{' || p.input[p.pos] == ',' {
			p.pos = savedPos // restore; outer loop handles trailing whitespace
			break
		}
		c := p.input[p.pos]
		if c == '>' {
			// child combinator
			p.pos++ // skip '>'
			p.skipWhitespace()
			part := p.parseSimpleSelector()
			if part.TagName == "" && part.ID == "" && len(part.Classes) == 0 && part.PseudoClass == "" {
				break
			}
			part.DirectParent = true
			parts = append(parts, part)
			savedPos = p.pos
			p.skipWhitespace()
			if p.pos >= len(p.input) || p.input[p.pos] == '{' || p.input[p.pos] == ',' {
				p.pos = savedPos
				break
			}
			continue
		}
		if c == '#' || c == '.' || isIdentChar(rune(c)) {
			// descendant combinator: continue parsing next simple selector
			continue
		}
		// unknown char: not a descendant combinator
		p.pos = savedPos
		break
	}

	if len(parts) == 0 {
		return Selector{}, false
	}
	if len(parts) == 1 {
		return parts[0], true
	}

	// Build ancestor chain right-to-left: parts[last] is subject, ancestors link leftward
	subject := parts[len(parts)-1]
	ptr := &subject
	for i := len(parts) - 2; i >= 0; i-- {
		anc := parts[i]
		ptr.Ancestor = &anc
		ptr = ptr.Ancestor
	}
	return subject, true
}

// parseSimpleSelector parses a single simple selector (tag, #id, .class, :pseudo-class combinations).
// Stops at whitespace or any non-selector character.
func (p *Parser) parseSimpleSelector() Selector {
	p.skipWhitespace()
	sel := Selector{}

	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == '#' {
			p.pos++
			sel.ID = p.parseIdentifier()
		} else if c == '.' {
			p.pos++
			sel.Classes = append(sel.Classes, p.parseIdentifier())
		} else if isIdentChar(rune(c)) {
			sel.TagName = p.parseIdentifier()
		} else {
			break
		}
	}

	// Parse optional pseudo-class (e.g. :link, :visited, :hover) or pseudo-element (::before)
	if p.pos < len(p.input) && p.input[p.pos] == ':' {
		p.pos++ // consume first ':'
		if p.pos < len(p.input) && p.input[p.pos] == ':' {
			p.pos++ // consume second ':' for pseudo-elements like ::before
		}
		sel.PseudoClass = p.parseIdentifier()
	}

	return sel
}

func (p *Parser) parseDeclarations() []Declaration {
	var decls []Declaration
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '{' {
		p.pos++ // skip {
	}

	for p.pos < len(p.input) && p.input[p.pos] != '}' {
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] == '}' {
			break
		}
		decl := p.parseDeclaration()
		if decl.Property != "" {
			if strings.EqualFold(decl.Property, "font") {
				if expanded, ok := expandFontShorthand(decl.Value, decl.Important); ok {
					decls = append(decls, expanded...)
				}
				continue
			}
			decls = append(decls, decl)
		}
	}

	if p.pos < len(p.input) && p.input[p.pos] == '}' {
		p.pos++ // skip }
	}
	return decls
}

func (p *Parser) parseDeclaration() Declaration {
	p.skipWhitespace()
	property := p.parseIdentifier()
	p.skipWhitespace()

	if p.pos < len(p.input) && p.input[p.pos] == ':' {
		p.pos++ // skip :
	}

	p.skipWhitespace()
	value := p.parseValue()

	if p.pos < len(p.input) && p.input[p.pos] == ';' {
		p.pos++ // skip ;
	}

	// Check for !important flag
	value, important := stripImportant(value)

	return Declaration{Property: property, Value: value, Important: important}
}

func (p *Parser) parseIdentifier() string {
	start := p.pos
	for p.pos < len(p.input) && isIdentChar(rune(p.input[p.pos])) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *Parser) parseValue() string {
	start := p.pos
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == ';' || c == '}' {
			break
		}
		p.pos++
	}
	return strings.TrimSpace(p.input[start:p.pos])
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) {
		if unicode.IsSpace(rune(p.input[p.pos])) {
			p.pos++
			continue
		}

		if p.pos+1 < len(p.input) && p.input[p.pos] == '/' && p.input[p.pos+1] == '*' {
			p.pos += 2 // skip past /*
			for p.pos+1 < len(p.input) {
				if p.input[p.pos] == '*' && p.input[p.pos+1] == '/' {
					p.pos += 2 // skip past */
					break
				}
				p.pos++
			}
			continue
		}
		break
	}
}

// parseAtImport handles an at-rule after the '@' has been consumed.
// Returns the import URL if it's an @import, otherwise skips the at-rule and returns "".
func (p *Parser) parseAtImport() string {
	keyword := strings.ToLower(p.parseIdentifier())
	if keyword != "import" {
		p.skipAtRule()
		return ""
	}

	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return ""
	}

	var importURL string
	c := p.input[p.pos]
	if c == '"' || c == '\'' {
		importURL = p.parseQuotedString(c)
	} else if c == 'u' && p.pos+3 < len(p.input) && p.input[p.pos:p.pos+4] == "url(" {
		importURL = p.parseURLFunction()
	}

	p.skipToSemicolon()
	return importURL
}

// parseQuotedString reads a string between matching quotes. The opening quote char
// must be at p.pos. Returns the content between quotes.
func (p *Parser) parseQuotedString(quote byte) string {
	p.pos++ // skip opening quote
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != quote {
		p.pos++
	}
	result := p.input[start:p.pos]
	if p.pos < len(p.input) {
		p.pos++ // skip closing quote
	}
	return result
}

// parseURLFunction parses url("..."), url('...'), or url(bare).
// Expects p.pos at the 'u' of 'url('.
func (p *Parser) parseURLFunction() string {
	p.pos += 4 // skip "url("
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return ""
	}

	var result string
	c := p.input[p.pos]
	if c == '"' || c == '\'' {
		result = p.parseQuotedString(c)
	} else {
		// Unquoted URL
		start := p.pos
		for p.pos < len(p.input) && p.input[p.pos] != ')' && !unicode.IsSpace(rune(p.input[p.pos])) {
			p.pos++
		}
		result = p.input[start:p.pos]
	}

	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == ')' {
		p.pos++ // skip ')'
	}
	return result
}

// skipAtRule skips an unknown at-rule. For statement at-rules (@charset, etc.)
// it advances to the next ';'. For block at-rules (@media, etc.) it skips
// the entire { ... } block.
func (p *Parser) skipAtRule() {
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == ';' {
			p.pos++
			return
		}
		if c == '{' {
			// Block at-rule: skip to matching '}'
			depth := 1
			p.pos++
			for p.pos < len(p.input) && depth > 0 {
				if p.input[p.pos] == '{' {
					depth++
				} else if p.input[p.pos] == '}' {
					depth--
				}
				p.pos++
			}
			return
		}
		p.pos++
	}
}

// skipToSemicolon advances past the next ';' or to end of input.
func (p *Parser) skipToSemicolon() {
	for p.pos < len(p.input) && p.input[p.pos] != ';' {
		p.pos++
	}
	if p.pos < len(p.input) {
		p.pos++ // skip ';'
	}
}

func isIdentChar(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsDigit(c) || c == '-' || c == '_'
}
