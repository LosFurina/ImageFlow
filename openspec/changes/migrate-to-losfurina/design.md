## Context

The project was forked from `Yuri-NagaSaki/ImageFlow` on GitHub. While the local repository has been cloned independently with no upstream remote, two critical areas still reference the original namespace:

1. **Go module path** — `go.mod` declares `module github.com/Yuri-NagaSaki/ImageFlow`, and every Go source file imports packages under this path (14+ files affected).
2. **GitHub fork metadata** — GitHub's API still shows a `parent` field linking to the upstream repo, causing the "Sync fork" UI.
3. **Docker images** — `docker-compose.yaml` pulls `soyorins/imageflow-backend:latest` from the upstream author's Docker Hub.
4. **Documentation** — README files may reference the upstream repository URL.

Current state:
- Local repo: `~/Desktop/workspace/repos/ImageFlow-LosFurina`
- Remote: `git@github.com:LosFurina/ImageFlow.git` (SSH)
- `gh` CLI authenticated as LosFurina with full permissions
- OpenSpec v1.3.1 initialized

## Goals / Non-Goals

**Goals:**
- Change Go module path to `github.com/LosFurina/ImageFlow` across all files
- Update all Go import paths to match the new module
- Detach the GitHub fork relationship
- Update Docker image references to a neutral or LosFurina namespace
- Update documentation to reference LosFurina/ImageFlow
- Ensure the project builds and passes validation after migration

**Non-Goals:**
- No feature changes — this is purely a namespace/repo restructuring
- No changes to API endpoints, data formats, or runtime behavior
- Not publishing new Docker images (that's a separate deployment concern)
- Not modifying the original upstream repository

## Decisions

1. **Module path**: `github.com/LosFurina/ImageFlow` — follows Go convention of matching the GitHub repository URL, enabling `go get` to work correctly.

2. **Import migration strategy**: Use `find + sed` for bulk replacement across all `.go` files, then verify with `go build ./...`. This is safer than manual editing for 14+ files.

3. **Fork detachment**: Use GitHub API (`DELETE /repos/{owner}/{repo}/fork`) via `gh api` to sever the fork parent relationship. This is irreversible but desired.

4. **Docker images**: Change `soyorins/imageflow-*` to `losfurina/imageflow-*` in docker-compose files. The images won't exist yet on Docker Hub, but the build compose file (`docker-compose.build.yaml`) builds locally, so it will still work.

5. **Documentation**: Replace all `Yuri-NagaSaki/ImageFlow` references with `LosFurina/ImageFlow` in README.md, README_CN.md, and CLAUDE.md.

6. **Execution order**: Go module migration first (local, testable), then documentation, then Docker, then GitHub fork detachment (remote, irreversible, last).

## Risks / Trade-offs

- **Fork detachment is irreversible**: Once the fork parent is removed, it cannot be restored. If we ever need to sync from upstream again, we'd need to add it as a manual remote. This is acceptable per the user's explicit requirement for independence.
- **Docker image names**: Changing to `losfurina/imageflow-*` means pre-built images won't exist until we push them. The `docker-compose.build.yaml` workflow still works (builds locally). Users relying on `docker-compose.yaml` (pre-built images) will need to build locally until images are published.
- **Go module path change**: Any downstream consumers (if any) would need to update their imports. Since this is a personal fork with no known downstream users, the risk is minimal.
- **go.sum changes**: After changing the module path, `go mod tidy` will regenerate go.sum. This is expected and correct.
