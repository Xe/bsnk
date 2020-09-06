let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs { };
  nur = import (builtins.fetchTarball
    "https://github.com/nix-community/NUR/archive/master.tar.gz") {
      inherit pkgs;
    };
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    go goimports golint nur.repos.xe.gopls
  ];

  PORT = "8000";
}
