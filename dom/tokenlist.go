package dom

import "strings"

type DOMTokenList struct {
	node     *Node
	attrName string
}

func NewDOMTokenList(node *Node, attrName string) *DOMTokenList {
	return &DOMTokenList{
		node:     node,
		attrName: attrName,
	}
}

func (d *DOMTokenList) tokens() []string {
	value := d.node.Attributes[d.attrName]
	if value == "" {
		return []string{}
	}
	return strings.Fields(value)
}

func (d *DOMTokenList) save(tokens []string) {
	d.node.Attributes[d.attrName] = strings.Join(tokens, " ")
}

func (d *DOMTokenList) Length() int {
	return len(d.tokens())
}

func (d *DOMTokenList) Item(index int) string {
	tokens := d.tokens()
	if index < 0 || index >= len(tokens) {
		return ""
	}
	return tokens[index]
}

func (d *DOMTokenList) Contains(token string) bool {
	for _, t := range d.tokens() {
		if t == token {
			return true
		}
	}
	return false
}

func (d *DOMTokenList) Add(token string) {
	if d.Contains(token) {
		return
	}
	tokens := d.tokens()
	tokens = append(tokens, token)
	d.save(tokens)
}

func (d *DOMTokenList) Remove(token string) {
	tokens := d.tokens()
	result := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if t != token {
			result = append(result, t)
		}
	}
	d.save(result)
}

func (d *DOMTokenList) Toggle(token string) bool {
	if d.Contains(token) {
		d.Remove(token)
		return false
	}
	d.Add(token)
	return true
}
