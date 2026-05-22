# Publishing stemsplit-cli

Package: `stemsplit` CLI binary (GitHub Releases + Homebrew)
GitHub: https://github.com/StemSplit/stemsplit-cli

## Publish workflow

Publishing is fully automated via GitHub Actions using GoReleaser (`release.yml`). Never build or publish binaries manually from a local machine.

### Steps to release a new version

1. **Bump the version** in the code if needed (e.g., `main.go` version constant, if any).

2. **Update `CHANGELOG.md`** or release notes.

3. **Commit and push** to `main`:
   ```bash
   git add .
   git commit -m "chore: bump version to 0.2.0"
   git push origin main
   ```

4. **Create and push a tag** — this triggers the CI release:
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

5. GoReleaser CI will:
   - Build cross-platform binaries (macOS, Linux, Windows)
   - Create a GitHub Release with attached binaries
   - Generate checksums

### After releasing: update Homebrew tap

After a new GitHub release is created, update the formula in `scripts/packages/homebrew-tap/`:

```bash
cd scripts/packages/homebrew-tap
# Get the sha256 of the new tarball:
curl -sL https://github.com/StemSplit/stemsplit-cli/archive/refs/tags/v0.2.0.tar.gz | sha256sum

# Edit Formula/stemsplit.rb to update `url` and `sha256`, then:
git add Formula/stemsplit.rb
git commit -m "chore: update stemsplit to v0.2.0"
git push origin main
```

### Required secrets

| Secret | Description |
|--------|-------------|
| `GITHUB_TOKEN` | Automatically provided by GitHub Actions for GoReleaser |

### Local dev / testing

```bash
cd scripts/packages/stemsplit-cli
go build ./...
go test ./...
./stemsplit --help
```

Build snapshot (do NOT push tags):
```bash
goreleaser build --snapshot --clean
```
