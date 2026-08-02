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
      go = pkgs.go_1_26;

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
            ./.golangci.yml
            ./go.mod
            ./internal
            ./shell
            ./completions
            ./underscore
            ./t
          ];
        };

        nativeBuildInputs = with pkgs; [
          elvish
          go
          gotools
          golangci-lint
          zsh
        ];

        buildInputs = with pkgs; [
          elvish
          go
          gotools
          golangci-lint
          zsh
        ];

        installPhase = ''
          make install PREFIX=$out
        '';

        doCheck = true;
        nativeCheckInputs = with pkgs; [
          coreutils
          elvish
          go
          gotools
          golangci-lint
          perl
          yq
          zsh
          elvish-tap
        ];

        preCheck = ''
          patchShebangs t/
          export XDG_DATA_DIRS="${elvish-tap}/share:''${XDG_DATA_DIRS:-}"
        '';
      };

      devShell = pkgs.mkShell {
        packages = with pkgs; [
          coreutils
          elvish
          go
          gotools
          golangci-lint
          perl
          yq
          zsh
          elvish-tap
        ];

        shellHook = ''
          export ZDOTDIR="$PWD/shell/rc"
        '';
      };
    in
    {
      packages.${system}.default = default;
      devShells.${system}.default = devShell;
    };
}
