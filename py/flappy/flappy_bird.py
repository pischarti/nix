#!/usr/bin/env python3
"""
Flappy Bird Game
A simple implementation of the classic Flappy Bird game using Pygame.
"""

import pygame
import random
import sys
from typing import List, Tuple


class Bird:
    """Represents the bird character in the game."""
    
    def __init__(self, x: int, y: int):
        self.x = x
        self.y = y
        self.velocity = 0
        self.gravity = 0.5
        self.jump_strength = -8
        self.size = 20
        
    def jump(self):
        """Make the bird jump."""
        self.velocity = self.jump_strength
        
    def update(self):
        """Update bird position based on velocity and gravity."""
        self.velocity += self.gravity
        self.y += self.velocity
        
    def draw(self, screen: pygame.Surface):
        """Draw the bird on the screen."""
        pygame.draw.circle(screen, (255, 255, 0), (int(self.x), int(self.y)), self.size)
        # Draw a simple eye
        pygame.draw.circle(screen, (0, 0, 0), (int(self.x + 5), int(self.y - 5)), 3)
        
    def get_rect(self) -> pygame.Rect:
        """Get the bird's collision rectangle."""
        return pygame.Rect(self.x - self.size, self.y - self.size, 
                          self.size * 2, self.size * 2)


class Pipe:
    """Represents a pipe obstacle."""
    
    def __init__(self, x: int, gap_y: int, gap_size: int = 150):
        self.x = x
        self.gap_y = gap_y
        self.gap_size = gap_size
        self.width = 50
        self.speed = 3
        
    def update(self):
        """Move the pipe to the left."""
        self.x -= self.speed
        
    def draw(self, screen: pygame.Surface, screen_height: int):
        """Draw the pipe on the screen."""
        # Top pipe
        top_height = self.gap_y - self.gap_size // 2
        pygame.draw.rect(screen, (0, 255, 0), 
                        (self.x, 0, self.width, top_height))
        
        # Bottom pipe
        bottom_y = self.gap_y + self.gap_size // 2
        bottom_height = screen_height - bottom_y
        pygame.draw.rect(screen, (0, 255, 0), 
                        (self.x, bottom_y, self.width, bottom_height))
        
    def get_top_rect(self) -> pygame.Rect:
        """Get the top pipe's collision rectangle."""
        top_height = self.gap_y - self.gap_size // 2
        return pygame.Rect(self.x, 0, self.width, top_height)
        
    def get_bottom_rect(self) -> pygame.Rect:
        """Get the bottom pipe's collision rectangle."""
        bottom_y = self.gap_y + self.gap_size // 2
        return pygame.Rect(self.x, bottom_y, self.width, 400)  # Assume max height


class Game:
    """Main game class that manages the game state and logic."""
    
    def __init__(self):
        pygame.init()
        self.width = 800
        self.height = 600
        self.screen = pygame.display.set_mode((self.width, self.height))
        pygame.display.set_caption("Flappy Bird")
        self.clock = pygame.time.Clock()
        
        # Game objects
        self.bird = Bird(100, self.height // 2)
        self.pipes: List[Pipe] = []
        self.score = 0
        self.game_over = False
        self.font = pygame.font.Font(None, 36)
        
        # Game settings
        self.pipe_spawn_distance = 300
        
    def spawn_pipe(self):
        """Spawn a new pipe."""
        gap_y = random.randint(150, self.height - 150)
        self.pipes.append(Pipe(self.width, gap_y))
        
    def update(self):
        """Update game state."""
        if self.game_over:
            return
            
        # Update bird
        self.bird.update()
        
        # Check if bird hits ground or ceiling
        if self.bird.y > self.height or self.bird.y < 0:
            self.game_over = True
            
        # Spawn new pipes
        if len(self.pipes) == 0 or self.pipes[-1].x < self.width - self.pipe_spawn_distance:
            self.spawn_pipe()
            
        # Update pipes
        for pipe in self.pipes[:]:
            pipe.update()
            
            # Check collision with bird
            bird_rect = self.bird.get_rect()
            if (bird_rect.colliderect(pipe.get_top_rect()) or 
                bird_rect.colliderect(pipe.get_bottom_rect())):
                self.game_over = True
                
            # Remove pipes that are off screen
            if pipe.x + pipe.width < 0:
                self.pipes.remove(pipe)
                self.score += 1
                
    def draw(self):
        """Draw all game elements."""
        # Clear screen with sky blue background
        self.screen.fill((135, 206, 235))
        
        # Draw pipes
        for pipe in self.pipes:
            pipe.draw(self.screen, self.height)
            
        # Draw bird
        self.bird.draw(self.screen)
        
        # Draw score
        score_text = self.font.render(f"Score: {self.score}", True, (0, 0, 0))
        self.screen.blit(score_text, (10, 10))
        
        # Draw game over screen
        if self.game_over:
            game_over_text = self.font.render("GAME OVER! Press R to restart", True, (255, 0, 0))
            text_rect = game_over_text.get_rect(center=(self.width // 2, self.height // 2))
            self.screen.blit(game_over_text, text_rect)
            
        pygame.display.flip()
        
    def handle_events(self):
        """Handle user input events."""
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                return False
            elif event.type == pygame.KEYDOWN:
                if event.key == pygame.K_SPACE and not self.game_over:
                    self.bird.jump()
                elif event.key == pygame.K_r and self.game_over:
                    self.restart()
                elif event.key == pygame.K_ESCAPE:
                    return False
        return True
        
    def restart(self):
        """Restart the game."""
        self.bird = Bird(100, self.height // 2)
        self.pipes = []
        self.score = 0
        self.game_over = False
        
    def run(self):
        """Main game loop."""
        running = True
        while running:
            running = self.handle_events()
            self.update()
            self.draw()
            self.clock.tick(60)  # 60 FPS
            
        pygame.quit()
        sys.exit()


def main():
    """Entry point of the application."""
    print("🐦 Starting Flappy Bird Game!")
    print("Controls:")
    print("  SPACE - Jump")
    print("  R - Restart (when game over)")
    print("  ESC - Quit")
    print()
    
    game = Game()
    game.run()


if __name__ == "__main__":
    main()