package main

import (
	"fmt"
	"math"
	"os"
	"unsafe"

	"github.com/go-gl/gl"
	"github.com/go-gl/glfw"
	"github.com/go-gl/glh"
	"github.com/go-gl/glu"
)

type Planetoid struct {
	apogee, perigee, inclination, phase0, phase, rising_node float64

	radius float64

	quadric unsafe.Pointer
	circle  *glh.MeshBuffer
}

// func NewPlanetoid(apogee, perigee, inclination, phase0, phase, radius float64,
// 	circle *glh.MeshBuffer) *Planetoid {

// 	// return &Planetoid{apogee, perigee, inclination, phase0, phase, radius, glu.NewQuadric(), circle}
// }

func (p *Planetoid) Render(dp float64) {
	gl.PushMatrix()

	gl.Rotated(p.rising_node, 0, 1, 0)
	gl.Rotated(p.inclination, 0, 0, 1)

	gl.Rotated(p.phase0+p.phase, 0, 1, 0)
	p.phase += dp

	gl.PushMatrix()
	gl.Translated(p.apogee, 0, 0)
	glu.Sphere(p.quadric, float32(p.radius), 20, 20)
	gl.PopMatrix()

	gl.PushMatrix()
	gl.Rotated(90, 1, 0, 0)
	gl.Scaled(p.apogee, p.apogee, 1)
	gl.Disable(gl.LIGHTING)
	p.circle.Render(gl.LINE_STRIP)
	gl.Enable(gl.LIGHTING)
	gl.PopMatrix()

	gl.PopMatrix()
}

func main() {
	var err error
	if err = glfw.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "[e] %v\n", err)
		return
	}

	defer glfw.Terminate()

	w, h := 1980, 1080
	if err = glfw.OpenWindow(w, h, 8, 8, 8, 16, 0, 32, glfw.Fullscreen); err != nil {
		fmt.Fprintf(os.Stderr, "[e] %v\n", err)
		return
	}

	defer glfw.CloseWindow()

	glfw.SetSwapInterval(1)
	glfw.SetWindowTitle("Simple GLFW window")

	glfw.SetWindowSizeCallback(onResize)
	glfw.SetWindowCloseCallback(onClose)
	glfw.SetMouseButtonCallback(onMouseBtn)
	glfw.SetMouseWheelCallback(onMouseWheel)
	glfw.SetKeyCallback(onKey)
	glfw.SetCharCallback(onChar)

	q := glu.NewQuadric()

	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.ShadeModel(gl.SMOOTH)
	// gl.Enable(gl.LIGHTING_BIT)
	gl.Enable(gl.LIGHTING)
	// gl.LightModelf(gl.LIGHT_MODEL_AMBIENT, 0.5)
	// gl.Materialf(face, pname, param)
	// gl.L
	// gl.Lightf(gl.LIGHT0, gl.LIGHT0, 0)
	// log.Printf("Blah: %v", gl.GetError())
	var (
		ambient  []float32 = []float32{0.1, 0.3, 0.6, 1} // ambient light colour.
		diffuse  []float32 = []float32{1, 1, 0.5, 1}     // diffuse light colour.
		lightpos []float32 = []float32{100000, 0, 0, 1}  // Position of light source.
	)

	gl.Lightfv(gl.LIGHT1, gl.AMBIENT, ambient)
	gl.Lightfv(gl.LIGHT1, gl.DIFFUSE, diffuse)
	// gl.Lightfv(gl.LIGHT1, gl.SPECULAR, diffuse)
	gl.Lightfv(gl.LIGHT1, gl.POSITION, lightpos)
	gl.Enable(gl.LIGHT1)

	gl.MatrixMode(gl.MODELVIEW)

	gl.ClearColor(0, 0, 0, 1)
	gl.ClearDepth(1)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	b := createBuffer()

	planetoids := []*Planetoid{}
	for i := 0; i < 10; i++ {
		r := 0.05 * (math.Cos(float64(i)*79) + 1)
		// p := NewPlanetoid(r, 1.5, 1.5, 0, , b)
		p := &Planetoid{
			apogee:      1.2,
			perigee:     1.5,
			inclination: 0,
			phase0:      float64(i) * 57.1,
			phase:       0,
			radius:      r,
			quadric:     glu.NewQuadric(),
			circle:      b,
		}
		planetoids = append(planetoids, p)
	}

	// Initial projection matrix:

	gl.MatrixMode(gl.PROJECTION)
	glu.Perspective(1, float64(w)/float64(h), 0.1, 300)

	gl.Translated(0, 0, -200)
	gl.Rotated(20, 1, 0, 0)

	running := true
	for running {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Rotate the planet
		gl.MatrixMode(gl.PROJECTION)
		// gl.Rotated(0.05, 0, 1, 0)
		gl.Rotated(0.5, 0, 1, 0)

		// Start afresh each time
		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()

		// Earth
		glu.Sphere(q, 1, 100, 100)

		// Atmosphere
		gl.Disable(gl.LIGHTING)
		gl.Disable(gl.DEPTH_TEST)
		gl.Color4f(0.25, 0.25, 1, 0.25)
		glu.Sphere(q, 1.025, 100, 100)

		gl.Enable(gl.DEPTH_TEST)

		gl.PointSize(10)
		gl.Begin(gl.POINTS)
		gl.Color4f(0.75, 0.75, 0.75, 1)
		gl.Vertex3d(-1.02, 0, 0)
		gl.End()

		gl.Enable(gl.LIGHTING)

		for _, p := range planetoids {
			const dt = 0.1 // TODO: Frame update
			p.Render(dt)
		}

		glfw.SwapBuffers()

		running = (glfw.Key(glfw.KeyEsc) == 0 &&
			glfw.WindowParam(glfw.Opened) == 1)
	}
}

func onResize(w, h int) {
	fmt.Printf("resized: %dx%d\n", w, h)
}

func onClose() int {
	fmt.Println("closed")
	return 1 // return 0 to keep window open.
}

func onMouseBtn(button, state int) {
	fmt.Printf("mouse button: %d, %d\n", button, state)
}

func onMouseWheel(delta int) {
	fmt.Printf("mouse wheel: %d\n", delta)
}

func onKey(key, state int) {
	fmt.Printf("key: %d, %d\n", key, state)
}

func onChar(key, state int) {
	fmt.Printf("char: %d, %d\n", key, state)
}

func createBuffer() *glh.MeshBuffer {
	const N = 128
	const f = 0.05
	pos := make([]float64, N*2)
	clr := make([]float64, N*4)
	for i := 0; i < N; i += 1 {
		pos[2*i] = math.Cos(f * 2 * math.Pi * 2 * float64(i) / N)
		pos[2*i+1] = math.Sin(f * 2 * math.Pi * 2 * float64(i) / N)
		clr[4*i] = 0.75
		clr[4*i+1] = 0.2
		clr[4*i+2] = 0.9
		clr[4*i+3] = 1 - float64(i)/float64(N)
	}

	// Create a mesh buffer with the given attributes.
	mb := glh.NewMeshBuffer(
		glh.RenderArrays,

		glh.NewPositionAttr(2, gl.DOUBLE, gl.STATIC_DRAW),
		glh.NewColorAttr(4, gl.DOUBLE, gl.STATIC_DRAW),
	)

	// Add the mesh to the buffer.
	mb.Add(pos, clr)
	return mb
}
