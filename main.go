package main

import (
	"fmt"
	// "log"
	"math"
	"os"

	"github.com/go-gl/gl"
	"github.com/go-gl/glfw"
	"github.com/go-gl/glh"
	"github.com/go-gl/glu"
)

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

	gl.MatrixMode(gl.PROJECTION)
	glu.Perspective(1, float64(w)/float64(h), 0.1, 1000)

	gl.Translated(0, 0, -200)
	gl.Rotated(10, 1, 0, 0)

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
		ambient  []float32 = []float32{0.1, 0.1, 0.5, 1} // ambient light colour.
		diffuse  []float32 = []float32{1, 1, 0.5, 1}     // diffuse light colour.
		lightpos []float32 = []float32{5, 0, 0, 1}       // Position of light source.
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

	var d float64
	d -= 90

	running := true
	for running {
		glfw.SwapBuffers()

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.MatrixMode(gl.PROJECTION)
		// gl.Rotated(1, 0, 1, 0)

		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()

		// gl.Translated(0, 0, -5)

		glu.Sphere(q, 1, 40, 40)

		gl.Rotated(d, 0, 1, 0)
		d += 2

		gl.PushMatrix()
		gl.Translated(1.5, 0, 0)
		glu.Sphere(q, 0.05, 20, 20)
		gl.PopMatrix()

		gl.PushMatrix()
		gl.Rotated(90, 1, 0, 0)
		gl.Scaled(1.5, 1.5, 1)
		gl.Disable(gl.LIGHTING)
		b.Render(gl.LINE_LOOP)
		gl.Enable(gl.LIGHTING)
		gl.PopMatrix()

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
	pos := make([]float64, N*2)
	clr := make([]float64, N*4)
	for i := 0; i < N; i += 1 {
		pos[2*i] = math.Cos(math.Pi * 2 * float64(i) / N)
		pos[2*i+1] = math.Sin(math.Pi * 2 * float64(i) / N)
		clr[4*i] = 1
		clr[4*i+1] = 0
		clr[4*i+2] = 0
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
