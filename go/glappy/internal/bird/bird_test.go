package bird

import (
	"testing"
)

func TestBirdInitialization(t *testing.T) {
	b := NewBird(100, 200)

	if b.X != 100 {
		t.Errorf("Expected bird X to be 100, got %f", b.X)
	}
	if b.Y != 200 {
		t.Errorf("Expected bird Y to be 200, got %f", b.Y)
	}
	if b.Velocity != 0 {
		t.Errorf("Expected bird velocity to be 0, got %f", b.Velocity)
	}
	if b.Size != BirdSize {
		t.Errorf("Expected bird size to be %d, got %d", BirdSize, b.Size)
	}
}

func TestBirdJump(t *testing.T) {
	b := NewBird(100, 200)
	initialVelocity := b.Velocity

	b.Jump()

	if b.Velocity != BirdJumpSpeed {
		t.Errorf("Expected bird velocity to be %f after jump, got %f", float64(BirdJumpSpeed), b.Velocity)
	}
	if b.Velocity == initialVelocity {
		t.Error("Bird velocity should change after jump")
	}
}

func TestBirdUpdate(t *testing.T) {
	b := NewBird(100, 200)
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
	b := NewBird(100, 200)
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

func TestBirdConstants(t *testing.T) {
	if BirdSize <= 0 {
		t.Error("BirdSize should be positive")
	}
	if BirdJumpSpeed >= 0 {
		t.Error("BirdJumpSpeed should be negative (upward)")
	}
	if BirdGravity <= 0 {
		t.Error("BirdGravity should be positive (downward)")
	}
}
