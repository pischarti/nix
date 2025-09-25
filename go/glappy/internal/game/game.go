package game

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/pischarti/nix/go/glappy/internal/bird"
)

const (
	// Screen dimensions
	ScreenWidth  = 800
	ScreenHeight = 600

	// Bird properties
	BirdStartX = 100
	BirdStartY = ScreenHeight / 2

	// Pipe properties
	PipeWidth     = 50
	PipeGapSize   = 150
	PipeSpeed     = 3
	PipeSpawnDist = 300
)

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
	Bird      *bird.Bird
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
		Bird:      bird.NewBird(BirdStartX, BirdStartY),
		Pipes:     make([]*Pipe, 0),
		Score:     0,
		GameOver:  false,
		LastSpawn: 0,
	}
}

// Restart resets the game to initial state
func (g *GameState) Restart() {
	g.Bird = bird.NewBird(BirdStartX, BirdStartY)
	g.Pipes = make([]*Pipe, 0)
	g.Score = 0
	g.GameOver = false
	g.LastSpawn = 0
}

// Game represents the main game instance
type Game struct {
	*GameState
	font font.Face
}

// NewGame creates a new game instance
func NewGame() *Game {
	// Create basic font
	f := basicfont.Face7x13

	return &Game{
		GameState: NewGameState(),
		font:      f,
	}
}

// spawnPipe creates a new pipe at the right edge of the screen
func (g *Game) spawnPipe() {
	gapY := float64(rand.Intn(ScreenHeight-300) + 150)
	g.Pipes = append(g.Pipes, NewPipe(float64(ScreenWidth), gapY))
	g.LastSpawn = float64(ScreenWidth)
}

// Update updates the game state
func (g *Game) Update() error {
	// Handle input
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && !g.GameOver {
		g.Bird.Jump()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) && g.GameOver {
		g.Restart()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if g.GameOver {
		return nil
	}

	// Update bird
	g.Bird.Update()

	// Check if bird hits ground or ceiling
	if g.Bird.Y > ScreenHeight || g.Bird.Y < 0 {
		g.GameOver = true
	}

	// Spawn new pipes
	if len(g.Pipes) == 0 || g.Pipes[len(g.Pipes)-1].X < float64(ScreenWidth)-PipeSpawnDist {
		g.spawnPipe()
	}

	// Update pipes and check collisions
	for i := len(g.Pipes) - 1; i >= 0; i-- {
		pipe := g.Pipes[i]
		pipe.Update()

		// Check collision with bird
		bx, by, bw, bh := g.Bird.GetRect()
		topX, topY, topW, topH := pipe.GetTopRect()
		bottomX, bottomY, bottomW, bottomH := pipe.GetBottomRect()

		if (bx < topX+topW && bx+bw > topX && by < topY+topH && by+bh > topY) ||
			(bx < bottomX+bottomW && bx+bw > bottomX && by < bottomY+bottomH && by+bh > bottomY) {
			g.GameOver = true
		}

		// Remove pipes that are off screen and increment score
		if pipe.X+float64(pipe.Width) < 0 {
			g.Pipes = append(g.Pipes[:i], g.Pipes[i+1:]...)
			if !pipe.Passed {
				g.Score++
				pipe.Passed = true
			}
		}
	}

	return nil
}

// Draw draws the game state
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen with sky blue background
	screen.Fill(color.RGBA{135, 206, 235, 255})

	// Draw pipes
	for _, pipe := range g.Pipes {
		pipe.Draw(screen)
	}

	// Draw bird
	g.Bird.Draw(screen)

	// Draw score
	scoreText := fmt.Sprintf("Score: %d", g.Score)
	text.Draw(screen, scoreText, g.font, 10, 30, color.RGBA{0, 0, 0, 255})

	// Draw game over screen
	if g.GameOver {
		gameOverText := "GAME OVER! Press R to restart"
		text.Draw(screen, gameOverText, g.font, ScreenWidth/2-100, ScreenHeight/2,
			color.RGBA{255, 0, 0, 255})
	}
}

// Layout returns the game's screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// Run starts the game
func Run() {
	fmt.Println("🐦 Starting Glappy Bird Game!")
	fmt.Println("Controls:")
	fmt.Println("  SPACE - Jump")
	fmt.Println("  R - Restart (when game over)")
	fmt.Println("  ESC - Quit")
	fmt.Println()

	// Set window properties
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Glappy Bird")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	// Create and run game
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
