package game

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	// Screen dimensions
	ScreenWidth  = 800
	ScreenHeight = 600

	// Bird properties
	BirdSize      = 20
	BirdJumpSpeed = -8
	BirdGravity   = 0.5
	BirdStartX    = 100
	BirdStartY    = ScreenHeight / 2

	// Pipe properties
	PipeWidth     = 50
	PipeGapSize   = 150
	PipeSpeed     = 3
	PipeSpawnDist = 300
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

// Pipe represents a pipe obstacle
type Pipe struct {
	X       float64
	GapY    float64
	GapSize int
	Width   int
	speed   float64
	Passed  bool
}

// NewPipe creates a new pipe at the specified position
func NewPipe(x, gapY float64) *Pipe {
	return &Pipe{
		X:       x,
		GapY:    gapY,
		GapSize: PipeGapSize,
		Width:   PipeWidth,
		speed:   PipeSpeed,
		Passed:  false,
	}
}

// Speed returns the pipe's speed
func (p *Pipe) Speed() float64 {
	return p.speed
}

// Update moves the pipe to the left
func (p *Pipe) Update() {
	p.X -= p.speed
}

// GetTopRect returns the top pipe's collision rectangle
func (p *Pipe) GetTopRect() (x, y, width, height float64) {
	topHeight := p.GapY - float64(p.GapSize/2)
	return p.X, 0, float64(p.Width), topHeight
}

// GetBottomRect returns the bottom pipe's collision rectangle
func (p *Pipe) GetBottomRect() (x, y, width, height float64) {
	bottomY := p.GapY + float64(p.GapSize/2)
	bottomHeight := ScreenHeight - bottomY
	return p.X, bottomY, float64(p.Width), bottomHeight
}

// Draw draws the pipe on the screen
func (p *Pipe) Draw(screen *ebiten.Image) {
	// Top pipe
	topHeight := p.GapY - float64(p.GapSize/2)
	vector.DrawFilledRect(screen, float32(p.X), 0, float32(p.Width), float32(topHeight),
		color.RGBA{0, 255, 0, 255}, false)

	// Bottom pipe
	bottomY := p.GapY + float64(p.GapSize/2)
	bottomHeight := ScreenHeight - bottomY
	vector.DrawFilledRect(screen, float32(p.X), float32(bottomY), float32(p.Width), float32(bottomHeight),
		color.RGBA{0, 255, 0, 255}, false)
}

// GameState represents the main game state (for testing)
type GameState struct {
	Bird      *Bird
	Pipes     []*Pipe
	Score     int
	GameOver  bool
	LastSpawn float64
}

// NewGameState creates a new game state instance
func NewGameState() *GameState {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	return &GameState{
		Bird:      NewBird(BirdStartX, BirdStartY),
		Pipes:     make([]*Pipe, 0),
		Score:     0,
		GameOver:  false,
		LastSpawn: 0,
	}
}

// Restart resets the game to initial state
func (g *GameState) Restart() {
	g.Bird = NewBird(BirdStartX, BirdStartY)
	g.Pipes = make([]*Pipe, 0)
	g.Score = 0
	g.GameOver = false
	g.LastSpawn = 0
}
