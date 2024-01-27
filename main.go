package main

import (
	"fmt"
	"math"
	"time"

	"github.com/gowebapi/webapi"
	"github.com/gowebapi/webapi/core/js"
	"github.com/gowebapi/webapi/core/jsconv"
	"github.com/gowebapi/webapi/graphics/webgl"
	"github.com/gowebapi/webapi/html/canvas"
)

//see https://github.com/golang/go/wiki/WebAssembly
//see https://github.com/bobcob7/wasm-basic-triangle

var gl *webgl.RenderingContext
var vBuffer, iBuffer *webgl.Buffer
var icount int
var prog *webgl.Program
var angle float32
var width, height int

func main() {
	c := make(chan struct{}, 0)
	fmt.Println("Go/WASM loaded")

	addCanvas()

	<-c
}

func addCanvas() {
	doc := webapi.GetWindow().Document()
	app := doc.GetElementById("app")
	body := doc.GetElementById("body")
	width := body.ClientWidth()
	height := body.ClientHeight()

	canvasE := webapi.GetWindow().Document().CreateElement("canvas", &webapi.Union{js.ValueOf("dom.Node")})
	canvasE.SetId("canvas")
	app.AppendChild(&canvasE.Node)
	canvasHTML := canvas.HTMLCanvasElementFromWrapper(canvasE)
	canvasHTML.SetWidth(uint(width))
	canvasHTML.SetHeight(uint(height))

	contextU := canvasHTML.GetContext("webgl", nil)
	gl = webgl.RenderingContextFromWrapper(contextU)

	vBuffer, iBuffer, icount = createBuffers(gl)

	prog = setupShaders(gl)

	// Start the animation loop
	js.Global().Call("requestAnimationFrame", js.FuncOf(drawScene))
}

func drawScene(this js.Value, p []js.Value) interface{} {
	// Start a timer
	startTime := time.Now()

	angle += 0.01 // Update the angle for rotation

	gl.ClearColor(0.5, 0.5, 0.5, 0.9)
	gl.Clear(webgl.COLOR_BUFFER_BIT)

	// Enable the depth test
	gl.Enable(webgl.DEPTH_TEST)

	// Set the view port
	gl.Viewport(0, 0, 800, 800)

	// Update the model-view matrix for rotation
	rotationMatrix := getRotationMatrix(angle)
	coord := gl.GetAttribLocation(prog, "coordinates")

	// Bind vertex buffer object
	gl.BindBuffer(webgl.ARRAY_BUFFER, vBuffer)

	// Bind index buffer object
	gl.BindBuffer(webgl.ELEMENT_ARRAY_BUFFER, iBuffer)

	// Point an attribute to the currently bound VBO
	gl.VertexAttribPointer(uint(coord), 3, webgl.FLOAT, false, 0, 0)

	// Enable the attribute
	gl.EnableVertexAttribArray(uint(coord))

	// Set the model-view matrix in the vertex shader
	modelviewLoc := gl.GetUniformLocation(prog, "modelview")
	gl.UniformMatrix4fv(modelviewLoc, false, rotationMatrix)

	// Draw the triangle
	gl.DrawElements(webgl.TRIANGLES, icount, webgl.UNSIGNED_SHORT, 0)

	// Stop the timer and calculate the FPS
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	fps := float64(1.0 / elapsedTime.Seconds())

	// Display the FPS
	fpsDisplay := fmt.Sprintf("FPS: %.2f", fps)
	doc := webapi.GetWindow().Document()
	fpsElem := doc.GetElementById("fps")
	fpsElem.SetTextContent(&fpsDisplay)

	// Request the next animation frame
	js.Global().Call("requestAnimationFrame", js.FuncOf(drawScene))
	return nil
}

func createBuffers(gl *webgl.RenderingContext) (*webgl.Buffer, *webgl.Buffer, int) {
	//// VERTEX BUFFER ////
	var verticesNative = []float32{
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,
	}
	var vertices = jsconv.Float32ToJs(verticesNative)
	// Create buffer
	vBuffer := gl.CreateBuffer()
	// Bind to buffer
	gl.BindBuffer(webgl.ARRAY_BUFFER, vBuffer)
	// Pass data to buffer
	gl.BufferData2(webgl.ARRAY_BUFFER, webgl.UnionFromJS(vertices), webgl.STATIC_DRAW)
	// Unbind buffer
	gl.BindBuffer(webgl.ARRAY_BUFFER, &webgl.Buffer{})

	// INDEX BUFFER ////
	var indicesNative = []uint32{
		2, 1, 0,
	}
	var indices = jsconv.UInt32ToJs(indicesNative)

	// Create buffer
	iBuffer := gl.CreateBuffer()

	// Bind to buffer
	gl.BindBuffer(webgl.ELEMENT_ARRAY_BUFFER, iBuffer)

	// Pass data to buffer
	gl.BufferData2(webgl.ELEMENT_ARRAY_BUFFER, webgl.UnionFromJS(indices), webgl.STATIC_DRAW)

	// Unbind buffer
	gl.BindBuffer(webgl.ELEMENT_ARRAY_BUFFER, &webgl.Buffer{})
	return vBuffer, iBuffer, len(indicesNative)
}

func setupShaders(gl *webgl.RenderingContext) *webgl.Program {
	// Vertex shader source code
	vertCode := `
  attribute vec3 coordinates;
  uniform mat4 modelview;

  void main(void) {
      gl_Position = modelview * vec4(coordinates, 1.0);
  }
	`

	// Create a vertex shader object
	vShader := gl.CreateShader(webgl.VERTEX_SHADER)

	// Attach vertex shader source code
	gl.ShaderSource(vShader, vertCode)

	// Compile the vertex shader
	gl.CompileShader(vShader)

	//fragment shader source code
	fragCode := `
	void main(void) {
		gl_FragColor = vec4(0.0, 1.0, 0.0, 0.7);
	}`

	// Create fragment shader object
	fShader := gl.CreateShader(webgl.FRAGMENT_SHADER)

	// Attach fragment shader source code
	gl.ShaderSource(fShader, fragCode)

	// Compile the fragmentt shader
	gl.CompileShader(fShader)

	// Create a shader program object to store
	// the combined shader program
	prog := gl.CreateProgram()

	// Attach a vertex shader
	gl.AttachShader(prog, vShader)

	// Attach a fragment shader
	gl.AttachShader(prog, fShader)

	// Link both the programs
	gl.LinkProgram(prog)

	// Use the combined shader program object
	gl.UseProgram(prog)

	// Get the location of the model-view matrix in the vertex shader
	modelviewLoc := gl.GetUniformLocation(prog, "modelview")

	// Set the initial model-view matrix
	rotationMatrix := getRotationMatrix(angle)
	gl.UniformMatrix4fv(modelviewLoc, false, rotationMatrix)

	return prog
}

func getRotationMatrix(angle float32) *webgl.Union {
	// Create a 2D rotation matrix for the given angle
	cosA := float32(math.Cos(float64(angle)))
	sinA := float32(math.Sin(float64(angle)))

	return webgl.UnionFromJS(jsconv.Float32ToJs([]float32{
		cosA, -sinA, 0, 0,
		sinA, cosA, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}))
}
