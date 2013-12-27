package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"unsafe"

	"image"
	// "image/color"
	"image/png"

	"github.com/go-gl/gl"
	"github.com/go-gl/glfw"
	"github.com/go-gl/glh"
	"github.com/go-gl/glu"
)

// TODO:

// * Proper orbits
// * Texturing
// * Shaders

type Planetoid struct {
	apogee, perigee, inclination, phase0, phase, rising_node float64

	radius float32

	// quadric unsafe.Pointer
	circle *glh.MeshBuffer
}

var quadric unsafe.Pointer

// Render a sphere with the correct orientation (poles along y-axis)
// Also makes a correction to avoid lighting flicker.
// TODO: Replace with geodesic
func Sphere(radius float32, granularity int) {
	glh.With(glh.Matrix{gl.MODELVIEW}, func() {
		// Needed to avoid a strange lighting effect
		gl.Rotatef(90, 0, 1, 0)

		// Rotate so that poles are on the y-axis
		gl.Rotatef(90, 1, 0, 0)
		glu.Sphere(quadric, radius, granularity, granularity)
	})
}

func (p *Planetoid) Render(dp float64) {
	glh.With(glh.Matrix{gl.MODELVIEW}, func() {

		gl.Rotated(p.rising_node, 0, 1, 0)
		gl.Rotated(p.inclination, 0, 0, 1)

		gl.Rotated(p.phase0+p.phase, 0, 1, 0)
		p.phase += dp*10 + (1 - p.apogee)

		glh.With(glh.Matrix{gl.MODELVIEW}, func() {
			// TODO: Compute position correctly
			gl.Translated(p.apogee, 0, 0)
			Sphere(p.radius, 3)
		})

		// TODO: Avoid depth thrashing when trails overlap
		// Trail
		glh.With(glh.Matrix{gl.MODELVIEW}, func() {
			gl.Rotated(90, 1, 0, 0)
			gl.Scaled(p.apogee, p.apogee, 1)

			gl.LineWidth(2)
			glh.With(glh.Disable(gl.LIGHTING), func() {
				p.circle.Render(gl.LINE_STRIP)
			})
		})
	})
}

func (p *Planetoid) RenderShadowVolume() {
	glh.With(glh.Matrix{gl.MODELVIEW}, func() {

		// gl.Rotated(90, 0, 1, 0)

		gl.Rotated(p.rising_node, 0, 1, 0)
		gl.Rotated(p.inclination, 0, 0, 1)

		gl.Rotated(p.phase0+p.phase, 0, 1, 0)

		// TODO: Compute position correctly
		gl.Translated(p.apogee, 0, 0)

		gl.Rotated(p.phase0+p.phase, 0, -1, 0)
		gl.Rotated(p.inclination, 0, 0, -1)
		gl.Rotated(p.rising_node, 0, -1, 0)

		// gl.Rotated(p.phase0+p.phase, 0, -1, 0)

		gl.Rotated(90, 0, -1, 0)
		glu.Cylinder(quadric, p.radius, p.radius, 2, 10, 1)
	})
}

func main() {
	var err error
	if err = glfw.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "[e] %v\n", err)
		return
	}

	defer glfw.Terminate()

	w, h := 1980, 1080
	// w, h := 1280, 768
	if err = glfw.OpenWindow(w, h, 8, 8, 8, 16, 0, 32, glfw.Fullscreen); err != nil {
		fmt.Fprintf(os.Stderr, "[e] %v\n", err)
		return
	}

	defer glfw.CloseWindow()

	glfw.SetSwapInterval(1)
	glfw.SetWindowTitle("Debris")

	quadric = glu.NewQuadric()

	gl.Enable(gl.CULL_FACE)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)

	gl.Enable(gl.NORMALIZE)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.ShadeModel(gl.SMOOTH)
	gl.Enable(gl.LIGHTING)

	var (
		ambient        = []float32{0.1, 0.3, 0.6, 1}
		diffuse        = []float32{1, 1, 0.5, 1}
		specular       = []float32{0.4, 0.4, 0.4, 1}
		light_position = []float32{1, 0, 0, 0}

		// mat_specular  []float32 = []float32{1, 1, 0.5, 1}
		mat_specular  = []float32{1, 1, 0.75, 1}
		mat_shininess = float32(120)
		// light_position []float32 = []float32{0.0, 0.0, 1.0, 0.0}
	)

	const (
		fov               = 1.1 // degrees
		znear             = 145
		zfar              = 155
		camera_z_offset   = -150
		camera_x_rotation = 0 // degrees
		// camera_x_rotation = 20 // degrees

		starfield_fov = 45

		faces        = 1000
		earth_radius = 1
	)

	gl.Lightfv(gl.LIGHT1, gl.AMBIENT, ambient)
	gl.Lightfv(gl.LIGHT1, gl.DIFFUSE, diffuse)
	gl.Lightfv(gl.LIGHT1, gl.SPECULAR, specular)
	gl.Lightfv(gl.LIGHT1, gl.POSITION, light_position)
	gl.Enable(gl.LIGHT1)

	mat_emission := []float32{0, 0, 0.1, 1}
	gl.Materialfv(gl.FRONT_AND_BACK, gl.EMISSION, mat_emission)
	gl.Materialfv(gl.FRONT_AND_BACK, gl.SPECULAR, mat_specular)
	gl.Materialf(gl.FRONT_AND_BACK, gl.SHININESS, mat_shininess)

	gl.ClearColor(0.02, 0.02, 0.02, 1)
	gl.ClearDepth(1)
	gl.ClearStencil(0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	b := createBuffer()

	planetoids := []*Planetoid{}
	for i := 0; i < 1000; i++ {
		p := &Planetoid{
			apogee:  1.2 + rand.Float64()*0.7,
			perigee: 1.5,
			// inclination: 45,
			inclination: rand.Float64()*20 - 10,
			// inclination: 0,
			phase0:      rand.Float64() * 360,
			rising_node: rand.Float64() * 10,
			phase:       0,
			// radius:      rand.Float32()*0.05 + 0.01, //float32(r),
			radius: rand.Float32()*0.0125 + 0.005, //float32(r),
			// quadric:     glu.NewQuadric(),
			circle: b,
		}
		planetoids = append(planetoids, p)
	}

	// Initial projection matrix:

	var aspect float64

	glfw.SetWindowSizeCallback(func(w, h int) {
		gl.Viewport(0, 0, w, h)

		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()
		aspect = float64(w) / float64(h)
		glu.Perspective(fov, aspect, znear, zfar)
	})

	d := float64(0)

	wireframe := false
	atmosphere := false
	polar := false
	rotating := false
	front := false
	earth := true
	cone := true
	shadowing := true
	tilt := false
	running := true

	glfw.SetKeyCallback(func(key, state int) {
		if state != glfw.KeyPress {
			// Don't act on key coming up
			return
		}

		switch key {
		case 'A':
			atmosphere = !atmosphere
		case 'C':
			cone = !cone
		case 'E':
			earth = !earth
		case 'R':
			rotating = !rotating
		case 'F':
			front = !front
			if front {
				gl.FrontFace(gl.CW)
			} else {
				gl.FrontFace(gl.CCW)
			}
		case 'S':
			shadowing = !shadowing
		case 'T':
			tilt = !tilt
		case 'W':
			wireframe = !wireframe
			method := gl.GLenum(gl.FILL)
			if wireframe {
				method = gl.LINE
			}
			gl.PolygonMode(gl.FRONT_AND_BACK, method)

		case glfw.KeyF2:
			println("Screenshot captured")
			// glh.CaptureToPng("screenshot.png")

			w, h := glh.GetViewportWH()
			im := image.NewRGBA(image.Rect(0, 0, w, h))
			glh.ClearAlpha(1)
			gl.Flush()
			glh.CaptureRGBA(im)

			go func() {
				fd, err := os.Create("screenshot.png")
				if err != nil {
					panic("Unable to open file")
				}
				defer fd.Close()

				png.Encode(fd, im)
			}()

		case 'Q', glfw.KeyEsc:
			running = !running

		case glfw.KeySpace:
			polar = !polar
		}
	})

	_ = rand.Float64

	stars := glh.NewMeshBuffer(
		glh.RenderArrays,
		glh.NewPositionAttr(3, gl.DOUBLE, gl.STATIC_DRAW),
		glh.NewColorAttr(3, gl.DOUBLE, gl.STATIC_DRAW))

	const Nstars = 50000
	points := make([]float64, 3*Nstars)
	colors := make([]float64, 3*Nstars)

	for i := 0; i < Nstars; i++ {
		const R = 1

		phi := rand.Float64() * 2 * math.Pi
		z := R * (2*rand.Float64() - 1)
		theta := math.Asin(z / R)

		points[i*3+0] = R * math.Cos(theta) * math.Cos(phi)
		points[i*3+1] = R * math.Cos(theta) * math.Sin(phi)
		points[i*3+2] = z

		const r = 0.8
		v := rand.Float64()*r + (1 - r)
		colors[i*3+0] = v
		colors[i*3+1] = v
		colors[i*3+2] = v
	}

	stars.Add(points, colors)

	render_stars := func() {
		glh.With(glh.Attrib{gl.DEPTH_BUFFER_BIT | gl.ENABLE_BIT}, func() {
			gl.Disable(gl.LIGHTING)
			gl.PointSize(1)
			gl.Color4f(1, 1, 1, 1)

			gl.Disable(gl.DEPTH_TEST)
			gl.DepthMask(false)

			stars.Render(gl.POINTS)
		})
	}

	render_scene := func() {

		// Update light position (sensitive to current modelview matrix)
		gl.Lightfv(gl.LIGHT1, gl.POSITION, light_position)
		gl.Lightfv(gl.LIGHT2, gl.POSITION, light_position)

		if earth {
			Sphere(earth_radius, faces)
		}

		unlit_points := glh.Compound(glh.Disable(gl.LIGHTING), glh.Primitive{gl.POINTS})
		glh.With(unlit_points, func() {
			gl.Vertex3d(1, 0, 0)
		})

		for _, p := range planetoids {
			const dt = 0.1 // TODO: Frame update
			p.Render(dt)
		}

		glh.With(glh.Disable(gl.LIGHTING), func() {
			// Atmosphere
			gl.Color4f(0.25, 0.25, 1, 0.1)

			if atmosphere && earth {
				Sphere(earth_radius*1.025, 100)
			}

			gl.PointSize(10)

			glh.With(glh.Primitive{gl.POINTS}, func() {
				gl.Color4f(1.75, 0.75, 0.75, 1)
				gl.Vertex3d(-1.04, 0, 0)
			})
		})

	}

	render_shadow_volume := func() {

		glh.With(glh.Attrib{
			gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.ENABLE_BIT |
				gl.POLYGON_BIT | gl.STENCIL_BUFFER_BIT,
		}, func() {

			gl.Disable(gl.LIGHTING)

			if shadowing {
				// gl.Disable(gl.DEPTH_TEST)
				gl.DepthMask(false)
				gl.DepthFunc(gl.LEQUAL)

				gl.Enable(gl.STENCIL_TEST)
				gl.ColorMask(false, false, false, false)
				gl.StencilFunc(gl.ALWAYS, 1, 0xffffffff)
			}

			shadow_volume := func() {
				const sv_length = 2
				const sv_granularity = 100
				const sv_radius = earth_radius * 1.001

				// Shadow cone
				glh.With(glh.Matrix{gl.MODELVIEW}, func() {
					gl.Rotatef(90, 1, 0, 0)
					gl.Rotatef(90, 0, -1, 0)
					gl.Color4f(0.5, 0.5, 0.5, 1)
					glu.Cylinder(quadric, sv_radius, sv_radius*1.05,
						sv_length, sv_granularity, 1)

					glu.Disk(quadric, 0, sv_radius, sv_granularity, 1)

					glh.With(glh.Matrix{gl.MODELVIEW}, func() {
						gl.Translated(0, 0, sv_length)
						glu.Disk(quadric, 0, sv_radius*1.05, sv_granularity, 1)
					})
				})

				for _, p := range planetoids {
					p.RenderShadowVolume()
				}

			}

			if cone {
				gl.FrontFace(gl.CCW)
				gl.StencilOp(gl.KEEP, gl.KEEP, gl.INCR)

				shadow_volume()

				gl.FrontFace(gl.CW)
				gl.StencilOp(gl.KEEP, gl.KEEP, gl.DECR)

				shadow_volume()
			}

			if shadowing {
				gl.StencilFunc(gl.NOTEQUAL, 0, 0xffffffff)
				gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)

				gl.ColorMask(true, true, true, true)
				// gl.Disable(gl.STENCIL_TEST)

				gl.Disable(gl.DEPTH_TEST)

				gl.FrontFace(gl.CCW)
				// gl.Color4f(1, 0, 0, 0.75)
				gl.Color4f(0, 0, 0, 0.75)
				// gl.Color4f(1, 1, 1, 0.75)

				gl.LoadIdentity()
				gl.Translated(0, 0, camera_z_offset)
				// TODO: Figure out why this doesn't draw over the whole screen
				glh.With(glh.Disable(gl.LIGHTING), func() {
					glh.DrawQuadd(-10, -10, 20, 20)
				})

				// gl.FrontFace(gl.CW)
				// gl.Enable(gl.LIGHTING)
				// gl.Disable(gl.LIGHT1)
				// render_scene()
				// gl.Enable(gl.LIGHT1)
			}
		})
	}
	_ = render_stars

	for running {
		running = glfw.WindowParam(glfw.Opened) == 1

		glfw.SwapBuffers()

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)

		rotation := func() {
			if tilt {
				gl.Rotated(20, 1, 0, 0)
			}

			if polar {
				gl.Rotated(90, 1, 0, 0)
			}

			gl.Rotated(d, 0, -1, 0)
		}

		// Star field

		glh.With(glh.Matrix{gl.PROJECTION}, func() {
			gl.LoadIdentity()
			glu.Perspective(starfield_fov, aspect, 0, 1)

			glh.With(glh.Matrix{gl.MODELVIEW}, func() {
				gl.LoadIdentity()
				rotation()
				render_stars()
			})
		})

		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()
		gl.Translated(0, 0, camera_z_offset)

		rotation()
		if rotating {
			d += 0.2
		}

		_ = render_scene
		render_scene()

		_ = render_shadow_volume
		render_shadow_volume()
	}
}

func createBuffer() *glh.MeshBuffer {
	const N = 128
	const f = 0.025
	pos := make([]float64, N*2)
	clr := make([]float64, N*4)
	for i := 0; i < N; i += 1 {
		pos[2*i+0] = math.Cos(f * 2 * math.Pi * 2 * float64(i) / N)
		pos[2*i+1] = math.Sin(f * 2 * math.Pi * 2 * float64(i) / N)
		clr[4*i+0] = 0.2
		clr[4*i+1] = 0.75
		clr[4*i+2] = 0.9
		clr[4*i+3] = 0.5 * (1 - float64(i)/float64(N-1))
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
