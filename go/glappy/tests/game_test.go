package tests

import (
	"testing"

	"github.com/pischarti/nix/go/glappy/internal/game"
)

func TestBirdInitialization(t *testing.T) {
	bird := game.NewBird(100, 200)

	if bird.X != 100 {
		t.Errorf("Expected bird X to be 100, got %f", bird.X)
	}
	if bird.Y != 200 {
		t.Errorf("Expected bird Y to be 200, got %f", bird.Y)
	}
	if bird.Velocity != 0 {
		t.Errorf("Expected bird velocity to be 0, got %f", bird.Velocity)
	}
	if bird.Size != 20 {
		t.Errorf("Expected bird size to be 20, got %d", bird.Size)
	}
}

func TestBirdJump(t *testing.T) {
	bird := game.NewBird(100, 200)
	bird.Jump()

	if bird.Velocity != -8 {
		t.Errorf("Expected bird velocity to be -8 after jump, got %f", bird.Velocity)
	}
}

func TestBirdUpdate(t *testing.T) {
	bird := game.NewBird(100, 200)
	initialY := bird.Y
	bird.Update()

	expectedY := initialY + 0.5 // gravity
	if bird.Y != expectedY {
		t.Errorf("Expected bird Y to be %f after update, got %f", expectedY, bird.Y)
	}
}

func TestBirdGetRect(t *testing.T) {
	bird := game.NewBird(100, 200)
	x, y, width, height := bird.GetRect()

	expectedX := 90.0  // 100 - 20/2
	expectedY := 190.0 // 200 - 20/2
	expectedWidth := 20.0
	expectedHeight := 20.0

	if x != expectedX {
		t.Errorf("Expected bird rect X to be %f, got %f", expectedX, x)
	}
	if y != expectedY {
		t.Errorf("Expected bird rect Y to be %f, got %f", expectedY, y)
	}
	if width != expectedWidth {
		t.Errorf("Expected bird rect width to be %f, got %f", expectedWidth, width)
	}
	if height != expectedHeight {
		t.Errorf("Expected bird rect height to be %f, got %f", expectedHeight, height)
	}
}

func TestPipeInitialization(t *testing.T) {
	pipe := game.NewPipe(300, 250)

	if pipe.X != 300 {
		t.Errorf("Expected pipe X to be 300, got %f", pipe.X)
	}
	if pipe.GapY != 250 {
		t.Errorf("Expected pipe gap Y to be 250, got %f", pipe.GapY)
	}
	if pipe.GapSize != 150 {
		t.Errorf("Expected pipe gap size to be 150, got %d", pipe.GapSize)
	}
	if pipe.Width != 50 {
		t.Errorf("Expected pipe width to be 50, got %d", pipe.Width)
	}
	if pipe.Speed() != 3 {
		t.Errorf("Expected pipe speed to be 3, got %f", pipe.Speed())
	}
}

func TestPipeUpdate(t *testing.T) {
	pipe := game.NewPipe(300, 250)
	initialX := pipe.X
	pipe.Update()

	expectedX := initialX - 3 // speed
	if pipe.X != expectedX {
		t.Errorf("Expected pipe X to be %f after update, got %f", expectedX, pipe.X)
	}
}

func TestPipeGetTopRect(t *testing.T) {
	pipe := game.NewPipe(300, 250)
	x, y, width, height := pipe.GetTopRect()

	expectedX := 300.0
	expectedY := 0.0
	expectedWidth := 50.0
	expectedHeight := 175.0 // 250 - 150/2

	if x != expectedX {
		t.Errorf("Expected top pipe rect X to be %f, got %f", expectedX, x)
	}
	if y != expectedY {
		t.Errorf("Expected top pipe rect Y to be %f, got %f", expectedY, y)
	}
	if width != expectedWidth {
		t.Errorf("Expected top pipe rect width to be %f, got %f", expectedWidth, width)
	}
	if height != expectedHeight {
		t.Errorf("Expected top pipe rect height to be %f, got %f", expectedHeight, height)
	}
}

func TestPipeGetBottomRect(t *testing.T) {
	pipe := game.NewPipe(300, 250)
	x, y, width, height := pipe.GetBottomRect()

	expectedX := 300.0
	expectedY := 325.0 // 250 + 150/2
	expectedWidth := 50.0
	expectedHeight := 275.0 // 600 - 325

	if x != expectedX {
		t.Errorf("Expected bottom pipe rect X to be %f, got %f", expectedX, x)
	}
	if y != expectedY {
		t.Errorf("Expected bottom pipe rect Y to be %f, got %f", expectedY, y)
	}
	if width != expectedWidth {
		t.Errorf("Expected bottom pipe rect width to be %f, got %f", expectedWidth, width)
	}
	if height != expectedHeight {
		t.Errorf("Expected bottom pipe rect height to be %f, got %f", expectedHeight, height)
	}
}

func TestGameInitialization(t *testing.T) {
	gameState := game.NewGameState()

	if gameState.Score != 0 {
		t.Errorf("Expected game score to be 0, got %d", gameState.Score)
	}
	if gameState.GameOver != false {
		t.Errorf("Expected game over to be false, got %t", gameState.GameOver)
	}
	if len(gameState.Pipes) != 0 {
		t.Errorf("Expected game pipes to be empty, got %d pipes", len(gameState.Pipes))
	}
}

func TestGameRestart(t *testing.T) {
	gameState := game.NewGameState()

	// Modify game state
	gameState.Score = 10
	gameState.GameOver = true

	// Restart
	gameState.Restart()

	if gameState.Score != 0 {
		t.Errorf("Expected game score to be 0 after restart, got %d", gameState.Score)
	}
	if gameState.GameOver != false {
		t.Errorf("Expected game over to be false after restart, got %t", gameState.GameOver)
	}
	if len(gameState.Pipes) != 0 {
		t.Errorf("Expected game pipes to be empty after restart, got %d pipes", len(gameState.Pipes))
	}
}
