## Why

The project was originally forked from `Yuri-NagaSaki/ImageFlow`. We are making this repository fully independent under the `LosFurina` GitHub account. Currently:

1. **Go module path still references the upstream**: All Go source files import `github.com/Yuri-NagaSaki/ImageFlow/...`, which ties the codebase to the original author's namespace. This breaks independence and can cause confusion when publishing or importing.
2. **GitHub still treats this as a fork**: The remote repository shows a "Sync fork" button because GitHub maintains a parent-child fork relationship. This makes the repo appear as a dependent copy rather than a standalone project.

These issues must be resolved before any further development to ensure a clean, independent codebase.

## What Changes

- **Go module path migration**: Change `module github.com/Yuri-NagaSaki/ImageFlow` to `module github.com/LosFurina/ImageFlow` in `go.mod`, and update all import paths across every Go source file.
- **GitHub fork detachment**: Sever the fork relationship on GitHub so the repository is no longer linked to `Yuri-NagaSaki/ImageFlow` as its parent.
- **Docker image references**: Update any container image names or references that point to the old namespace (e.g., `soyorins/imageflow-backend:latest`).
- **Documentation updates**: Update README, CLAUDE.md, and any other docs that reference the old repository URL or author.

## Capabilities

### New Capabilities
- `independent-module-path`: The Go module uses `github.com/LosFurina/ImageFlow` as its canonical import path, making the project self-contained and publishable under the LosFurina namespace.
- `detached-github-repo`: The GitHub repository has no fork parent, appearing as an independent project.

### Modified Capabilities
- `architecture`: Module path changes affect the project's import structure and build system.

## Impact

- **All Go source files**: Import paths change from `github.com/Yuri-NagaSaki/ImageFlow/...` to `github.com/LosFurina/ImageFlow/...`
- **go.mod / go.sum**: Module declaration and indirect references
- **Docker configurations**: Image names in docker-compose files
- **Documentation**: README, CLAUDE.md, README_CN.md
- **GitHub repository**: Fork relationship removed via GitHub API (irreversible)
- **No API changes**: This is purely an internal restructuring; no endpoints, data formats, or user-facing behavior changes
