{
  description = "Generate JSON Schema for Jsonnet packages containing Docsonnet type annotations";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    git-hooks-nix = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    { self, ... }@inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.git-hooks-nix.flakeModule
      ];
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      perSystem =
        {
          pkgs,
          system,
          config,
          ...
        }:
        {
          packages =
            let
              _version = builtins.getEnv "VERSION";
            in
            {
              schemasonnet = pkgs.buildGoModule (finalAttrs: {
                pname = "schemasonnet";
                version = if _version != "" then _version else toString (self.rev or self.dirtyRev or "unknown");
                src = ./.;
                vendorHash = null;
                env.CGO_ENABLED = 0;
                ldflags = [
                  "-s -w -X github.com/squat/schemasonnet/version.Version=${finalAttrs.version}"
                ];

                nativeBuildInputs = [ pkgs.installShellFiles ];

                postInstall = ''
                  installShellCompletion --cmd schemasonnet \
                    --bash <($out/bin/schemasonnet completion bash) \
                    --fish <($out/bin/schemasonnet completion fish) \
                    --zsh <($out/bin/schemasonnet completion zsh)
                '';

                meta = {
                  description = "Generate JSON Schema for Jsonnet packages containing Docsonnet type annotations";
                  mainProgram = "schemasonnet";
                  homepage = "https://github.com/squat/schemasonnet";
                };
              });
            };

          pre-commit = {
            check.enable = true;
            settings = {
              src = ./.;
              hooks = {
                actionlint.enable = true;
                nixfmt-rfc-style.enable = true;
                nixfmt-rfc-style.excludes = [ "vendor" ];
                gofmt.enable = true;
                gofmt.excludes = [ "vendor" ];
                golangci-lint.enable = true;
                golangci-lint.excludes = [ "vendor" ];
                golangci-lint.extraPackages = [ pkgs.go ];
                govet.enable = true;
                govet.excludes = [ "vendor" ];
                readme = {
                  enable = true;
                  name = "README.md";
                  entry =
                    let
                      readmeCheck = pkgs.writeShellApplication {
                        name = "readme-check";
                        text = ''
                          (go run ./... --help || [ $? -eq 1 ]) > help.txt
                          go tool embedmd -d README.md
                        '';
                      };
                    in
                    pkgs.lib.getExe readmeCheck;
                  files = "^README\\.md$";
                  extraPackages = [ pkgs.go ];
                };
              };
            };
          };

          devShells = {
            default = pkgs.mkShell {
              inherit (config.pre-commit.devShell) shellHook;
              packages =
                with pkgs;
                [
                  go
                  (config.packages.schemasonnet.overrideAttrs {
                    version = "dev";
                    __intentionallyOverridingVersion = true;
                  })
                ]
                ++ config.pre-commit.settings.enabledPackages;
            };
          };
        };
    };
}
