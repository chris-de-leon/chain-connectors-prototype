# https://github.com/NixOS/nixpkgs/commits/master
{
  inputs = {
    nixpkgs.url = "https://github.com/NixOS/nixpkgs/archive/be77f97455fc7f17d3deabe790d7a7a1c3cdd899.tar.gz";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem(system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.grpc-tools
            pkgs.go_1_23
            pkgs.redis
          ];
        };
      }
    );
}
