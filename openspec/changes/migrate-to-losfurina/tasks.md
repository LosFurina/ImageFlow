## 1. Go Module Path Migration

- [ ] 1.1 Update `go.mod` module declaration from `github.com/Yuri-NagaSaki/ImageFlow` to `github.com/LosFurina/ImageFlow`
- [ ] 1.2 Replace all Go import paths `github.com/Yuri-NagaSaki/ImageFlow` → `github.com/LosFurina/ImageFlow` in all `.go` files (main.go, handlers/*, utils/*, utils/logger/*, utils/errors/*, scripts/convert.go)
- [ ] 1.3 Run `go mod tidy` to refresh dependencies
- [ ] 1.4 Run `go build ./...` to verify the build succeeds with new module path
- [ ] 1.5 Verify no残留 references: `grep -r "Yuri-NagaSaki" --include="*.go" .`

## 2. Documentation Updates

- [ ] 2.1 Replace all `Yuri-NagaSaki/ImageFlow` references with `LosFurina/ImageFlow` in `README.md`
- [ ] 2.2 Replace all `Yuri-NagaSaki/ImageFlow` references with `LosFurina/ImageFlow` in `README_CN.md`
- [ ] 2.3 Update `CLAUDE.md` if it contains any upstream references
- [ ] 2.4 Verify no残留 references: `grep -r "Yuri-NagaSaki" --include="*.md" .`

## 3. Docker Configuration Updates

- [ ] 3.1 Update `docker-compose.yaml`: change `soyorins/imageflow-backend:latest` to `losfurina/imageflow-backend:latest`
- [ ] 3.2 Check `docker-compose.build.yaml` for any upstream image references and update if needed
- [ ] 3.3 Check `Dockerfile.backend` and `Dockerfile.frontend` for any namespace references

## 4. GitHub Fork Detachment

- [ ] 4.1 Verify current fork status: `gh api repos/LosFurina/ImageFlow --jq '.parent.full_name'`
- [ ] 4.2 Detach fork from parent via GitHub API
- [ ] 4.3 Verify detachment: `gh api repos/LosFurina/ImageFlow` should show no `parent` field

## 5. Final Validation

- [ ] 5.1 Run `go build ./...` — must succeed
- [ ] 5.2 Run `go vet ./...` — no issues
- [ ] 5.3 Full text search: no `Yuri-NagaSaki` or `soyorins` references remain in the project (excluding .git history and OpenSpec change docs)
- [ ] 5.4 Commit all changes with message: `refactor: migrate from Yuri-NagaSaki to LosFurina namespace`
