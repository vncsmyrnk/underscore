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

      elvish-tap = pkgs.stdenv.mkDerivation {
        pname = "elvish-tap";
        version = "main";

        src = pkgs.fetchFromGitHub {
          owner = "tesujimath";
          repo = "elvish-tap";
          rev = "main";
          hash = "sha256-4M3Kh814aQ0Sv075G2q8DsKCZDFx7Hi5B0kJ7OApAPg=";
        };

        installPhase = ''
          mkdir -p $out/share/elvish/lib/github.com/tesujimath/elvish-tap
          cp -r * $out/share/elvish/lib/github.com/tesujimath/elvish-tap/
        '';
      };

      default = pkgs.stdenv.mkDerivation {
        pname = "underscore";
        version = "0.0.0";

        src = pkgs.lib.fileset.toSource {
          root = ./.;
          fileset = pkgs.lib.fileset.unions [
            ./scripts
            ./Makefile
            ./entrypoints
            ./completions
            ./underscore.elv
            ./t
          ];
        };

        nativeBuildInputs = with pkgs; [ elvish ];

        buildInputs = with pkgs; [ elvish ];

        installPhase = ''
          make install PREFIX=$out
        '';

        doCheck = true;
        nativeCheckInputs = with pkgs; [
          coreutils
          elvish
          perl
          yq
          elvish-tap
        ];

        preCheck = ''
          patchShebangs t/ scripts/
          export XDG_DATA_DIRS="${elvish-tap}/share:''${XDG_DATA_DIRS:-}"
        '';
      };

      devShell = pkgs.mkShell {
        packages = with pkgs; [
          coreutils
          elvish
          perl
          yq
          elvish-tap
        ];
      };
    in
    {
      packages.${system}.default = default;
      devShells.${system}.default = devShell;
    };
}
