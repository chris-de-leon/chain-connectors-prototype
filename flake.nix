{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        vers = builtins.replaceStrings [ "\n" ] [ "" ] (builtins.readFile ./VERSION);
        pkgs = import nixpkgs { inherit system; };

        cc = pkgs.buildGo123Module {
          vendorHash = "sha256-OAy1DoQhmv6hbMSjtANBu8jkc7It3GWruzbaKlxB9E8=";
          name = "cc";
          src = pkgs.lib.cleanSourceWith {
            src = ./.;
            filter = path: type:
              let
                rel = pkgs.lib.removePrefix (toString ./. + "/") (toString path);
              in
              builtins.all (prefix: !pkgs.lib.hasPrefix prefix rel) [ "src/examples/" "src/plugins/" ];
          };
          buildFlags = [ "-trimpath" ];
          ldflags = [ "-s" "-w" ];
          env = {
            CGO_ENABLED = 0;
          };
        };

        docker-sandbox = pkgs.dockerTools.buildImage {
          name = "cc";
          tag = "${vers}-sandbox";
          copyToRoot = [
            pkgs.coreutils
            pkgs.bash
            cc
          ];
          config = {
            Cmd = [ "${cc}/bin/cc" ];
          };
        };
      in
      {
        formatter = pkgs.nixpkgs-fmt;

        devShells = {
          default = pkgs.mkShell {
            packages = [
              pkgs.grpc-tools
              pkgs.goreleaser
              pkgs.go_1_23
              pkgs.nodejs
              pkgs.redis
            ];
          };
        };

        packages = {
          docker-sandbox = docker-sandbox;
          cc = cc;
        };
      }
    );
}
