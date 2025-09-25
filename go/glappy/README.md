# Glappy Bird 🐦

A Go implementation of the classic Flappy Bird game using the Ebiten 2D game library.

## Features

- **Simple Gameplay**: Navigate a bird through pipe obstacles by jumping
- **Collision Detection**: Realistic collision detection between bird and pipes
- **Score System**: Track your progress as you pass through pipes
- **Game Over & Restart**: Restart functionality when you crash
- **Clean Graphics**: Simple but effective visual design

## Controls

- **SPACE** - Make the bird jump
- **R** - Restart the game (when game over)
- **ESC** - Quit the game

## Installation & Running

### Prerequisites

- Go 1.21 or later
- Ebiten v2 game library (managed by root go.mod)

### Running the Game

```bash
# Navigate to the glappy directory
cd go/glappy

# Run the game (uses root go.mod)
go run .

# Or from root directory
cd /path/to/nix
go run ./go/glappy
```

### Building

```bash
# Build the executable (uses root go.mod)
go build -o glappy .

# Run the built executable
./glappy

# Or build from root directory
cd /path/to/nix
go build -o glappy ./go/glappy
```

## Game Mechanics

### Bird Physics
- **Gravity**: Bird falls down continuously
- **Jump**: Press SPACE to make the bird jump upward
- **Velocity**: Bird has realistic physics with acceleration

### Pipe Obstacles
- **Random Generation**: Pipes spawn at random heights
- **Movement**: Pipes move from right to left
- **Gap Size**: Fixed gap size of 150 pixels between top and bottom pipes
- **Collision**: Bird must avoid hitting pipes or screen boundaries

### Scoring
- **Point System**: Earn 1 point for each pipe passed
- **Game Over**: Collision with pipes, ground, or ceiling ends the game

## Architecture

The game is built with a clean, modular architecture:

### Core Components

- **`Bird`**: Player character with physics and collision detection
- **`Pipe`**: Obstacle objects with movement and collision detection
- **`GameState`**: Manages game state, scoring, and object coordination
- **`Game`**: Main game loop and Ebiten integration

### Package Structure

```
glappy/
├── main.go          # Main game loop and Ebiten integration
├── game.go          # Game objects and state management
├── go.mod           # Go module dependencies
├── tests/           # Unit tests
│   └── game_test.go
└── README.md        # This file
```

## Testing

The game includes comprehensive unit tests for all major components:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Test Coverage

Tests cover:
- Bird initialization and movement
- Pipe generation and collision detection
- Game state management
- Score tracking and restart functionality

## Dependencies

- **[Ebiten v2](https://github.com/hajimehoshi/ebiten)**: 2D game library for Go
- **[golang.org/x/image](https://pkg.go.dev/golang.org/x/image)**: Image processing and font support

## Game Configuration

Key game constants can be modified in `main.go`:

```go
const (
    screenWidth   = 800
    screenHeight  = 600
    birdSize      = 20
    birdJumpSpeed = -8
    birdGravity   = 0.5
    pipeWidth     = 50
    pipeGapSize   = 150
    pipeSpeed     = 3
    fps          = 60
)
```

## Comparison with Python Version

This Go implementation provides the same core gameplay as the Python version but with:

- **Better Performance**: Go's compiled nature provides smoother gameplay
- **Cross-Platform**: Single binary runs on Windows, macOS, and Linux
- **Modern Graphics**: Uses Ebiten's hardware-accelerated rendering
- **Type Safety**: Go's static typing prevents many runtime errors
- **Easy Distribution**: Single executable file deployment

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is part of the larger nix repository. See the main LICENSE file for details.
