// This defines a simple program for testing the turtle graphics package.
// Attempts to create two images, one with an arc and one without.
package main

import (
	"fmt"
	"github.com/yalue/turtle_graphics"
	"image"
	"image/color"
	"image/png"
	"os"
)

// Saves the given image as a PNG file with the given name.
func saveImage(pic image.Image, name string) error {
	f, e := os.Create(name)
	if e != nil {
		return fmt.Errorf("Couldn't create %s: %s", name, e)
	}
	defer f.Close()
	e = png.Encode(f, pic)
	if e != nil {
		return fmt.Errorf("Failed creating PNG image: %s", e)
	}
	return nil
}

// Takes a turtle with a set of instructions and saves the image it creates to
// a PNG file with the given name.
func renderTurtle(t *turtle_graphics.Turtle, name string) error {
	// Get a dummy canvas to compute the image bounds with.
	dummyCanvas := turtle_graphics.NewDummyCanvas()
	e := t.RenderToCanvas(dummyCanvas)
	if e != nil {
		return fmt.Errorf("Failed rendering to dummy canvas: %s", e)
	}
	minX, minY, maxX, maxY := dummyCanvas.GetExtents()
	aspectRatio := (maxX - minX) / (maxY - minY)
	height := 1000
	width := int(float64(height) * aspectRatio)

	// Get a "real" canvas to draw the image on.
	rgbaCanvas, e := turtle_graphics.NewRGBACanvas(width, height, minX, minY,
		maxX, maxY, color.White)
	if e != nil {
		return fmt.Errorf("Failed initializing RGBA canvas: %s", e)
	}
	e = t.RenderToCanvas(rgbaCanvas)
	if e != nil {
		return fmt.Errorf("Failed rendering to RGBA canvas: %s", e)
	}
	e = saveImage(rgbaCanvas, name)
	if e != nil {
		return fmt.Errorf("Failed saving image %s: %s", name, e)
	}
	fmt.Printf("Created %s OK.\n", name)
	return nil
}

func run() int {
	// We'll start by making a basic "Y" shape.
	t := turtle_graphics.NewTurtle()
	t.Turn(90)
	t.MoveForward(1)
	t.Turn(-30)
	t.PushPosition()
	t.MoveForward(1)
	t.PopPosition()
	t.Turn(60)
	t.MoveForward(1)
	e := renderTurtle(t, "basic_y.png")
	if e != nil {
		fmt.Printf("Failed drawing 'Y' image: %s\n", e)
		return 1
	}

	// Now add a couple arcs for testing.
	t.PushPosition()
	t.MoveArc(-0.25, 180)
	t.MoveArc(0.25, 360)
	t.PopPosition()
	t.MoveArc(0.25, 90)
	e = renderTurtle(t, "with_arcs.png")
	if e != nil {
		fmt.Printf("Failed drawing image with arcs: %s\n", e)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
