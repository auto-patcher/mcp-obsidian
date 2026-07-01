{
  description = "MCP server for Obsidian via the Local REST API";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    {
      lib.settingsModule = { };
    }
    // flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            git
          ];

          shellHook = ''
            echo "MCP Obsidian dev environment loaded"
            echo "Go version: $(go version)"
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "mcp-obsidian";
          version = "0.3.0";
          src = ./.;

          # Run `nix build` once to get the real hash, then replace this.
          vendorHash = pkgs.lib.fakeHash;

          meta = {
            description = "MCP server for Obsidian via the Local REST API";
            longDescription = ''
              An MCP (Model Context Protocol) server that bridges the Go/CLI ecosystem
              with Obsidian, allowing scripts, tools, and AI agents to interact with an
              Obsidian vault through the Local REST API plugin.
            '';
            homepage = "https://github.com/auto-patcher/mcp-obsidian";
            license = pkgs.lib.licenses.mit;
            platforms = pkgs.lib.platforms.all;
            mainProgram = "mcp-obsidian";
          };
        };

        formatter = pkgs.nixfmt;
      }
    );
}
