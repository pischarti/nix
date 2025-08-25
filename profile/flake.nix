{
  description = "Profile";

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
              nil
              nixd
              
              ungit
              
              qemu
              docker
              podman
              podman-tui
              
              kind
              k3d
              k9s

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

              awscli2
              ssm-session-manager-plugin

              azure-cli
              google-cloud-sdk

              jq
              git
              gh
              
              neovim
              xclip
              
              certbot
              python313Packages.certbot-dns-route53
              
              kratos
              ory
              
              codex
            ];
          };
        });
    };
}
