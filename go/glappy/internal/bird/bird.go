package bird

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	// Bird properties
	BirdSize      = 20
	BirdJumpSpeed = -8
	BirdGravity   = 0.5
)

// Bird represents the player character
type Bird struct {
	X, Y      float64
	Velocity  float64
	Size      int
	gravity   float64
	jumpSpeed float64
}

// NewBird creates a new bird at the specified position
func NewBird(x, y float64) *Bird {
	return &Bird{
		X:         x,
		Y:         y,
		Velocity:  0,
		Size:      BirdSize,
		gravity:   BirdGravity,
		jumpSpeed: BirdJumpSpeed,
	}
}

// Jump makes the bird jump
func (b *Bird) Jump() {
	b.Velocity = b.jumpSpeed
}

// Update updates the bird's position based on velocity and gravity
func (b *Bird) Update() {
	b.Velocity += b.gravity
	b.Y += b.Velocity
}

// GetRect returns the bird's collision rectangle
func (b *Bird) GetRect() (x, y, width, height float64) {
	return b.X - float64(b.Size/2), b.Y - float64(b.Size/2),
		float64(b.Size), float64(b.Size)
}

// Draw draws the bird on the screen
func (b *Bird) Draw(screen *ebiten.Image) {
	// Draw bird body (yellow rectangle)
	vector.DrawFilledRect(screen, float32(b.X-float64(b.Size/2)), float32(b.Y-float64(b.Size/2)),
		float32(b.Size), float32(b.Size), color.RGBA{255, 255, 0, 255}, false)

	// Draw simple eye (black rectangle)
	eyeSize := float32(3.0)
	vector.DrawFilledRect(screen, float32(b.X+5)-eyeSize/2, float32(b.Y-5)-eyeSize/2,
		eyeSize, eyeSize, color.RGBA{0, 0, 0, 255}, false)
}
