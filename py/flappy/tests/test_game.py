"""
Tests for the Flappy Bird game.
"""

import pytest
import pygame
from flappy_bird import Bird, Pipe, Game


class TestBird:
    """Test cases for the Bird class."""
    
    def test_bird_initialization(self):
        """Test that a bird initializes with correct properties."""
        bird = Bird(100, 200)
        assert bird.x == 100
        assert bird.y == 200
        assert bird.velocity == 0
        assert bird.gravity == 0.5
        assert bird.jump_strength == -8
        assert bird.size == 20
        
    def test_bird_jump(self):
        """Test that bird jump sets correct velocity."""
        bird = Bird(100, 200)
        bird.jump()
        assert bird.velocity == -8
        
    def test_bird_update(self):
        """Test that bird update applies gravity correctly."""
        bird = Bird(100, 200)
        initial_y = bird.y
        bird.update()
        assert bird.y == initial_y + bird.gravity
        
    def test_bird_get_rect(self):
        """Test that bird get_rect returns correct rectangle."""
        bird = Bird(100, 200)
        rect = bird.get_rect()
        assert rect.x == 80  # 100 - 20
        assert rect.y == 180  # 200 - 20
        assert rect.width == 40  # 20 * 2
        assert rect.height == 40  # 20 * 2


class TestPipe:
    """Test cases for the Pipe class."""
    
    def test_pipe_initialization(self):
        """Test that a pipe initializes with correct properties."""
        pipe = Pipe(300, 250)
        assert pipe.x == 300
        assert pipe.gap_y == 250
        assert pipe.gap_size == 150
        assert pipe.width == 50
        assert pipe.speed == 3
        
    def test_pipe_update(self):
        """Test that pipe update moves it left."""
        pipe = Pipe(300, 250)
        initial_x = pipe.x
        pipe.update()
        assert pipe.x == initial_x - pipe.speed
        
    def test_pipe_get_top_rect(self):
        """Test that pipe get_top_rect returns correct rectangle."""
        pipe = Pipe(300, 250)
        rect = pipe.get_top_rect()
        expected_height = 250 - 150 // 2  # gap_y - gap_size // 2
        assert rect.x == 300
        assert rect.y == 0
        assert rect.width == 50
        assert rect.height == expected_height
        
    def test_pipe_get_bottom_rect(self):
        """Test that pipe get_bottom_rect returns correct rectangle."""
        pipe = Pipe(300, 250)
        rect = pipe.get_bottom_rect()
        expected_y = 250 + 150 // 2  # gap_y + gap_size // 2
        assert rect.x == 300
        assert rect.y == expected_y
        assert rect.width == 50
        assert rect.height == 400


class TestGame:
    """Test cases for the Game class."""
    
    def test_game_initialization(self):
        """Test that game initializes with correct properties."""
        # Mock pygame.init to avoid display issues in tests
        pygame.init = lambda: None
        pygame.display.set_mode = lambda size: None
        pygame.display.set_caption = lambda caption: None
        pygame.time.Clock = lambda: None
        pygame.font.Font = lambda font, size: None
        
        game = Game()
        assert game.width == 800
        assert game.height == 600
        assert game.score == 0
        assert game.game_over == False
        assert len(game.pipes) == 0
        
    def test_game_restart(self):
        """Test that game restart resets all properties."""
        # Mock pygame.init to avoid display issues in tests
        pygame.init = lambda: None
        pygame.display.set_mode = lambda size: None
        pygame.display.set_caption = lambda caption: None
        pygame.time.Clock = lambda: None
        pygame.font.Font = lambda font, size: None
        
        game = Game()
        game.score = 10
        game.game_over = True
        game.pipes = [Pipe(100, 200)]
        
        game.restart()
        
        assert game.score == 0
        assert game.game_over == False
        assert len(game.pipes) == 0
        assert game.bird.x == 100
        assert game.bird.y == game.height // 2


if __name__ == "__main__":
    pytest.main([__file__])