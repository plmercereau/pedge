{pkgs, ...}:
pkgs.dockerTools.buildImage {
  name = "plmercereau/hello";
  # tag = "0.1.0";

  config = {Cmd = ["${pkgs.hello}/bin/hello"];};

  created = "now";
}
