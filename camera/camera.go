package camera

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Camera handles the view matrix and input
type Camera struct {
	Position mgl32.Vec3
	Front    mgl32.Vec3
	Up       mgl32.Vec3
	Right    mgl32.Vec3
	WorldUp  mgl32.Vec3

	Yaw   float64
	Pitch float64

	// Options
	Speed       float32
	Sensitivity float64

	// State for mouse input
	CursorLocked bool
	LastX        float64
	LastY        float64
}

func New(position mgl32.Vec3) *Camera {
	c := &Camera{
		Position:     position,
		WorldUp:      mgl32.Vec3{0, 1, 0},
		Yaw:          -90.0,
		Pitch:        0.0,
		Speed:        5.0,
		Sensitivity:  0.1,
		CursorLocked: false,
	}
	c.updateVectors()
	return c
}

// GetViewMatrix returns the LookAt matrix
func (c *Camera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
}

// HandleKeyboard processes WASD + Space/Shift
func (c *Camera) HandleKeyboard(window *glfw.Window, deltaTime float64) {
	velocity := c.Speed * float32(deltaTime)

	if window.GetKey(glfw.KeyW) == glfw.Press {
		c.Position = c.Position.Add(c.Front.Mul(velocity))
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		c.Position = c.Position.Sub(c.Front.Mul(velocity))
	}
	if window.GetKey(glfw.KeyA) == glfw.Press {
		c.Position = c.Position.Sub(c.Right.Mul(velocity))
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		c.Position = c.Position.Add(c.Right.Mul(velocity))
	}
	if window.GetKey(glfw.KeySpace) == glfw.Press {
		c.Position = c.Position.Add(c.WorldUp.Mul(velocity))
	}
	if window.GetKey(glfw.KeyLeftShift) == glfw.Press {
		c.Position = c.Position.Sub(c.WorldUp.Mul(velocity))
	}
}

// HandleMouse processes rotation
func (c *Camera) HandleMouse(window *glfw.Window) {
	if window.GetMouseButton(glfw.MouseButtonRight) == glfw.Press {
		if !c.CursorLocked {
			// First frame click: Lock and reset
			window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
			c.CursorLocked = true
			x, y := window.GetCursorPos()
			c.LastX, c.LastY = x, y
			return // Skip this frame to prevent snap
		}

		// Calculate delta
		xpos, ypos := window.GetCursorPos()
		xoffset := xpos - c.LastX
		yoffset := c.LastY - ypos // Reversed for Y
		c.LastX, c.LastY = xpos, ypos

		// Sanity check for huge jumps
		if math.Abs(xoffset) > 100 || math.Abs(yoffset) > 100 {
			return
		}

		c.Yaw += xoffset * c.Sensitivity
		c.Pitch += yoffset * c.Sensitivity

		// Constrain Pitch
		if c.Pitch > 89.0 {
			c.Pitch = 89.0
		}
		if c.Pitch < -89.0 {
			c.Pitch = -89.0
		}

		c.updateVectors()

	} else {
		// Release lock
		if c.CursorLocked {
			window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
			c.CursorLocked = false
		}
	}
}

func (c *Camera) updateVectors() {
	// Calculate new Front vector
	front := mgl32.Vec3{
		float32(math.Cos(degToRad(c.Yaw)) * math.Cos(degToRad(c.Pitch))),
		float32(math.Sin(degToRad(c.Pitch))),
		float32(math.Sin(degToRad(c.Yaw)) * math.Cos(degToRad(c.Pitch))),
	}
	c.Front = front.Normalize()

	// Recalculate Right and Up
	c.Right = c.Front.Cross(c.WorldUp).Normalize()
	c.Up = c.Right.Cross(c.Front).Normalize()
}

func degToRad(d float64) float64 {
	return d * (math.Pi / 180.0)
}
