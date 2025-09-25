package game

import (
	"testing"

	"github.com/pischarti/nix/go/glappy/internal/bird"
)

func TestBirdInitialization(t *testing.T) {
	b := bird.NewBird(100, 200)

	if b.X != 100 {
		t.Errorf("Expected bird X to be 100, got %f", b.X)
	}
	if b.Y != 200 {
		t.Errorf("Expected bird Y to be 200, got %f", b.Y)
	}
	if b.Velocity != 0 {
		t.Errorf("Expected bird velocity to be 0, got %f", b.Velocity)
	}
	if b.Size != bird.BirdSize {
		t.Errorf("Expected bird size to be %d, got %d", bird.BirdSize, b.Size)
	}
}

func TestBirdJump(t *testing.T) {
	b := bird.NewBird(100, 200)
	b.Jump()

	if b.Velocity != bird.BirdJumpSpeed {
		t.Errorf("Expected bird velocity to be %f after jump, got %f", float64(bird.BirdJumpSpeed), b.Velocity)
	}
}

func TestBirdUpdate(t *testing.T) {
	b := bird.NewBird(100, 200)
	initialY := b.Y
	initialVelocity := b.Velocity

	b.Update()

	// Bird should fall due to gravity
	if b.Y <= initialY {
		t.Error("Bird should fall due to gravity")
	}
	if b.Velocity <= initialVelocity {
		t.Error("Bird velocity should increase due to gravity")
	}
}

func TestBirdGetRect(t *testing.T) {
	b := bird.NewBird(100, 200)
	x, y, width, height := b.GetRect()

	expectedX := b.X - float64(b.Size/2)
	expectedY := b.Y - float64(b.Size/2)
	expectedSize := float64(b.Size)

	if x != expectedX {
		t.Errorf("Expected rect X to be %f, got %f", expectedX, x)
	}
	if y != expectedY {
		t.Errorf("Expected rect Y to be %f, got %f", expectedY, y)
	}
	if width != expectedSize {
		t.Errorf("Expected rect width to be %f, got %f", expectedSize, width)
	}
	if height != expectedSize {
		t.Errorf("Expected rect height to be %f, got %f", expectedSize, height)
	}
}

func TestPipeInitialization(t *testing.T) {
	pipe := NewPipe(100, 200)

	if pipe.X != 100 {
		t.Errorf("Expected pipe X to be 100, got %f", pipe.X)
	}
	if pipe.GapY != 200 {
		t.Errorf("Expected pipe GapY to be 200, got %f", pipe.GapY)
	}
	if pipe.GapSize != PipeGapSize {
		t.Errorf("Expected pipe GapSize to be %d, got %d", PipeGapSize, pipe.GapSize)
	}
	if pipe.Width != PipeWidth {
		t.Errorf("Expected pipe Width to be %d, got %d", PipeWidth, pipe.Width)
	}
	if pipe.Passed {
		t.Error("New pipe should not be marked as passed")
	}
}

func TestPipeUpdate(t *testing.T) {
	pipe := NewPipe(100, 200)
	initialX := pipe.X

	pipe.Update()

	expectedX := initialX - PipeSpeed
	if pipe.X != expectedX {
		t.Errorf("Expected pipe X to be %f after update, got %f", expectedX, pipe.X)
	}
}

func TestPipeGetTopRect(t *testing.T) {
	pipe := NewPipe(100, 200)
	x, y, width, height := pipe.GetTopRect()

	expectedHeight := pipe.GapY - float64(pipe.GapSize/2)
	if x != pipe.X {
		t.Errorf("Expected top rect X to be %f, got %f", pipe.X, x)
	}
	if y != 0 {
		t.Errorf("Expected top rect Y to be 0, got %f", y)
	}
	if width != float64(pipe.Width) {
		t.Errorf("Expected top rect width to be %f, got %f", float64(pipe.Width), width)
	}
	if height != expectedHeight {
		t.Errorf("Expected top rect height to be %f, got %f", expectedHeight, height)
	}
}

func TestPipeGetBottomRect(t *testing.T) {
	pipe := NewPipe(100, 200)
	x, y, width, height := pipe.GetBottomRect()

	expectedY := pipe.GapY + float64(pipe.GapSize/2)
	expectedHeight := ScreenHeight - expectedY
	if x != pipe.X {
		t.Errorf("Expected bottom rect X to be %f, got %f", pipe.X, x)
	}
	if y != expectedY {
		t.Errorf("Expected bottom rect Y to be %f, got %f", expectedY, y)
	}
	if width != float64(pipe.Width) {
		t.Errorf("Expected bottom rect width to be %f, got %f", float64(pipe.Width), width)
	}
	if height != expectedHeight {
		t.Errorf("Expected bottom rect height to be %f, got %f", expectedHeight, height)
	}
}

func TestGameStateInitialization(t *testing.T) {
	state := NewGameState()

	if state.Bird == nil {
		t.Error("GameState should have a bird")
	}
	if state.Pipes == nil {
		t.Error("GameState should have a pipes slice")
	}
	if len(state.Pipes) != 0 {
		t.Errorf("Expected initial pipes length to be 0, got %d", len(state.Pipes))
	}
	if state.Score != 0 {
		t.Errorf("Expected initial score to be 0, got %d", state.Score)
	}
	if state.GameOver {
		t.Error("Game should not be over initially")
	}
	if state.LastSpawn != 0 {
		t.Errorf("Expected initial LastSpawn to be 0, got %f", state.LastSpawn)
	}
}

func TestGameStateRestart(t *testing.T) {
	state := NewGameState()

	// Modify the state
	state.Score = 10
	state.GameOver = true
	state.LastSpawn = 100
	state.Pipes = append(state.Pipes, NewPipe(50, 100))

	// Restart
	state.Restart()

	// Check that everything is reset
	if state.Score != 0 {
		t.Errorf("Expected score to be 0 after restart, got %d", state.Score)
	}
	if state.GameOver {
		t.Error("Game should not be over after restart")
	}
	if state.LastSpawn != 0 {
		t.Errorf("Expected LastSpawn to be 0 after restart, got %f", state.LastSpawn)
	}
	if len(state.Pipes) != 0 {
		t.Errorf("Expected pipes to be empty after restart, got %d", len(state.Pipes))
	}
	if state.Bird.X != BirdStartX {
		t.Errorf("Expected bird X to be %f after restart, got %f", float64(BirdStartX), state.Bird.X)
	}
	if state.Bird.Y != BirdStartY {
		t.Errorf("Expected bird Y to be %f after restart, got %f", float64(BirdStartY), state.Bird.Y)
	}
}

func TestGameConstants(t *testing.T) {
	// Test screen dimensions
	if ScreenWidth <= 0 {
		t.Error("ScreenWidth should be positive")
	}
	if ScreenHeight <= 0 {
		t.Error("ScreenHeight should be positive")
	}

	// Test pipe properties
	if PipeWidth <= 0 {
		t.Error("PipeWidth should be positive")
	}
	if PipeGapSize <= 0 {
		t.Error("PipeGapSize should be positive")
	}
	if PipeSpeed <= 0 {
		t.Error("PipeSpeed should be positive")
	}
	if PipeSpawnDist <= 0 {
		t.Error("PipeSpawnDist should be positive")
	}

	// Test bird properties
	if BirdStartX <= 0 {
		t.Error("BirdStartX should be positive")
	}
	if BirdStartY <= 0 {
		t.Error("BirdStartY should be positive")
	}
}

func TestGameInitialization(t *testing.T) {
	g := NewGame()

	if g.GameState == nil {
		t.Error("Game should have a GameState")
	}
	// Note: font field is unexported, so we can't test it directly
}
