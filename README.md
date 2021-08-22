Turtle Graphics
===============

A library for generating "Turtle graphics", based on instruction strings. For
the Go language.  Made for fun, there are better alternatives if you care about
features or efficiency.


Usage
-----

See `simple_tester/simple_tester.go` for a more complete example. But, as a
simple showcase:

```
import (
    "github.com/yalue/turtle_graphics"
)

func main() {
    // Draw a simple 'Y' shape. The turtle starts out facing 0 degrees: left.
    t := turtle_graphics.NewTurtle()
    t.Turn(90)
    t.MoveForward(1)
    t.Turn(-30)
    t.PushPosition()
    t.MoveForward(1)
    t.PopPosition()
    t.Turn(60)
    t.MoveForward(1)

    // Get a dummy canvas to compute the image bounds with.
    dummyCanvas := turtle_graphics.NewDummyCanvas()
    err := t.RenderToCanvas(dummyCanvas)
    if err != nil {
        return fmt.Errorf("Failed rendering to dummy canvas: %s", err)
    }
    minX, minY, maxX, maxY := dummyCanvas.GetExtents()
    aspectRatio := (maxX - minX) / (maxY - minY)
    height := 1000
    width := int(float64(height) * aspectRatio)

    // Get a "real" canvas to draw the image on.
    rgbaCanvas, err := turtle_graphics.NewRGBACanvas(width, height, minX, minY,
        maxX, maxY, color.White)
    if err != nil {
        // Handle error
        return fmt.Errorf("Failed initializing RGBA canvas: %s", err)
    }
    err = t.RenderToCanvas(rgbaCanvas)
    if err != nil {
        return fmt.Errorf("Failed rendering to RGBA canvas: %s", err)
    }
    // The rgbaCanvas now satisfies the image/Image interface, and can be saved
    // using any of the standard image libraries.
}

```
