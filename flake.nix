{
  description = "A collection of wrappers for commonly used commands.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};

      default = pkgs.stdenv.mkDerivation {
        pname = "underscore";
        version = "0.0.0";

        src = pkgs.lib.fileset.toSource {
          root = ./.;
          fileset = pkgs.lib.fileset.unions [
            ./scripts
            ./underscore.elv
            ./Makefile
            ./entrypoints
            ./completions
          ];
        };

        nativeBuildInputs = with pkgs; [ elvish ];

        buildInputs = with pkgs; [ elvish ];

        installPhase = ''
          make install PREFIX=$out
        '';

      };

      devShell = pkgs.mkShell {
        packages = with pkgs; [
          bash
          coreutils
          elvish
        ];
      };
    in
    {
      packages.${system}.default = default;
      devShells.${system}.default = devShell;
    };
}
