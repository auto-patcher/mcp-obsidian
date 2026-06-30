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
        python = pkgs.python312;
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            python
            pkgs.uv
            pkgs.git
            (python.pkgs.pyright or pkgs.pyright)
            python.pkgs.pytest
            python.pkgs.pytest-cov
          ];

          shellHook = ''
            echo "MCP Obsidian dev environment loaded"
            echo "Python version: $(python --version)"
            echo "uv version: $(uv --version)"
          '';
        };

        packages.default = python.pkgs.buildPythonPackage rec {
          pname = "mcp-obsidian";
          version = "0.2.3";
          pyproject = true;
          src = ./.;

          build-system = with python.pkgs; [
            hatchling
          ];

          dependencies = with python.pkgs; [
            mcp
            python-dotenv
            requests
          ];

          nativeCheckInputs = with python.pkgs; [
            pytest
            pytest-cov
          ];

          checkPhase = ''
            python -m pytest
          '';

          meta = {
            description = "MCP server for Obsidian via the Local REST API";
            longDescription = ''
              An MCP (Model Context Protocol) server that bridges the Python/CLI ecosystem
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
