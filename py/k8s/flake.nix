{
  description = "K8s utilities dev environment (Nix + uv)";

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
            python311Packages.kubernetes
            uv
            kubectl
            # Optional tooling
            just
            git
          ];

          shellHook = ''
            echo "🚀 K8s utilities dev shell"
            echo "Python: $(python --version)"
            echo "uv: $(uv --version)"
            echo "kubectl: $(kubectl version --client --short 2>/dev/null || true)"
            echo
            echo "Common commands:"
            echo "  uv sync"
            echo "  uv run image-list -- --counts"
          '';
        };

        packages.default = pkgs.stdenv.mkDerivation {
          name = "k8s-utils";
          src = ./.;
          buildInputs = [ python pkgs.uv ];
          buildPhase = ''
            uv sync --frozen
          '';
          installPhase = ''
            mkdir -p $out
            cp -r . $out/
          '';
        };
      });
}


