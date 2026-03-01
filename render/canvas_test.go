package render

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/canvas"
	"github.com/stretchr/testify/assert"
)

func TestIsLocalFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// file:// protocol
		{"file protocol basic", "file:///home/user/image.png", true},
		{"file protocol with spaces", "file:///home/user/my image.png", true},

		// Absolute paths
		{"absolute path", "/home/user/image.png", true},
		{"root path", "/image.png", true},

		// Should NOT be local
		{"http url", "http://example.com/image.png", false},
		{"https url", "https://example.com/image.png", false},
		{"protocol relative", "//example.com/image.png", false},
		{"relative path", "images/bg.png", false},
		{"relative with dot", "./images/bg.png", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLocalFile(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToLocalPath(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"file protocol", "file:///home/user/image.png", "/home/user/image.png"},
		{"file protocol root", "file:///image.png", "/image.png"},
		{"already absolute path", "/home/user/image.png", "/home/user/image.png"},
		{"relative path unchanged", "images/bg.png", "images/bg.png"},
		{"http url unchanged", "http://example.com/img.png", "http://example.com/img.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toLocalPath(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveImageURL(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		baseURL  string
		expected string
	}{
		// HTTP URLs - unchanged
		{"http absolute", "http://example.com/img.png", "http://localhost", "http://example.com/img.png"},
		{"https absolute", "https://example.com/img.png", "http://localhost", "https://example.com/img.png"},

		// Protocol-relative
		{"protocol relative", "//cdn.example.com/img.png", "http://localhost", "https://cdn.example.com/img.png"},

		// Local files - unchanged (should NOT prepend baseURL)
		{"file protocol", "file:///home/user/img.png", "http://localhost", "file:///home/user/img.png"},
		{"absolute path", "/home/user/img.png", "http://localhost", "/home/user/img.png"},

		// Relative paths - should prepend baseURL
		{"relative path", "images/bg.png", "http://localhost:8080", "http://localhost:8080/images/bg.png"},

		// Root-relative paths starting with / are treated as local files (not prepended)
		{"root relative treated as local", "/images/bg.png", "http://localhost:8080", "/images/bg.png"},

		// No base URL
		{"no base url", "images/bg.png", "", "images/bg.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveImageURL(tt.src, tt.baseURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSVG(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		contentType string
		expected    bool
	}{
		{"svg extension lowercase", "https://example.com/icon.svg", "", true},
		{"svg extension uppercase", "https://example.com/icon.SVG", "", true},
		{"svg extension with query", "https://example.com/icon.svg?v=1", "", false},
		{"svg content type exact", "https://example.com/icon.png", "image/svg+xml", true},
		{"svg content type with charset", "https://example.com/icon.png", "image/svg+xml; charset=utf-8", true},
		{"non svg", "https://example.com/icon.png", "image/png", false},
		{"empty inputs", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSVG(tt.url, tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderToCanvasDrawTextLetterSpacing(t *testing.T) {
	cmds := []DisplayCommand{
		DrawText{
			Text:          "ABC",
			X:             10,
			Y:             10,
			Width:         100,
			Color:         color.Black,
			Size:          14,
			LetterSpacing: 3,
		},
	}

	objects := RenderToCanvas(cmds, "", "", false, nil)
	textCount := 0
	for _, obj := range objects {
		if _, ok := obj.(*canvas.Text); ok {
			textCount++
		}
	}
	assert.Equal(t, 3, textCount, "expected one canvas.Text object per rune when letter-spacing is set")
}

func TestRenderToCanvasDrawTextWordSpacing(t *testing.T) {
	cmds := []DisplayCommand{
		DrawText{
			Text:        "A B",
			X:           10,
			Y:           10,
			Width:       100,
			Color:       color.Black,
			Size:        14,
			WordSpacing: 4,
		},
	}

	objects := RenderToCanvas(cmds, "", "", false, nil)
	textCount := 0
	for _, obj := range objects {
		if _, ok := obj.(*canvas.Text); ok {
			textCount++
		}
	}
	assert.Equal(t, 3, textCount, "expected one canvas.Text object per rune when word-spacing is set")
}
