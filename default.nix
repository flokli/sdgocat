{
  nixpkgs ? <nixpkgs>,
  pkgs ? import nixpkgs { },
  lib ? pkgs.lib,
}:

pkgs.buildGoModule {
  pname = "sdgocat";
  version = "0.0.0";
  src =
    let
      fs = lib.fileset;
    in
    fs.toSource {
      root = ./.;
      fileset = fs.unions [
        ./main.go
        ./go.sum
        ./go.mod
      ];
    };
  vendorHash = null;
}
