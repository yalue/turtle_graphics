// A library for generating "turtle graphics."
// See https://en.wikipedia.org/wiki/Turtle_graphics.
//
// Generally, this library should be used by first creating a "Turtle", calling
// various functions on the Turtle instance to build a list of instructions for
// the turtle to follow, and finally calling Turtle.Render(...) to draw an
// image to a Canvas instance.
package turtle_graphics

import (
	"fmt"
	"image/color"
	"math"
)

// The base interface for setting the style of the line to draw. Different
// canvas implementations may allow richer interfaces here, but they all must
// at least support GetColor.
type StrokeStyle interface {
	// Returns the color of the stroke.
	GetColor() color.Color
}

// A basic implementation of the StrokeStyle interface.
type basicStrokeStyle struct {
	c color.Color
}

func (s *basicStrokeStyle) GetColor() color.Color {
	return s.c
}

// Returns a basic StrokeStyle instance simply specifying the given color.
func GetColorStyle(c color.Color) StrokeStyle {
	return &basicStrokeStyle{
		c: c,
	}
}

// To be as generic as possible, a "Canvas" in this case must be able to handle
// arbitrary floating-point coordinates. Ideally, drawing should work by first
// writing to an instance of the provided DummyCanvas to obtain extents, and
// then using an RGBACanvas to draw a scaled image.
type Canvas interface {
	// Sets the style of the next stroke to draw.
	SetStyle(s StrokeStyle) error
	// To simplify drawing turtle graphics, DrawLine requires the base x, y
	// coordinate, the angle of direction (in degrees), and the length of the
	// line in the given direction.
	DrawLine(x, y, angle, length float64) error
	// DrawArc requires the x, y position of the "turtle". The radius is the
	// distance to the left of the turtle where the center of the circle will
	// be located. (A negative radius puts the circle to the turtle's right.)
	// The startAngle gives the turtle's initial angle, and the degrees is the
	// distance around the circle that the turtle will travel.
	DrawArc(x, y, angle, radius, degrees float64) error
}

// Implements the Canvas interface, but does not actually record lines.
// Instead, this can be used to figure out the extents of the image before it
// is actually drawn, allowing a resulting bitmap rasterization to be properly
// scaled. After drawing an image, call GetExtents to get the boundaries of the
// rectangle. The bounds may not be exact, depending on styles and arcs used.
type DummyCanvas struct {
	minX, maxX, minY, maxY float64
	// Needed so the initial values of 0 don't cause us to miss a proper
	// minimum or maximum, since bounds can be negative.
	initialized bool
}

// Returns a new dummy canvas.
func NewDummyCanvas() *DummyCanvas {
	return &DummyCanvas{
		minX:        0,
		maxX:        0,
		minY:        0,
		maxY:        0,
		initialized: false,
	}
}

// Returns the boundaries of the image that has been drawn to the canvas. May
// not be tight, but will at least contain the image.
func (c *DummyCanvas) GetExtents() (minX, minY, maxX, maxY float64) {
	xTolerance := (c.maxX - c.minX) * 0.001
	yTolerance := (c.maxY - c.minY) * 0.001
	minX = c.minX - xTolerance
	minY = c.minY - yTolerance
	maxX = c.maxX + xTolerance
	maxY = c.maxY + yTolerance
	return
}

func (c *DummyCanvas) SetStyle(s StrokeStyle) error {
	// This is a no-op for the DummyCanvas.
	return nil
}

// Updates the min and max extents of the canvas to ensure they contain x and
// y.
func (c *DummyCanvas) updateBounds(x, y float64) {
	// We store an extra flag in case the max bounds are still negative.
	if !c.initialized {
		c.minX = x
		c.maxX = x
		c.minY = y
		c.maxY = y
		c.initialized = true
		return
	}
	if x < c.minX {
		c.minX = x
	}
	if x > c.maxX {
		c.maxX = x
	}
	if y < c.minY {
		c.minY = y
	}
	if y > c.maxY {
		c.maxY = y
	}
}

// Takes an existing (x, y) position, a direction to move in (in degrees), and
// a distance to move. Returns the new x, y location after moving.
func moveDegrees(x, y, angle, distance float64) (float64, float64) {
	radians := angle * math.Pi / 180.0
	x += math.Cos(radians) * distance
	y += math.Sin(radians) * distance
	return x, y
}

func (c *DummyCanvas) DrawLine(x, y, angle, distance float64) error {
	// Update the bounds based on the start point.
	c.updateBounds(x, y)
	// Update the bounds based on the end point.
	x, y = moveDegrees(x, y, angle, distance)
	c.updateBounds(x, y)
	return nil
}

func (c *DummyCanvas) DrawArc(x, y, angle, radius, degrees float64) error {
	centerX, centerY := moveDegrees(x, y, angle+90.0, radius)
	// Rather than trying to do this specifically, we'll just treat this as if
	// the entire circle must be in bounds. A tighter solution should look at
	// endpoints and the edges instead.
	c.updateBounds(centerX, centerY+radius)
	c.updateBounds(centerX, centerY-radius)
	c.updateBounds(centerX+radius, centerY)
	c.updateBounds(centerX-radius, centerY)
	return nil
}

// An "instruction" that manipulates the turtle's state. This uses an interface
// to allow storing a list of all instructions that can be replayed.
type turtleInstruction interface {
	// Carries out the instruction, using the turtle and the given canvas.
	apply(t *Turtle, c Canvas) error
	// Returns a string representation of the instruction.
	String() string
}

// An instruction telling the turtle to move forward a certain amount.
type moveForwardInstruction struct {
	distance float64
}

func (n *moveForwardInstruction) String() string {
	return fmt.Sprintf("Move forward by %f units", n.distance)
}

func (n *moveForwardInstruction) apply(t *Turtle, c Canvas) error {
	x, y, angle := t.getPosition()
	e := c.DrawLine(x, y, angle, n.distance)
	if e != nil {
		return fmt.Errorf("Failed applying move-forward instruction: %w", e)
	}
	// Update the turtle's position (moving forward won't change its angle)
	x, y = moveDegrees(x, y, angle, n.distance)
	t.position.x = x
	t.position.y = y
	return nil
}

// An instruction telling the turtle to add a certain number of degrees to its
// current angle.
type turnInstruction struct {
	degrees float64
}

func (n *turnInstruction) String() string {
	return fmt.Sprintf("Turn by %f degrees", n.degrees)
}

func (n *turnInstruction) apply(t *Turtle, c Canvas) error {
	angle := n.degrees + t.position.angle
	angle = math.Mod(angle, 360.0)
	t.position.angle = angle
	return nil
}

// Instructs the turtle to change the style of the line it's drawing. Doesn't
// change the turtle's state; only affects the canvas.
type setStyleInstruction struct {
	style StrokeStyle
}

func (n *setStyleInstruction) String() string {
	return "Set style"
}

func (n *setStyleInstruction) apply(t *Turtle, c Canvas) error {
	return c.SetStyle(n.style)
}

// An instruction telling the turtle to draw an arc. Changes the turtle's
// position and direction.
type moveArcInstruction struct {
	// The distance to the turtle's left where the center of the circle is
	// located.
	radius float64
	// The degrees around the perimeter of the circle to move, starting from
	// the turtle's current position.
	degrees float64
}

func (n *moveArcInstruction) String() string {
	return fmt.Sprintf("Move %f degrees along arc radius %f", n.degrees,
		n.radius)
}

func (n *moveArcInstruction) apply(t *Turtle, c Canvas) error {
	x, y, angle := t.getPosition()
	e := c.DrawArc(x, y, angle, n.radius, n.degrees)
	if e != nil {
		return e
	}
	// I had to draw a picture to get this stuff right:
	// angle - 90 = the turtle's original position around the circle
	// degrees + (angle - 90) = the turtle's new position around the circle
	//
	// So, the turtle's new global position can be calculated by moving it a
	// radius from the circle's center, in the direction of its new angle along
	// the circle. Its new global angle is simply its old angle plus the
	// degrees it traveled along the circle.
	centerX, centerY := moveDegrees(x, y, angle+90.0, n.radius)
	newX, newY := moveDegrees(centerX, centerY, n.degrees+(angle-90.0),
		n.radius)
	newAngle := math.Mod(angle+n.degrees, 360.0)
	t.position.x = newX
	t.position.y = newY
	t.position.angle = newAngle
	return nil
}

// Pushes the turtle's current position onto the position stack.
type pushPositionInstruction struct{}

func (n *pushPositionInstruction) String() string {
	return "Push position"
}

func (n *pushPositionInstruction) apply(t *Turtle, c Canvas) error {
	t.positionStack = append(t.positionStack, t.position)
	return nil
}

// Sets the turtle's position to the top position on the stack, removing the
// position from the stack in the process.
type popPositionInstruction struct{}

func (n *popPositionInstruction) String() string {
	return "Pop position"
}

func (n *popPositionInstruction) apply(t *Turtle, c Canvas) error {
	if len(t.positionStack) == 0 {
		return fmt.Errorf("Can't pop the turtle's position: empty stack")
	}
	topIndex := len(t.positionStack) - 1
	t.position = t.positionStack[topIndex]
	t.positionStack = t.positionStack[0:topIndex]
	return nil
}

// Holds the turtle's x and y coordinate, as well as the angle it's facing.
type turtlePosition struct {
	x, y, angle float64
}

func (p *turtlePosition) String() string {
	return fmt.Sprintf("Turtle position: (%f, %f), facing %f degrees", p.x,
		p.y, p.angle)
}

// The "turtle" that moves around.
type Turtle struct {
	// The turtle's current position.
	position turtlePosition
	// A stack of positions, that may be manipulated by instructions. Starts
	// empty. Popping an empty stack is an error.
	positionStack []turtlePosition
	// The instructions the turtle must follow.
	instructions []turtleInstruction
}

// Returns the x, y position of the turtle, followed by the angle it is facing.
func (t *Turtle) getPosition() (float64, float64, float64) {
	return t.position.x, t.position.y, t.position.angle
}

// Adds an instruction to move forward by the given distance to the turtle's
// list of instructions.
func (t *Turtle) MoveForward(distance float64) {
	n := &moveForwardInstruction{
		distance: distance,
	}
	t.instructions = append(t.instructions, n)
}

// Adds an instruction to turn by the given amount to the turtle's list of
// instructions.
func (t *Turtle) Turn(degrees float64) {
	n := &turnInstruction{
		degrees: degrees,
	}
	t.instructions = append(t.instructions, n)
}

// Adds an instruction to change the stroke style to the turtle's list of
// instructions.
func (t *Turtle) SetStyle(style StrokeStyle) {
	n := &setStyleInstruction{
		style: style,
	}
	t.instructions = append(t.instructions, n)
}

// Adds an instruction for the turtle to move the given number of degrees along
// an arc with the specified radius. The center of the circle defining the
// arc will be 90 degrees to the turtle's left.
func (t *Turtle) MoveArc(radius, degrees float64) {
	n := &moveArcInstruction{
		radius:  radius,
		degrees: degrees,
	}
	t.instructions = append(t.instructions, n)
}

// Adds an instruction to push the turtle's current position and orientation
// onto the top of a stack of past positions and orientations.
func (t *Turtle) PushPosition() {
	n := &pushPositionInstruction{}
	t.instructions = append(t.instructions, n)
}

// Adds an instruction to set the turtle's position to whatever is on top of
// the stack of past positions, removing the position from the stack in the
// process.
func (t *Turtle) PopPosition() {
	n := &popPositionInstruction{}
	t.instructions = append(t.instructions, n)
}

// Carries out all of the turtle's stored instructions, writing the results to
// the given canvas.
func (t *Turtle) RenderToCanvas(c Canvas) error {
	var e error
	// Reset any remaining state from past renderings.
	t.positionStack = t.positionStack[0:0]
	t.position = turtlePosition{
		x:     0,
		y:     0,
		angle: 0,
	}
	for i, n := range t.instructions {
		e = n.apply(t, c)
		if e != nil {
			return fmt.Errorf("Error executing instruction %d/%d (%s): %w",
				i+1, len(t.instructions), n.String(), e)
		}
	}
	return nil
}

// Returns an initialized Turtle instance, with no instructions.
func NewTurtle() *Turtle {
	return &Turtle{
		position: turtlePosition{
			x:     0,
			y:     0,
			angle: 0,
		},
		positionStack: make([]turtlePosition, 0, 128),
		instructions:  make([]turtleInstruction, 0, 4096),
	}
}
