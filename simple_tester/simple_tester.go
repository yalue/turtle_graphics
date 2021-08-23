// This defines a simple program for testing the turtle graphics package.
// Attempts to create two images, one with an arc and one without.
package main

import (
	"fmt"
	"github.com/yalue/turtle_graphics"
	"os"
)

// Saves the given image as a PNG file with the given name.
func saveImage(t *turtle_graphics.Turtle, name string) error {
	f, e := os.Create(name)
	if e != nil {
		return fmt.Errorf("Couldn't create %s: %s", name, e)
	}
	defer f.Close()
	e = turtle_graphics.SaveTurtleAsPNG(t, 1000, f)
	if e != nil {
		return fmt.Errorf("Failed rendering turtle to %s: %s", name, e)
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
	e := saveImage(t, "basic_y.png")
	if e != nil {
		fmt.Printf("Failed drawing 'Y' image: %s\n", e)
		return 1
	}

	// Now add a couple arcs for testing.
	t.PushPosition()
	t.MoveArc(-0.25, 180)
	t.MoveArc(0.25, 360)
	t.PopPosition()
	t.PushPosition()
	t.MoveArc(0.25, 90)
	t.PopPosition()
	t.MoveArc(0.25, -90)
	e = saveImage(t, "with_arcs.png")
	if e != nil {
		fmt.Printf("Failed drawing image with arcs: %s\n", e)
		return 1
	}

	// Now we'll draw a basic 'T' shape with rounded corners.
	t = turtle_graphics.NewTurtle()
	t.MoveForward(0.5)
	t.MoveArc(0.125, 90)
	t.MoveForward(2)
	t.MoveArc(-0.125, -90)
	t.MoveForward(0.75)
	t.MoveArc(0.125, 90)
	t.MoveForward(0.5)
	t.MoveArc(0.125, 90)
	t.MoveForward(2.5)
	t.MoveArc(0.125, 90)
	t.MoveForward(0.5)
	t.MoveArc(0.125, 90)
	t.MoveForward(0.75)
	t.MoveArc(-0.125, -90)
	t.MoveForward(2)
	t.MoveArc(0.125, 90)
	e = saveImage(t, "t_shape.png")
	if e != nil {
		fmt.Printf("Failed drawing t-shape image: %s\n", e)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
