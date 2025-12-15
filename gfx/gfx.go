package gfx

import (
	"embed"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg" // Import image decoders
	_ "image/png"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

//go:embed shaders/*
var shaderFS embed.FS

func InitGlfw(width, height int, title string) *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	return window
}

func InitOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Enable(gl.MULTISAMPLE)

	// 1. Enable Depth Testing (Z-buffer)
	gl.Enable(gl.DEPTH_TEST)

	// 2. Enable Back Face Culling
	gl.Enable(gl.CULL_FACE)

	// Optional: Configure it (These are the defaults)
	gl.CullFace(gl.BACK) // Cull the back faces
	gl.FrontFace(gl.CCW) // Counter-Clockwise vertices are "Front"

	vertSrc, errV := shaderFS.ReadFile("shaders/main.vert")
	fragSrc, errF := shaderFS.ReadFile("shaders/main.frag")

	if errV != nil {
		panic("Failed to load main.vert: " + errV.Error())
	}
	if errF != nil {
		panic("Failed to load main.frag: " + errF.Error())
	}

	return createProgram(string(vertSrc)+"\x00", string(fragSrc)+"\x00")
}

func CreateMesh(points []float32) (uint32, int32) {
	var vbo, vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	// STRIDE: Pos(3) + Normal(3) + UV(2) = 8 floats * 4 bytes = 32 bytes
	stride := int32(8 * 4)

	// 1. Position (Location 0, Offset 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))

	// 2. Normal (Location 1, Offset 3 floats = 12 bytes)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(12))

	// 3. Texture (Location 2, Offset 6 floats = 24 bytes)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(24))

	// Return Vertex Count (Total floats / 8 floats per vertex)
	return vao, int32(len(points) / 8)
}

func NewTexture(file string) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	defer imgFile.Close() // Good practice to close the file!

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}

	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	// --- FIX: FLIP IMAGE VERTICALLY ---
	// OpenGL expects (0,0) to be bottom-left.
	// Standard Images are top-left. We must swap the rows.
	height := rgba.Rect.Dy()
	stride := rgba.Stride

	// Create a temporary buffer to hold one row of pixels
	tempRow := make([]uint8, stride)

	for y := 0; y < height/2; y++ {
		// Calculate the indices for the top and bottom rows
		topIndex := y * stride
		bottomIndex := (height - 1 - y) * stride

		// Get references to the slices
		topRow := rgba.Pix[topIndex : topIndex+stride]
		bottomRow := rgba.Pix[bottomIndex : bottomIndex+stride]

		// Swap them using the temp buffer
		copy(tempRow, topRow)
		copy(topRow, bottomRow)
		copy(bottomRow, tempRow)
	}
	// ----------------------------------

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	gl.GenerateMipmap(gl.TEXTURE_2D)

	return texture, nil
}

func createProgram(vertSource, fragSource string) uint32 {
	vShader, _ := compileShader(vertSource, gl.VERTEX_SHADER)
	fShader, _ := compileShader(fragSource, gl.FRAGMENT_SHADER)

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vShader)
	gl.AttachShader(prog, fShader)
	gl.LinkProgram(prog)
	return prog
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("compile failed: %v", log)
	}
	return shader, nil
}
