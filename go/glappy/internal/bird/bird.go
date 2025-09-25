package bird

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	// Bird properties
	BirdSize      = 30
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
	wingCycle float64 // For wing flapping animation
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
		wingCycle: 0,
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

	// Update wing flapping animation
	b.wingCycle += 0.3
	if b.wingCycle > 6.28 { // 2*pi
		b.wingCycle = 0
	}
}

// GetRect returns the bird's collision rectangle
func (b *Bird) GetRect() (x, y, width, height float64) {
	return b.X - float64(b.Size/2), b.Y - float64(b.Size/2),
		float64(b.Size), float64(b.Size)
}

// Draw draws the bird on the screen
func (b *Bird) Draw(screen *ebiten.Image) {
	// Draw bird body (yellow circle)
	vector.DrawFilledCircle(screen, float32(b.X), float32(b.Y),
		float32(b.Size/2), color.RGBA{255, 255, 0, 255}, false)

	// Draw flapping wings (orange circles)
	wingRadius := float32(6.0)
	// Wing flapping animation based on sine wave
	wingOffset := float32(math.Sin(b.wingCycle)) * 3.0

	// Left wing
	vector.DrawFilledCircle(screen, float32(b.X-8), float32(b.Y-2)+wingOffset,
		wingRadius, color.RGBA{255, 165, 0, 255}, false)
	// Right wing
	vector.DrawFilledCircle(screen, float32(b.X+8), float32(b.Y-2)-wingOffset,
		wingRadius, color.RGBA{255, 165, 0, 255}, false)

	// Draw simple eye (black circle)
	eyeSize := float32(2.0)
	vector.DrawFilledCircle(screen, float32(b.X+5), float32(b.Y-5),
		eyeSize, color.RGBA{0, 0, 0, 255}, false)
}
