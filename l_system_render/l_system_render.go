// This is a simple package for playing around with L-systems and turtle
// graphics.
package main

import (
	"fmt"
	"github.com/yalue/l_system"
	"github.com/yalue/turtle_graphics"
	"os"
)

// A function that is intended to issue some instruction to a
// turtle_graphics.Turtle instance. This doesn't return any errors since the
// Turtle.MoveForward(...), PushPosition(), etc, don't return errors. Used when
// mapping string bytes to turtle movements.
type TurtleInstruction func(t *turtle_graphics.Turtle)

// Maintains all the information needed to associate an L-system-generated
// string with turtle movements.
type LSystemTurtle struct {
	L *l_system.LSystem
	// Contains 256 entries; one corresponding to each possible byte generated
	// by the L-system string. nil entries mean the corresponding byte does
	// nothing.
	CharMapping []TurtleInstruction
}

// Returns a new L-system turtle, initializing the L-system with the given
// string. CharMapping is allocated but filled with nil, so this won't draw
// anything until some instructions have been assigned to string bytes.
func NewLSystemTurtle(initialString []byte) *LSystemTurtle {
	return &LSystemTurtle{
		L:           l_system.NewLSystem(initialString),
		CharMapping: make([]TurtleInstruction, 256),
	}
}

// Returns a turtle that follows the instructions specified by the L-system.
func (s *LSystemTurtle) GetTurtle() (*turtle_graphics.Turtle, error) {
	t := turtle_graphics.NewTurtle()
	chars := s.L.GetValue()
	var f TurtleInstruction
	// The very simple loop where we apply the specified instructions.
	for _, c := range chars {
		f = s.CharMapping[c]
		if f != nil {
			f(t)
		}
	}
	// This can't return an error for now, but it's always nice to have it as
	// an option in the future.
	return t, nil
}

func moveForward(t *turtle_graphics.Turtle) {
	t.MoveForward(1.0)
}

func turnLeft(t *turtle_graphics.Turtle) {
	t.Turn(90.0)
}

func turnRight(t *turtle_graphics.Turtle) {
	t.Turn(-90.0)
}

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
	var e error

	// Define the productions and movement instructions to generate the dragon
	// curve.
	s := NewLSystemTurtle([]byte("F"))
	s.L.SetProduction('F', []byte("F+G"))
	s.L.SetProduction('G', []byte("F-G"))
	s.CharMapping['F'] = moveForward
	s.CharMapping['G'] = moveForward
	s.CharMapping['-'] = turnRight
	s.CharMapping['+'] = turnLeft

	// Iterate the dragon curve 16 times.
	for i := 0; i < 16; i++ {
		e = s.L.Iterate()
		if e != nil {
			fmt.Printf("Error iterating the dragon curve: %s\n", e)
			return 1
		}
	}

	// Get the turtle and save it to a PNG image.
	fmt.Printf("Length of instruction string: %d bytes.\n", s.L.GetSize())
	t, e := s.GetTurtle()
	if e != nil {
		fmt.Printf("Error getting the dragon-curve turtle: %s\n", e)
		return 1
	}
	e = saveImage(t, "dragon_curve.png")
	if e != nil {
		fmt.Printf("Error saving dragon curve to a PNG: %s\n", e)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
