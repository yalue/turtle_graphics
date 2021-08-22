package turtle_graphics

// This file contains a canvas implementation that satisfying the image.Image
// interface, that can actually be used for rendering the turtle graphics.

import (
	"fmt"
	"image"
	"image/color"
)

// Keeps track of an RGBA image, along with the canvas boundaries needed to
// draw the turtle's path. Implements both the image.Image and Canvas
// interfaces.
type RGBACanvas struct {
	// The style with which to draw strokes.
	style StrokeStyle
	// The underlying RGBA image.
	pic *image.RGBA
	// The width and height of the image, in pixels.
	pixelsWide, pixelsTall int
	// The bounds of the image, in canvas units. If the turtle is outside of
	// these bounds, its path will not be rendered.
	minX, maxX, minY, maxY float64
	// The distance between pixels in the X and Y directions.
	dX, dY float64
}

func (c *RGBACanvas) ColorModel() color.Model {
	return c.pic.ColorModel()
}

func (c *RGBACanvas) Bounds() image.Rectangle {
	return c.pic.Bounds()
}

func (c *RGBACanvas) At(x, y int) color.Color {
	return c.pic.At(x, y)
}

// Allocates a new black RGBA canvas. Requires the width and height in pixels
// as well as the boundaries of the actual "canvas" in the turtle's units.
// Paths outside of the boundaries will not be included in the image.
func NewRGBACanvas(pixelsWide, pixelsTall int, minX, minY, maxX,
	maxY float64, background color.Color) (*RGBACanvas, error) {
	// Check that we were given sane bounds.
	if pixelsWide <= 0 {
		return nil, fmt.Errorf("Pixels wide must be positive. Got %d",
			pixelsWide)
	}
	if pixelsTall <= 0 {
		return nil, fmt.Errorf("Pixels tall must be positive. Got %d",
			pixelsTall)
	}
	if maxX <= minX {
		return nil, fmt.Errorf("Min X boundary (%f) must be less than the "+
			"max X boundary (%f)", minX, maxX)
	}
	if maxY <= minY {
		return nil, fmt.Errorf("Min Y boundary (%f) must be less than the "+
			"max Y boundary (%f)", minY, maxY)
	}

	// Allocate the resulting image and fill in the background color.
	pic := image.NewRGBA(image.Rect(0, 0, pixelsWide, pixelsTall))
	for y := 0; y < pixelsTall; y++ {
		for x := 0; x < pixelsWide; x++ {
			pic.Set(x, y, background)
		}
	}

	toReturn := &RGBACanvas{
		style:      GetColorStyle(color.Black),
		pic:        pic,
		pixelsWide: pixelsWide,
		pixelsTall: pixelsTall,
		minX:       minX,
		maxX:       maxX,
		minY:       minY,
		maxY:       maxY,
		dX:         (maxX - minX) / float64(pixelsWide),
		dY:         (maxY - minY) / float64(pixelsTall),
	}
	return toReturn, nil
}

func (c *RGBACanvas) SetStyle(s StrokeStyle) error {
	c.style = s
	return nil
}

func abs(x int) int {
	if x >= 0 {
		return x
	}
	return -x
}

// Draws a line from integer coordinates (x0, y0) to (x1, y1) in the given
// image. Uses Bresenham's line algorithm, based on the C version found at
// https://rosettacode.org/wiki/Bitmap/Bresenham%27s_line_algorithm.
func drawLine(x0, y0, x1, y1 int, pic *image.RGBA, style StrokeStyle) {
	dx := abs(x1 - x0)
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	dy := abs(y1 - y0)
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx / 2
	if dx <= dy {
		err = -dy / 2
	}
	var e2 int
	c := style.GetColor()
	for {
		pic.Set(x0, y0, c)
		if (x0 == x1) && (y0 == y1) {
			break
		}
		e2 = err
		if e2 > -dx {
			err -= dy
			x0 += sx
		}
		if e2 < dy {
			err += dx
			y0 += sy
		}
	}
}

// Takes a position in the canvas units, and returns the pixel it maps to. May
// return negative pixel coordinates, or coordinates otherwise outside of the
// actual rendered image.
func (c *RGBACanvas) PointToPixel(x, y float64) (int, int) {
	// We'll use yMax to flip the y-coordinate of the image, so things aren't
	// drawn upside down.
	yMax := c.pixelsTall - 1
	pixelsX := int((x - c.minX) / c.dX)
	pixelsY := int((y - c.minY) / c.dY)
	return pixelsX, yMax - pixelsY
}

func (c *RGBACanvas) DrawLine(x, y, angle, length float64) error {
	x0, y0 := c.PointToPixel(x, y)
	newX, newY := moveDegrees(x, y, angle, length)
	x1, y1 := c.PointToPixel(newX, newY)
	drawLine(x0, y0, x1, y1, c.pic, c.style)
	return nil
}

func (c *RGBACanvas) DrawArc(x, y, angle, radius, degrees float64) error {
	centerX, centerY := moveDegrees(x, y, angle+90.0, radius)

	// Make degrees positive, and clamp to 360 (when rasterizing, we don't care
	// if a turtle goes around multiple times, or in what direction).
	if degrees < 0 {
		degrees = 360 - degrees
	}
	if degrees > 360 {
		degrees = 360
	}

	// Compute a loose bound on the number of pixels in the circumference based
	// on the number of pixels in the radius. Check the radius both horizontal
	// and vertically, since pixels may not be square.
	absRadius := radius
	if absRadius < 0 {
		absRadius = -absRadius
	}
	radiusPixelsX := int(absRadius / c.dX) + 1
	radiusPixelsY := int(absRadius / c.dY) + 1
	var sampleCount int
	if radiusPixelsX >= radiusPixelsY {
		sampleCount = radiusPixelsX * 7
	} else {
		sampleCount = radiusPixelsY * 7
	}

	// The number of degrees between samples in the arc
	dA := degrees / float64(sampleCount)

	// This is the angle pointing to the turtle from the center of the circle.
	// (Draw a picture if you need to.)
	currentAngle := angle - 90.0

	// Actually move along the arc and draw the points.
	var pixelX, pixelY int
	for i := 0; i < sampleCount; i++ {
		x, y = moveDegrees(centerX, centerY, currentAngle, radius)
		pixelX, pixelY = c.PointToPixel(x, y)
		c.pic.Set(pixelX, pixelY, c.style.GetColor())
		currentAngle += dA
	}

	return nil
}
