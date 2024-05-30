{pkgs, ...}:
pkgs.dockerTools.buildImage {
  name = "plmercereau/hello";
  tag = "latest";

  config = {Cmd = ["${pkgs.hello}/bin/hello"];};

  created = "now";
}
