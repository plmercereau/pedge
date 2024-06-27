{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    nixpkgs-devenv.url = "github:cachix/devenv-nixpkgs/rolling";
    flake-utils.url = "github:numtide/flake-utils";
    devenv.url = "github:cachix/devenv";
    devenv.inputs.nixpkgs.follows = "nixpkgs-devenv";
    esp-idf.url = "github:mirrexagon/nixpkgs-esp-dev";
    esp-idf.inputs.nixpkgs.follows = "nixpkgs-devenv";
    esp-idf.inputs.flake-utils.follows = "flake-utils";
  };

  nixConfig = {
    extra-trusted-public-keys = "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw=";
    extra-substituters = "https://devenv.cachix.org";
  };

  outputs = {
    self,
    nixpkgs,
    nixpkgs-devenv,
    devenv,
    flake-utils,
    esp-idf,
    ...
  } @ inputs:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      packages = rec {
        # ?
        devenv-up = self.devShells.${system}.default.config.procfileScript;
      };

      devShells = {
        default = let
          pkgs = import nixpkgs-devenv {
            inherit system;
            config.allowUnfree = true;
          };
        in
          devenv.lib.mkShell {
            inherit inputs pkgs;
            modules = [
              {
                pre-commit.hooks.make-device-operator-helm = {
                  enable = true;
                  name = "Generate the Device Operator Helm chart";
                  entry = "bash -c 'cd devices-operator && make helm'";
                  files = "^devices-operator/.*\\.(go|yaml)$";
                  pass_filenames = false;
                };
                packages = with pkgs; [
                  # go_1_21 # * See https://github.com/operator-framework/operator-sdk/issues/6681
                  go
                  kubectl
                  tilt
                  mqttui # MQTT client, for testing purpose
                  platformio-core
                  ngrok
                  kustomize
                  yq-go
                  act
                  (wrapHelm kubernetes-helm {plugins = [kubernetes-helmPlugins.helm-diff];})
                  esp-idf.packages.${system}.esp-idf-full
                  (
                    # operator-sdk package not working
                    # TODO not working for other architectures...
                    pkgs.stdenv.mkDerivation {
                      name = "operator-sdk";
                      src = pkgs.fetchurl {
                        url = "https://github.com/operator-framework/operator-sdk/releases/download/v1.34.2/operator-sdk_darwin_arm64";
                        sha256 = "sha256-NvMvZ80YQsNwDLq6KGxC0tEWt3GiM2Y1YJjBjxChJz0=";
                      };
                      phases = ["installPhase" "patchPhase"];
                      installPhase = ''
                        mkdir -p $out/bin
                        cp $src $out/bin/operator-sdk
                        chmod +x $out/bin/operator-sdk
                      '';
                    }
                  )
                  (
                    # TODO not working for other architectures...
                    pkgs.stdenv.mkDerivation {
                      name = "helmify";
                      src = pkgs.fetchurl {
                        url = "https://github.com/arttor/helmify/releases/download/v0.4.13/helmify_Darwin_arm64.tar.gz";
                        sha256 = "sha256-t4pddkkHCgjyRnhu7xNgkQTB16eaDRsUdUrdSYx2MQo=";
                      };
                      phases = ["installPhase" "patchPhase"];
                      installPhase = ''
                        mkdir -p $out/bin
                        tar xvzf $src -C $out/bin
                      '';
                    }
                  )
                ];
              }
            ];
          };
      };
    });
}
