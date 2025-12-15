package scene

import "github.com/go-gl/mathgl/mgl32"

// GameObject represents a renderable item in the world
type GameObject struct {
	Vao         uint32
	VertexCount int32
	Texture     uint32

	Position mgl32.Vec3
	Rotation float32 // Radians around Y axis
}

// NewGameObject is a helper to create objects cleaner
func NewGameObject(vao uint32, count int32, tex uint32, pos mgl32.Vec3, rot float32) GameObject {
	return GameObject{
		Vao:         vao,
		VertexCount: count,
		Texture:     tex,
		Position:    pos,
		Rotation:    rot,
	}
}
