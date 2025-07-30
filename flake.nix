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
              
              zarf
              kubernetes-helm
              fluxcd
              kustomize_4
                        
              nodejs
              pnpm_8
          
              go

              php

              python314
              uv

              awscli2
              azure-cli
              google-cloud-sdk

              pulumi
              terraform

              jq
              git
              gh
              neovim
              xclip
            ];
          };
        });
    };
}
