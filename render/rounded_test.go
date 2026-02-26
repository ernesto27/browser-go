package render

import (
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCircDist(t *testing.T) {
	tests := []struct {
		name     string
		x1, y1   float64
		x2, y2   float64
		expected float64
	}{
		{"3-4-5 triangle", 0, 0, 3, 4, 5},
		{"same point", 5, 5, 5, 5, 0},
		{"horizontal only", 0, 0, 10, 0, 10},
		{"vertical only", 0, 0, 0, 7, 7},
		{"negative direction", 3, 4, 0, 0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := circDist(tt.x1, tt.y1, tt.x2, tt.y2)
			assert.InDelta(t, tt.expected, got, 1e-9)
		})
	}
}

func TestInsideRoundedRect(t *testing.T) {
	const w, h = 100, 100

	tests := []struct {
		name     string
		x, y     int
		tl, tr   float64
		br, bl   float64
		expected bool
	}{
		// Center is always inside regardless of radii
		{"center no radii", 50, 50, 0, 0, 0, 0, true},
		{"center with radii", 50, 50, 20, 20, 20, 20, true},

		// Sharp corners (radius=0): all four corners are inside
		{"top-left corner sharp", 0, 0, 0, 0, 0, 0, true},
		{"top-right corner sharp", 99, 0, 0, 0, 0, 0, true},
		{"bottom-right corner sharp", 99, 99, 0, 0, 0, 0, true},
		{"bottom-left corner sharp", 0, 99, 0, 0, 0, 0, true},

		// Top-left radius=20: pixel at (0,0) is outside the arc
		// fx=0.5 < 20, fy=0.5 < 20: dist from center (20,20) ≈ 27.6 > 20 → outside
		{"tl corner outside arc", 0, 0, 20, 0, 0, 0, false},
		// Pixel at (15,15): dist from (20,20) ≈ 6.4 < 20 → inside
		{"tl corner inside arc", 15, 15, 20, 0, 0, 0, true},
		// Pixel just outside tl region (fx >= tl) → not in corner zone → inside
		{"tl region boundary x", 20, 0, 20, 0, 0, 0, true},

		// Top-right radius=20: pixel at (99,0) is outside the arc
		// fx=99.5 > 80, fy=0.5 < 20: dist from (80,20) ≈ 27.6 > 20 → outside
		{"tr corner outside arc", 99, 0, 0, 20, 0, 0, false},
		// Pixel at (84,15): fx=84.5>80, fy=15.5<20, dist from (80,20) ≈ 6.4 → inside
		{"tr corner inside arc", 84, 15, 0, 20, 0, 0, true},

		// Bottom-right radius=20: pixel at (99,99) outside arc
		// fx=99.5>80, fy=99.5>80: dist from (80,80) ≈ 27.6 > 20 → outside
		{"br corner outside arc", 99, 99, 0, 0, 20, 0, false},
		// Pixel at (84,84): dist from (80,80) ≈ 5.66 < 20 → inside
		{"br corner inside arc", 84, 84, 0, 0, 20, 0, true},

		// Bottom-left radius=20: pixel at (0,99) outside arc
		// fx=0.5<20, fy=99.5>80: dist from (20,80) ≈ 27.6 > 20 → outside
		{"bl corner outside arc", 0, 99, 0, 0, 0, 20, false},
		// Pixel at (15,84): dist from (20,80) ≈ 6.4 < 20 → inside
		{"bl corner inside arc", 15, 84, 0, 0, 0, 20, true},

		// Asymmetric: only tl=40, rest=0
		// Pixel (0,0) in tl zone: dist from (40,40) = sqrt(39.5^2+39.5^2) ≈ 55.9 > 40 → outside
		{"asymmetric tl only corner", 0, 0, 40, 0, 0, 0, false},
		// tr zone still sharp (radius=0): pixel (99,0) not in any arc zone → inside
		{"asymmetric tr sharp when tl set", 99, 0, 40, 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := insideRoundedRect(tt.x, tt.y, w, h, tt.tl, tt.tr, tt.br, tt.bl)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDrawRoundedRectImage(t *testing.T) {
	t.Run("image has correct bounds", func(t *testing.T) {
		img := drawRoundedRectImage(80, 60, color.RGBA{255, 0, 0, 255}, 0, 0, 0, 0)
		b := img.Bounds()
		assert.Equal(t, 80, b.Max.X)
		assert.Equal(t, 60, b.Max.Y)
	})

	t.Run("center pixel is filled with given color", func(t *testing.T) {
		col := color.RGBA{100, 150, 200, 255}
		img := drawRoundedRectImage(100, 100, col, 0, 0, 0, 0)
		r, g, b, a := img.At(50, 50).RGBA()
		assert.Equal(t, uint32(100), r>>8)
		assert.Equal(t, uint32(150), g>>8)
		assert.Equal(t, uint32(200), b>>8)
		assert.Equal(t, uint32(255), a>>8)
	})

	t.Run("corner pixel is transparent when radius clips it", func(t *testing.T) {
		// Large radius=40 on top-left: pixel (0,0) should be clipped → transparent
		img := drawRoundedRectImage(100, 100, color.RGBA{255, 0, 0, 255}, 40, 0, 0, 0)
		_, _, _, a := img.At(0, 0).RGBA()
		assert.Equal(t, uint32(0), a>>8, "corner pixel should be transparent (outside arc)")
	})

	t.Run("corner pixel is filled when no radius", func(t *testing.T) {
		img := drawRoundedRectImage(100, 100, color.RGBA{255, 0, 0, 255}, 0, 0, 0, 0)
		_, _, _, a := img.At(0, 0).RGBA()
		assert.Equal(t, uint32(255), a>>8, "corner pixel should be opaque with zero radius")
	})

	t.Run("all four corners clipped with symmetric radius", func(t *testing.T) {
		img := drawRoundedRectImage(100, 100, color.RGBA{0, 255, 0, 255}, 30, 30, 30, 30)
		corners := [][2]int{{0, 0}, {99, 0}, {99, 99}, {0, 99}}
		for _, c := range corners {
			_, _, _, a := img.At(c[0], c[1]).RGBA()
			assert.Equal(t, uint32(0), a>>8, "corner (%d,%d) should be transparent", c[0], c[1])
		}
	})

	t.Run("arc boundary: pixel just inside radius is filled", func(t *testing.T) {
		// tl=20: pixel (15,15) → dist from (20,20) ≈ 7.07 < 20 → inside
		img := drawRoundedRectImage(100, 100, color.RGBA{0, 0, 255, 255}, 20, 0, 0, 0)
		_, _, _, a := img.At(15, 15).RGBA()
		dist := math.Sqrt(math.Pow(15.5-20, 2) + math.Pow(15.5-20, 2))
		if dist <= 20 {
			assert.Equal(t, uint32(255), a>>8, "pixel inside arc should be filled")
		}
	})
}
