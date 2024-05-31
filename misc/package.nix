/*
usage: esp-monitor = import ./esp-monitor/package.nix {inherit pkgs;};
pkgs.writeBinScriptSomething if possible
export GSM_PIN='\"1234\"'
${pkgs.platformio}/bin/platformio
...and ideally, include the source in the script rather than using a separate mkDerivation...
# install and include the platformio dependencies, too
*/
{pkgs, ...}: let
  sourceDir = "/source";

  source = pkgs.stdenv.mkDerivation {
    name = "source";
    src = ./.;
    phases = ["unpackPhase" "installPhase"];
    buildInputs = [pkgs.platformio];
    buildPhase = ''
      echo "Installing platformio dependencies"
      export PLATFORMIO_CORE_DIR=.platformio
      ${pkgs.platformio}/bin/pio pkg install
      ${pkgs.platformio}/bin/pio platform install espressif32
      echo "DONE"
    '';
    installPhase = ''
      mkdir -p $out/source
      cp -r $src/* $out/source
    '';
  };

  builder =
    pkgs.writeShellScriptBin "builder"
    ''
      export GSM_PIN='\"1234\"'
      echo $GSM_PIN
      ${pkgs.platformio}/bin/pio run --environment esp32dev
    '';
in
  pkgs.dockerTools.buildImage {
    name = "plmercereau/hello";
    tag = "latest";
    created = "now";
    runAsRoot = ''
      #!${pkgs.runtimeShell}
      ${pkgs.dockerTools.shadowSetup}
      groupadd -r user
      useradd -r -g user user
      ${pkgs.coreutils}/bin/chown user:user /source
    '';

    copyToRoot = pkgs.buildEnv {
      name = "platformio";
      paths = [
        pkgs.platformio
        # pkgs.coreutils
        builder
        source
      ];
      pathsToLink = ["/bin" "/source" "/usr/bin"];
    };

    config = {
      Cmd = ["/bin/builder"];
      WorkingDir = sourceDir;
      Env = ["PLATFORMIO_CORE_DIR=/source/.platformio"];
      User = "user";
      # Volumes = { "/data" = { }; };
    };
  }
