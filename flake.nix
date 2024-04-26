{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    nixpkgs-devenv.url = "github:cachix/devenv-nixpkgs/rolling";
    flake-utils.url = "github:numtide/flake-utils";
    devenv.url = "github:cachix/devenv";
    devenv.inputs.nixpkgs.follows = "nixpkgs-devenv";
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
    ...
  } @ inputs:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      packages = rec {
        devenv-up = self.devShells.${system}.default.config.procfileScript;

        # TODO move to a separate flake / nix file
        installer = let
          # TODO multi-arch
          # targetSystem = "${pkgs.hostPlatform.uname.processor}-linux";
          targetSystem = "aarch64-linux";
          targetPkgs = nixpkgs.legacyPackages.${targetSystem};
          system =
            (nixpkgs.lib.nixosSystem {
              system = targetSystem;
              modules = [({config, ...}: {system.stateVersion = config.system.nixos.release;})];
            })
            .config
            .system;
          init-script = pkgs.writeScript "init-script" ''
            #! /bin/sh
            mount -t devtmpfs devtmpfs /dev
            mount -t proc none /proc
            mount -t sysfs none /sys

            # Function to configure interface with DHCP
            configure_interface() {
                interface=$1
                echo "Starting DHCP client on $interface..."
                /bin/udhcpc -i $interface -b
            }

            # Loop through all available network interfaces
            for interface in $(ls /sys/class/net); do
                # Exclude the loopback interface
                if [ "$interface" != "lo" ]; then
                    configure_interface $interface
                fi
            done

            echo "Welcome to my minimal BusyBox initrd!"
            exec /bin/sh
          '';
        in
          targetPkgs.runCommand "installer" {} ''
            mkdir -p {bin,sbin,etc,proc,sys,usr/bin,usr/sbin}
            cp ${targetPkgs.busybox}/bin/busybox bin/
            ./bin/busybox --install -s ./bin
            cp -p ${init-script} init

            # Pack everything into an initramfs
            mkdir -p $out/usr/share/installer
            ${targetPkgs.findutils}/bin/find . -print0 | ${targetPkgs.cpio}/bin/cpio --null -ov --format=newc | ${targetPkgs.gzip}/bin/gzip -9 > $out/usr/share/installer/initrd
            cp ${system.build.kernel}/${system.boot.loader.kernelFile} $out/usr/share/installer/vmlinuz
          '';

        # docker load < $(nix build .\#pixie --print-out-paths)
        pixie = let
          # TODO multi-arch
          # TODO not working on aarch64
          targetSystem = "${pkgs.hostPlatform.uname.processor}-linux";
          targetPkgs = nixpkgs.legacyPackages.${targetSystem};
          # TODO https://chat.openai.com/c/a9fe0be9-a84f-4643-ae73-8e83cd64a236
        in
          targetPkgs.dockerTools.buildImage {
            name = "pixie";
            tag = "0.1.0";

            copyToRoot = pkgs.buildEnv {
              name = "pixiecore";
              paths = [targetPkgs.pixiecore installer];
              pathsToLink = ["/bin" "/usr"];
            };

            config = {
              Cmd = [
                "pixiecore"
                "boot"
                "/usr/share/installer/vmlinuz"
                "/usr/share/installer/initrd"
                "--cmdline"
                "init=/init loglevel=4"
                "--debug"
                "--dhcp-no-bind"
                "--port"
                "64172"
                "--status-port"
                "64172"
              ];
            };
          };
      };

      devShells = {
        default = let
          pkgs = nixpkgs-devenv.legacyPackages.${system};
        in
          devenv.lib.mkShell {
            inherit inputs pkgs;
            modules = [
              {
                packages = with pkgs; [
                  go
                  kubectl
                  helmfile
                  (wrapHelm kubernetes-helm {plugins = [kubernetes-helmPlugins.helm-diff];})
                ];
              }
            ];
          };
      };
    });
}
