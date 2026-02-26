package render

import (
	"image"
	"image/color"
	"math"
)

// drawRoundedRectImage renders a rectangle with independent corner radii into
// an RGBA image. Used when the four corners differ (asymmetric border-radius).
func drawRoundedRectImage(w, h int, col color.Color, tl, tr, br, bl float64) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	r, g, b, a := col.RGBA()
	rc := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if insideRoundedRect(x, y, w, h, tl, tr, br, bl) {
				img.SetRGBA(x, y, rc)
			}
		}
	}
	return img
}

// insideRoundedRect returns true if pixel (x, y) is inside a rectangle with
// independent corner radii tl, tr, br, bl (top-left, top-right, bottom-right, bottom-left).
func insideRoundedRect(x, y, w, h int, tl, tr, br, bl float64) bool {
	fx := float64(x) + 0.5
	fy := float64(y) + 0.5
	fw := float64(w)
	fh := float64(h)

	// top-left corner
	if fx < tl && fy < tl {
		return circDist(fx, fy, tl, tl) <= tl
	}
	// top-right corner
	if fx > fw-tr && fy < tr {
		return circDist(fx, fy, fw-tr, tr) <= tr
	}
	// bottom-right corner
	if fx > fw-br && fy > fh-br {
		return circDist(fx, fy, fw-br, fh-br) <= br
	}
	// bottom-left corner
	if fx < bl && fy > fh-bl {
		return circDist(fx, fy, bl, fh-bl) <= bl
	}
	return true
}

func circDist(x1, y1, x2, y2 float64) float64 {
	dx, dy := x1-x2, y1-y2
	return math.Sqrt(dx*dx + dy*dy)
}
