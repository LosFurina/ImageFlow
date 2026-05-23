## Purpose

Ensure the ImageFlow project operates as a fully independent repository under the LosFurina GitHub namespace, with no residual dependencies on the original upstream.

## Requirements

### Requirement: independent-module-path
The Go module SHALL use `github.com/LosFurina/ImageFlow` as its canonical import path, and all Go source files SHALL import from this path exclusively.

#### Scenario: Go build succeeds with new module path
- **WHEN** `go build ./...` is executed after the module path change
- **THEN** the build completes successfully with no import errors

#### Scenario: All imports updated
- **WHEN** scanning all `.go` files in the project
- **THEN** no file contains the string `github.com/Yuri-NagaSaki/ImageFlow`

### Requirement: detached-github-repo
The GitHub repository SHALL have no fork parent relationship, appearing as an independent project rather than a fork.

#### Scenario: GitHub API confirms no parent
- **WHEN** querying `GET /repos/LosFurina/ImageFlow` via GitHub API
- **THEN** the response contains no `parent` field linking to `Yuri-NagaSaki/ImageFlow`

#### Scenario: No sync fork button
- **WHEN** viewing the repository on github.com
- **THEN** the "Sync fork" button is no longer displayed

### Requirement: updated-container-images
Docker Compose configurations SHALL reference container images under the LosFurina namespace or a neutral name, not the upstream author's Docker Hub account.

#### Scenario: docker-compose references updated
- **WHEN** reviewing all docker-compose YAML files
- **THEN** no image reference points to `soyorins/imageflow-*`

### Requirement: updated-documentation
All documentation files SHALL reference the LosFurina repository URL and namespace, not the original upstream.

#### Scenario: README points to LosFurina
- **WHEN** reading README.md and README_CN.md
- **THEN** all GitHub URLs reference `github.com/LosFurina/ImageFlow` and no URLs reference `Yuri-NagaSaki/ImageFlow`
