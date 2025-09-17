{
  description = "Development environment";

  inputs = { nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable"; };

  outputs = { self, nixpkgs }:
    let
      inherit (nixpkgs.lib) genAttrs;
      supportedSystems = [
        "aarch64-darwin"
        "x86_64-darwin"
        "x86_64-linux"
      ];
      forAllSystems = f: genAttrs supportedSystems (system: f system);
    in {
      devShells = forAllSystems (system:
        let pkgs = import nixpkgs { 
          inherit system; 
          config.allowUnfree = true; 
        };
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              hello
              cowsay
              lolcat
              git
                        
              nodejs
              pnpm_8
              supabase-cli
          
              go

              php

              python313
              python313Packages.uvicorn
              python313Packages.fastapi
              python313Packages.pydantic
              python313Packages.httpx
              python313Packages.boto3
              python313Packages.click
              python313Packages.rich
              uv
              awscli2
              just
            ];
          };
        });
    };
}
