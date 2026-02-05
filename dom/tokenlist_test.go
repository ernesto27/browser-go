package dom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDOMTokenList_Length(t *testing.T) {
	tests := []struct {
		name     string
		rel      string
		expected int
	}{
		{"empty string", "", 0},
		{"single token", "noopener", 1},
		{"two tokens", "noopener noreferrer", 2},
		{"three tokens", "noopener noreferrer sponsored", 3},
		{"extra whitespace", "  noopener   noreferrer  ", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewElement("a", map[string]string{"rel": tt.rel})
			list := NewDOMTokenList(node, "rel")
			assert.Equal(t, tt.expected, list.Length())
		})
	}
}

func TestDOMTokenList_Item(t *testing.T) {
	node := NewElement("a", map[string]string{"rel": "noopener noreferrer sponsored"})
	list := NewDOMTokenList(node, "rel")

	tests := []struct {
		name     string
		index    int
		expected string
	}{
		{"first item", 0, "noopener"},
		{"second item", 1, "noreferrer"},
		{"third item", 2, "sponsored"},
		{"out of bounds positive", 3, ""},
		{"out of bounds negative", -1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, list.Item(tt.index))
		})
	}
}

func TestDOMTokenList_Contains(t *testing.T) {
	node := NewElement("a", map[string]string{"rel": "noopener noreferrer"})
	list := NewDOMTokenList(node, "rel")

	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{"exists first", "noopener", true},
		{"exists second", "noreferrer", true},
		{"does not exist", "nofollow", false},
		{"empty string", "", false},
		{"partial match", "noopen", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, list.Contains(tt.token))
		})
	}
}

func TestDOMTokenList_Add(t *testing.T) {
	tests := []struct {
		name        string
		initial     string
		toAdd       string
		expectedRel string
	}{
		{"add to empty", "", "noopener", "noopener"},
		{"add new token", "noopener", "noreferrer", "noopener noreferrer"},
		{"add duplicate", "noopener noreferrer", "noopener", "noopener noreferrer"},
		{"add third token", "noopener noreferrer", "sponsored", "noopener noreferrer sponsored"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewElement("a", map[string]string{"rel": tt.initial})
			list := NewDOMTokenList(node, "rel")
			list.Add(tt.toAdd)
			assert.Equal(t, tt.expectedRel, node.Attributes["rel"])
		})
	}
}

func TestDOMTokenList_Remove(t *testing.T) {
	tests := []struct {
		name        string
		initial     string
		toRemove    string
		expectedRel string
	}{
		{"remove first", "noopener noreferrer", "noopener", "noreferrer"},
		{"remove last", "noopener noreferrer", "noreferrer", "noopener"},
		{"remove middle", "noopener noreferrer sponsored", "noreferrer", "noopener sponsored"},
		{"remove nonexistent", "noopener noreferrer", "nofollow", "noopener noreferrer"},
		{"remove only token", "noopener", "noopener", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewElement("a", map[string]string{"rel": tt.initial})
			list := NewDOMTokenList(node, "rel")
			list.Remove(tt.toRemove)
			assert.Equal(t, tt.expectedRel, node.Attributes["rel"])
		})
	}
}

func TestDOMTokenList_Toggle(t *testing.T) {
	tests := []struct {
		name           string
		initial        string
		toToggle       string
		expectedResult bool
		expectedRel    string
	}{
		{"toggle on", "noopener", "noreferrer", true, "noopener noreferrer"},
		{"toggle off", "noopener noreferrer", "noreferrer", false, "noopener"},
		{"toggle on empty", "", "noopener", true, "noopener"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewElement("a", map[string]string{"rel": tt.initial})
			list := NewDOMTokenList(node, "rel")
			result := list.Toggle(tt.toToggle)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedRel, node.Attributes["rel"])
		})
	}
}

func TestDOMTokenList_SyncsWithAttribute(t *testing.T) {
	node := NewElement("a", map[string]string{"rel": "noopener"})
	list := NewDOMTokenList(node, "rel")

	// Add token and verify attribute updated
	list.Add("noreferrer")
	assert.Equal(t, "noopener noreferrer", node.Attributes["rel"])

	// Remove token and verify attribute updated
	list.Remove("noopener")
	assert.Equal(t, "noreferrer", node.Attributes["rel"])

	// Toggle and verify attribute updated
	list.Toggle("sponsored")
	assert.Equal(t, "noreferrer sponsored", node.Attributes["rel"])

	// Verify list reads from updated attribute
	assert.Equal(t, 2, list.Length())
	assert.True(t, list.Contains("noreferrer"))
	assert.True(t, list.Contains("sponsored"))
}
