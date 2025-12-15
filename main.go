package main

import (
	"runtime"
	"time"

	"go-3d-engine/camera"
	"go-3d-engine/gfx"
	"go-3d-engine/loaders"
	"go-3d-engine/scene"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width     = 1280
	height    = 720
	targetFPS = 60
)

func main() {
	runtime.LockOSThread()
	window := gfx.InitGlfw(width, height, "Go 3D Engine")
	defer glfw.Terminate()
	glfw.SwapInterval(0)
	program := gfx.InitOpenGL()

	// 0.5, 0.8, 0.9 is a nice "Sky Blue"
	gl.ClearColor(0.5, 0.8, 0.9, 1.0)

	// --- 1. LOAD TEXTURES ---

	// Texture for Fences/Benches
	atlasTexture, err := gfx.NewTexture("colormap.png")
	if err != nil {
		panic(err)
	}

	// Texture for the Barrel body
	barrelTexture, err := gfx.NewTexture("barrel.png")
	if err != nil {
		panic(err)
	}

	// Texture for the Barrel stand (planks)
	planksTexture, err := gfx.NewTexture("planks.png") // <--- NEW
	if err != nil {
		panic(err)
	}

	// Texture for the Ground
	groundTexture, err := gfx.NewTexture("ground.png") // <-- New ground texture
	if err != nil {
		panic(err)
	}

	// --- 2. LOAD MESHES ---

	// BENCH (Single Material)
	benchRaw := loaders.LoadObj("bench.obj")
	var benchVAO uint32
	var benchCount int32
	for _, data := range benchRaw {
		benchVAO, benchCount = gfx.CreateMesh(data)
		break
	}

	// FENCE (Single Material)
	fenceRaw := loaders.LoadObj("fence.obj")
	var fenceVAO uint32
	var fenceCount int32
	for _, data := range fenceRaw {
		fenceVAO, fenceCount = gfx.CreateMesh(data)
		break
	}

	// Ground
	groundRaw := loaders.LoadObj("ground.obj")
	var groundVAO uint32
	var groundCount int32
	for _, data := range groundRaw {
		groundVAO, groundCount = gfx.CreateMesh(data)
		break
	}

	// BARRELS (Multi Material: "barrel" and "planks")
	barrelRaw := loaders.LoadObj("barrels.obj")

	// Mesh for the round barrel part
	barrelPartVAO, barrelPartCount := gfx.CreateMesh(barrelRaw["barrel"])

	// Mesh for the wooden stand part
	planksPartVAO, planksPartCount := gfx.CreateMesh(barrelRaw["planks"])

	// --- 3. BUILD SCENE ---
	var objects []scene.GameObject

	// Add Fences (Background)
	for i := -2; i <= 2; i++ {
		objects = append(objects, scene.NewGameObject(
			fenceVAO, fenceCount, atlasTexture,
			mgl32.Vec3{float32(i) * 2.0, 0, -2}, 0,
		))
	}

	// Add Benches
	objects = append(objects, scene.NewGameObject(
		benchVAO, benchCount, atlasTexture,
		mgl32.Vec3{0, 0, 1}, mgl32.DegToRad(180),
	))

	// Add Ground
	objects = append(objects, scene.NewGameObject(
		groundVAO, groundCount, groundTexture,
		mgl32.Vec3{0, 0, 0}, 0,
	))

	// --- ADD BARRELS ---
	addBarrel := func(pos mgl32.Vec3, rot float32) {
		// 1. The Body -> uses barrel.png
		objects = append(objects, scene.NewGameObject(
			barrelPartVAO, barrelPartCount, barrelTexture,
			pos, rot,
		))
		// 2. The Stand -> uses planks.png (Fixed)
		objects = append(objects, scene.NewGameObject(
			planksPartVAO, planksPartCount, planksTexture, // <--- Using specific texture
			pos, rot,
		))
	}

	addBarrel(mgl32.Vec3{1.5, 0, 1}, 0)
	addBarrel(mgl32.Vec3{-3.5, 0, -1.5}, mgl32.DegToRad(90))
	addBarrel(mgl32.Vec3{3, 0, 2}, mgl32.DegToRad(123))

	// --- 4. CAMERA & UNIFORMS ---
	cam := camera.New(mgl32.Vec3{0, 2, 6})

	gl.UseProgram(program)
	modelLoc := gl.GetUniformLocation(program, gl.Str("model\x00"))
	viewLoc := gl.GetUniformLocation(program, gl.Str("view\x00"))
	projLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))

	lightDirLoc := gl.GetUniformLocation(program, gl.Str("lightDir\x00"))
	gl.Uniform3fv(lightDirLoc, 1, &[]float32{-0.2, -0.5, -0.3}[0])

	ambientColorLoc := gl.GetUniformLocation(program, gl.Str("ambientColor\x00"))
	ambientStrengthLoc := gl.GetUniformLocation(program, gl.Str("ambientStrength\x00"))

	gl.Uniform3f(ambientColorLoc, 1.0, 1.0, 1.0) // white ambient
	gl.Uniform1f(ambientStrengthLoc, 0.25)       // tweak 0.15–0.35

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(width)/float32(height), 0.1, 100.0)
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])
	gl.Uniform1i(gl.GetUniformLocation(program, gl.Str("texture1\x00")), 0)

	lastTime := glfw.GetTime()
	targetFrameTime := 1.0 / float64(targetFPS)

	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := currentTime - lastTime
		lastTime = currentTime

		cam.HandleKeyboard(window, deltaTime)
		cam.HandleMouse(window)

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.UseProgram(program)

		view := cam.GetViewMatrix()
		gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])

		for _, obj := range objects {
			if obj.VertexCount == 0 {
				continue
			}

			gl.BindVertexArray(obj.Vao)
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, obj.Texture)

			translate := mgl32.Translate3D(obj.Position.X(), obj.Position.Y(), obj.Position.Z())
			rotate := mgl32.HomogRotate3D(obj.Rotation, mgl32.Vec3{0, 1, 0})
			model := translate.Mul4(rotate)

			gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])
			gl.DrawArrays(gl.TRIANGLES, 0, obj.VertexCount)
		}

		glfw.PollEvents()
		window.SwapBuffers()

		if workTime := glfw.GetTime() - currentTime; workTime < targetFrameTime {
			time.Sleep(time.Duration((targetFrameTime - workTime) * float64(time.Second)))
		}
	}
}
