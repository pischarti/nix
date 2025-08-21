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

              qemu
              
              docker
              podman
              podman-tui
              
              kind
              k3d

              zarf
              kubernetes-helm
              fluxcd
              kustomize_4
              kubelogin
              kubetui
              kubent
              kubecm
              kubexit
              kubepug
              kubefwd
              kubectx
              kubectl
              kubetail
              kubeseal
              kubeshark
              kubeswitch
              kubebuilder
              kubectl-ktop
              kubectl-graph
              kube-capacity
              kubectl-images
              kubemq-community 
                        
              nodejs
              pnpm_8
              supabase-cli
          
              go

              php

              python314
              uv

              awscli2
              ssm-session-manager-plugin

              azure-cli
              google-cloud-sdk

              pulumi

              jq
              git
              gh
              neovim
              xclip
              certbot
            ];
          };
        });
    };
}
