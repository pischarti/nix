{
  description = "Flappy Bird Game with Python, Nix, and uv";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        python = pkgs.python311;
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            python
            uv
            # Development tools
            git
            just
          ];

          shellHook = ''
            echo "🐦 Welcome to Flappy Bird Development Environment!"
            echo "Python version: $(python --version)"
            echo "uv version: $(uv --version)"
            echo ""
            echo "To get started:"
            echo "  uv sync          # Install dependencies"
            echo "  uv run python flappy_bird.py  # Run the game"
            echo ""
          '';
        };

        packages.default = pkgs.stdenv.mkDerivation {
          name = "flappy-bird";
          src = ./.;
          buildInputs = [ python pkgs.uv ];
          buildPhase = ''
            uv sync --frozen
          '';
          installPhase = ''
            mkdir -p $out/bin
            cp -r . $out/
            echo '#!/bin/sh' > $out/bin/flappy-bird
            echo 'cd $out && uv run python flappy_bird.py' >> $out/bin/flappy-bird
            chmod +x $out/bin/flappy-bird
          '';
        };
      });
}