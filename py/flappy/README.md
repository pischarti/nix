# 🐦 Flappy Bird Game

A classic Flappy Bird game implementation using Python, Pygame, Nix flakes, and uv for dependency management.

## 📋 Project Summary

This project delivers a complete Flappy Bird game with modern Python development practices:

### **Core Files Created:**

1. **`flake.nix`** - Nix flake configuration that provides:
   - Python 3.11 environment
   - uv package manager
   - Development tools (git, just)
   - Both development shell and buildable package

2. **`pyproject.toml`** - uv configuration with:
   - Pygame dependency for the game
   - Development dependencies (pytest, black, flake8)
   - Project metadata and build configuration

3. **`flappy_bird.py`** - Complete game implementation featuring:
   - Bird physics with gravity and jumping
   - Randomly generated pipe obstacles
   - Collision detection
   - Score tracking
   - Game over and restart functionality
   - Clean, colorful graphics

4. **`Justfile`** - Convenient commands for:
   - Installing dependencies
   - Running the game
   - Code formatting and linting
   - Testing
   - Development shell access

5. **`tests/test_game.py`** - Unit tests for all game components

6. **`.gitignore`** - Proper exclusions for Python, Nix, and development files

### **Key Features:**

- **Modern Python**: Uses Python 3.11+ with type hints
- **Clean Architecture**: Object-oriented design with separate Bird, Pipe, and Game classes
- **Smooth Gameplay**: 60 FPS with proper physics simulation
- **Development Ready**: Includes testing, linting, and formatting tools
- **Nix Integration**: Reproducible development environment
- **uv Management**: Fast Python package management

## 🎮 Game Features

- Classic Flappy Bird gameplay
- Smooth physics with gravity and jumping
- Randomly generated pipe obstacles
- Score tracking
- Game over and restart functionality
- Clean, colorful graphics

## 🎯 Controls

- **SPACE** - Make the bird jump
- **R** - Restart the game (when game over)
- **ESC** - Quit the game

## 🛠️ Setup and Installation

### Prerequisites

- [Nix](https://nixos.org/download.html) installed on your system
- [uv](https://github.com/astral-sh/uv) (automatically provided by the Nix flake)

### Quick Start

1. **Enter the development environment:**
   ```bash
   nix develop
   ```

2. **Install dependencies:**
   ```bash
   uv sync
   ```

3. **Run the game:**
   ```bash
   uv run python flappy_bird.py
   ```

### Alternative: Using Just Commands

If you have [Just](https://github.com/casey/just) installed, you can use these convenient commands:

```bash
# Install dependencies
just install

# Run the game
just run

# Enter development shell
just shell

# Format code
just format

# Lint code
just lint

# Run tests
just test

# Clean up
just clean
```

## 🏗️ Project Structure

```
py-sp/
├── flake.nix              # Nix flake configuration
├── pyproject.toml         # Python project configuration with uv
├── flappy_bird.py         # Main game implementation
├── Justfile               # Convenient commands
└── README.md              # This file
```

## 🧪 Development

### Code Quality

The project includes development tools for maintaining code quality:

- **Black** - Code formatting
- **Flake8** - Linting
- **Pytest** - Testing framework

Run these tools with:
```bash
uv run black .          # Format code
uv run flake8 .         # Lint code
uv run pytest           # Run tests
```

### Adding Dependencies

To add new dependencies:

1. Add them to `pyproject.toml` under `dependencies` or `[project.optional-dependencies].dev`
2. Run `uv sync` to update the environment

## 🎨 Game Architecture

The game is built with a clean, object-oriented design:

- **`Bird`** - Handles bird physics, rendering, and collision detection
- **`Pipe`** - Manages pipe obstacles, movement, and collision rectangles
- **`Game`** - Main game loop, state management, and event handling

## 🚀 Building with Nix

To build a standalone version of the game:

```bash
nix build
```

This creates a result symlink with the built game.

## 🐛 Troubleshooting

### Common Issues

1. **Nix not found**: Make sure Nix is installed and your shell is configured properly
2. **uv not found**: The Nix flake provides uv automatically - make sure you're in the development shell
3. **Pygame issues**: All dependencies are managed by uv and should work automatically

### Getting Help

- Check that you're in the Nix development shell: `nix develop`
- Verify uv is available: `uv --version`
- Check Python version: `python --version`

## 📝 License

This project is open source. Feel free to modify and distribute as needed.

## 🎉 Have Fun!

Enjoy playing Flappy Bird! Try to beat your high score and see how far you can fly! 🐦✨