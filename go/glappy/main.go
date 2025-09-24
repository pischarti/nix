package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/pischarti/nix/go/glappy/internal/game"
)

const (
	screenWidth  = 800
	screenHeight = 600
	fps          = 60
)

// Game represents the main game state
type Game struct {
	*game.GameState
	font font.Face
}

// NewGame creates a new game instance
func NewGame() *Game {
	// Create basic font
	f := basicfont.Face7x13

	return &Game{
		GameState: game.NewGameState(),
		font:      f,
	}
}

// spawnPipe creates a new pipe at the right edge of the screen
func (g *Game) spawnPipe() {
	gapY := float64(rand.Intn(screenHeight-300) + 150)
	g.Pipes = append(g.Pipes, game.NewPipe(float64(screenWidth), gapY))
	g.LastSpawn = float64(screenWidth)
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
	if g.Bird.Y > screenHeight || g.Bird.Y < 0 {
		g.GameOver = true
	}

	// Spawn new pipes
	if len(g.Pipes) == 0 || g.Pipes[len(g.Pipes)-1].X < float64(screenWidth)-game.PipeSpawnDist {
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
		text.Draw(screen, gameOverText, g.font, screenWidth/2-100, screenHeight/2,
			color.RGBA{255, 0, 0, 255})
	}
}

// Layout returns the game's screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	fmt.Println("🐦 Starting Glappy Bird Game!")
	fmt.Println("Controls:")
	fmt.Println("  SPACE - Jump")
	fmt.Println("  R - Restart (when game over)")
	fmt.Println("  ESC - Quit")
	fmt.Println()

	// Set window properties
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Glappy Bird")
	ebiten.SetWindowResizable(false)

	// Create and run game
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
