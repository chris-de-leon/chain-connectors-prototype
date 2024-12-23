# https://github.com/NixOS/nixpkgs/commits/master
{
  inputs = {
    nixpkgs.url = "https://github.com/NixOS/nixpkgs/archive/4bbb73beb26f5152a87d3460e6bf76227841ae21.tar.gz";
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
            pkgs.goreleaser
            pkgs.go_1_23
            pkgs.nodejs
            pkgs.redis
          ];
        };
      }
    );
}
