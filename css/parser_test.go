package css

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Stylesheet
	}{
		{
			name:  "empty stylesheet",
			input: "",
			expected: Stylesheet{
				Rules: nil,
			},
		},
		{
			name:  "whitespace only",
			input: "   \n\t  ",
			expected: Stylesheet{
				Rules: nil,
			},
		},
		{
			name:  "single rule with tag selector",
			input: "div { color: red; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div"}},
						Declarations: []Declaration{{Property: "color", Value: "red"}},
					},
				},
			},
		},
		{
			name:  "single rule with id selector",
			input: "#main { font-size: 16px; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{ID: "main"}},
						Declarations: []Declaration{{Property: "font-size", Value: "16px"}},
					},
				},
			},
		},
		{
			name:  "single rule with class selector",
			input: ".container { margin: 10px; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{Classes: []string{"container"}}},
						Declarations: []Declaration{{Property: "margin", Value: "10px"}},
					},
				},
			},
		},
		{
			name:  "combined selector tag and class",
			input: "div.foo { padding: 5px; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div", Classes: []string{"foo"}}},
						Declarations: []Declaration{{Property: "padding", Value: "5px"}},
					},
				},
			},
		},
		{
			name:  "combined selector tag class and id",
			input: "div.foo#bar { color: blue; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div", ID: "bar", Classes: []string{"foo"}}},
						Declarations: []Declaration{{Property: "color", Value: "blue"}},
					},
				},
			},
		},
		{
			name:  "multiple selectors",
			input: "div, p { color: red; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors: []Selector{
							{TagName: "div"},
							{TagName: "p"},
						},
						Declarations: []Declaration{{Property: "color", Value: "red"}},
					},
				},
			},
		},
		{
			name:  "multiple declarations",
			input: "div { color: red; font-size: 16px; margin: 10px; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors: []Selector{{TagName: "div"}},
						Declarations: []Declaration{
							{Property: "color", Value: "red"},
							{Property: "font-size", Value: "16px"},
							{Property: "margin", Value: "10px"},
						},
					},
				},
			},
		},
		{
			name:  "multiple rules",
			input: "div { color: red; } p { color: blue; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div"}},
						Declarations: []Declaration{{Property: "color", Value: "red"}},
					},
					{
						Selectors:    []Selector{{TagName: "p"}},
						Declarations: []Declaration{{Property: "color", Value: "blue"}},
					},
				},
			},
		},
		{
			name: "multiline stylesheet",
			input: `
				body {
					background-color: white;
					font-size: 14px;
				}
				h1 {
					color: black;
				}
			`,
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors: []Selector{{TagName: "body"}},
						Declarations: []Declaration{
							{Property: "background-color", Value: "white"},
							{Property: "font-size", Value: "14px"},
						},
					},
					{
						Selectors:    []Selector{{TagName: "h1"}},
						Declarations: []Declaration{{Property: "color", Value: "black"}},
					},
				},
			},
		},
		{
			name:  "multiple classes in selector",
			input: ".foo.bar { color: red; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{Classes: []string{"foo", "bar"}}},
						Declarations: []Declaration{{Property: "color", Value: "red"}},
					},
				},
			},
		},
		// !important tests
		{
			name:  "declaration with !important",
			input: "div { color: red !important; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div"}},
						Declarations: []Declaration{{Property: "color", Value: "red", Important: true}},
					},
				},
			},
		},
		{
			name:  "declaration with !important no space",
			input: "div { color: red!important; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div"}},
						Declarations: []Declaration{{Property: "color", Value: "red", Important: true}},
					},
				},
			},
		},
		{
			name:  "declaration with !IMPORTANT uppercase",
			input: "div { color: red !IMPORTANT; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors:    []Selector{{TagName: "div"}},
						Declarations: []Declaration{{Property: "color", Value: "red", Important: true}},
					},
				},
			},
		},
		{
			name:  "mixed important and normal declarations",
			input: "div { color: red !important; font-size: 16px; margin: 10px !important; }",
			expected: Stylesheet{
				Rules: []Rule{
					{
						Selectors: []Selector{{TagName: "div"}},
						Declarations: []Declaration{
							{Property: "color", Value: "red", Important: true},
							{Property: "font-size", Value: "16px", Important: false},
							{Property: "margin", Value: "10px", Important: true},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)

			assert.Len(t, result.Rules, len(tt.expected.Rules), "number of rules")

			for i, rule := range result.Rules {
				if i >= len(tt.expected.Rules) {
					break
				}
				expectedRule := tt.expected.Rules[i]
				assert.Equal(t, expectedRule.Selectors, rule.Selectors, "Rule[%d].Selectors", i)
				assert.Equal(t, expectedRule.Declarations, rule.Declarations, "Rule[%d].Declarations", i)
			}
		})
	}
}

func TestParseDescendantSelector(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantSels  []Selector
	}{
		{
			name:  "simple tag - no ancestor",
			input: `b { color: red; }`,
			wantSels: []Selector{
				{TagName: "b"},
			},
		},
		{
			name:  "two-level descendant: span.pagetop b",
			input: `span.pagetop b { font-size: 15px; }`,
			wantSels: []Selector{
				{TagName: "b", Ancestor: &Selector{TagName: "span", Classes: []string{"pagetop"}}},
			},
		},
		{
			name:  "two-level descendant: div p",
			input: `div p { color: blue; }`,
			wantSels: []Selector{
				{TagName: "p", Ancestor: &Selector{TagName: "div"}},
			},
		},
		{
			name:  "three-level descendant: div p a",
			input: `div p a { color: green; }`,
			wantSels: []Selector{
				{TagName: "a", Ancestor: &Selector{TagName: "p", Ancestor: &Selector{TagName: "div"}}},
			},
		},
		{
			name:  "comma-separated still works",
			input: `div, p { color: red; }`,
			wantSels: []Selector{
				{TagName: "div"},
				{TagName: "p"},
			},
		},
		{
			name:  "comma-separated with descendant: div p, span b",
			input: `div p, span b { color: red; }`,
			wantSels: []Selector{
				{TagName: "p", Ancestor: &Selector{TagName: "div"}},
				{TagName: "b", Ancestor: &Selector{TagName: "span"}},
			},
		},
		{
			// Pre-existing behaviour: ':' is skipped by the unknown-char branch,
			// so 'hover' gets parsed as a second (spurious) selector.
			// The important thing is 'a' appears first and is not corrupted.
			name:  "pseudo-class: a tag still parsed first",
			input: `a:hover { color: red; }`,
			wantSels: []Selector{
				{TagName: "a"},
				{TagName: "hover"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sheet := Parse(tt.input)
			assert.Len(t, sheet.Rules, 1, "expected one rule")
			if len(sheet.Rules) > 0 {
				assert.Equal(t, tt.wantSels, sheet.Rules[0].Selectors)
			}
		})
	}
}
