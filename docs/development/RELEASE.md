# Release Process

> Release process and versioning for Zero.

## Versioning

Zero follows [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH[-PRERELEASE]

Examples:
  1.0.0        - First stable release
  1.1.0        - New features, backward compatible
  1.1.1        - Bug fixes only
  1.0.0-alpha  - Pre-release: unstable, breaking changes expected
  1.0.0-beta   - Pre-release: feature complete, testing phase
  1.0.0-rc.1   - Release candidate
```

### Version Bumping

| Change Type | Version Bump | Example |
|-------------|--------------|---------|
| Breaking API change | MAJOR | 1.0.0 → 2.0.0 |
| New feature (backward compatible) | MINOR | 1.0.0 → 1.1.0 |
| Bug fix | PATCH | 1.0.0 → 1.0.1 |
| Pre-release | PRERELEASE | 1.0.0-alpha → 1.0.0-beta |

### Current Status

```
Current: 0.1.0-experimental
Next:    1.0.0-alpha
```

---

## Release Checklist

### Pre-Release

- [ ] All CI checks passing on main
- [ ] No critical or high-severity bugs open
- [ ] CHANGELOG.md updated with release notes
- [ ] Version number updated in code
- [ ] Documentation up to date
- [ ] Test coverage meets targets

### Version Update Locations

```bash
# Files to update:
cmd/zero/main.go           # var version = "x.y.z"
pkg/api/server.go          # Version in health response
docs/development/ROADMAP.md # Version header
web/package.json           # version field
```

### Release Steps

1. **Create release branch**
   ```bash
   git checkout main
   git pull origin main
   git checkout -b release/v1.0.0
   ```

2. **Update version**
   ```bash
   # Update version in all locations
   # Update CHANGELOG.md
   ```

3. **Create PR for release**
   ```bash
   git add -A
   git commit -m "chore: Prepare release v1.0.0"
   git push -u origin release/v1.0.0
   gh pr create --title "Release v1.0.0" --body "Release notes..."
   ```

4. **Merge and tag**
   ```bash
   # After PR approval and merge
   git checkout main
   git pull origin main
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

5. **Create GitHub Release**
   ```bash
   gh release create v1.0.0 \
     --title "Zero v1.0.0" \
     --notes-file RELEASE_NOTES.md \
     zero-linux-amd64 \
     zero-darwin-amd64 \
     zero-darwin-arm64
   ```

---

## Changelog Format

Follow [Keep a Changelog](https://keepachangelog.com/):

```markdown
# Changelog

## [Unreleased]

### Added
- New feature X

### Changed
- Modified behavior Y

### Fixed
- Bug fix Z

### Removed
- Deprecated feature W

## [1.0.0] - 2026-01-15

### Added
- Initial stable release
- 7 super scanners
- 12 specialist agents
- Web UI dashboard
- Report generation

### Security
- Fixed CORS vulnerability in WebSocket handler
```

---

## Release Artifacts

### Binary Builds

```yaml
# Built by GitHub Actions on tag push
Platforms:
  - linux/amd64
  - darwin/amd64 (Intel Mac)
  - darwin/arm64 (Apple Silicon)
  - windows/amd64 (future)
```

### Docker Images

```bash
# Built and pushed to GitHub Container Registry
ghcr.io/crashappsec/zero:latest
ghcr.io/crashappsec/zero:1.0.0
ghcr.io/crashappsec/zero:1.0
ghcr.io/crashappsec/zero:1
```

### Checksums

```bash
# SHA256 checksums for all artifacts
zero-linux-amd64.sha256
zero-darwin-amd64.sha256
zero-darwin-arm64.sha256
```

---

## GitHub Actions Release Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin]
        goarch: [amd64, arm64]
        exclude:
          - goos: linux
            goarch: arm64
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          go build -ldflags="-s -w" -o zero-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/zero
          sha256sum zero-${{ matrix.goos }}-${{ matrix.goarch }} > zero-${{ matrix.goos }}-${{ matrix.goarch }}.sha256
      - uses: actions/upload-artifact@v4
        with:
          name: zero-${{ matrix.goos }}-${{ matrix.goarch }}
          path: |
            zero-${{ matrix.goos }}-${{ matrix.goarch }}
            zero-${{ matrix.goos }}-${{ matrix.goarch }}.sha256

  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
      - uses: softprops/action-gh-release@v1
        with:
          files: |
            **/zero-*
          generate_release_notes: true

  docker:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          push: true
          tags: |
            ghcr.io/crashappsec/zero:${{ github.ref_name }}
            ghcr.io/crashappsec/zero:latest
```

---

## Hotfix Process

For critical fixes to released versions:

1. **Create hotfix branch from tag**
   ```bash
   git checkout -b hotfix/v1.0.1 v1.0.0
   ```

2. **Apply fix**
   ```bash
   # Make fix
   git commit -m "fix: Critical security issue"
   ```

3. **Release hotfix**
   ```bash
   git tag -a v1.0.1 -m "Hotfix v1.0.1"
   git push origin v1.0.1
   ```

4. **Merge back to main**
   ```bash
   git checkout main
   git merge hotfix/v1.0.1
   git push origin main
   ```

---

## Release Communication

### Announcement Template

```markdown
# Zero v1.0.0 Released! 🚀

We're excited to announce Zero v1.0.0, our first stable release!

## Highlights

- **7 Super Scanners** - Comprehensive security analysis
- **12 AI Agents** - Specialist analysis powered by Claude
- **Web Dashboard** - Interactive results viewer
- **Docker Support** - Run anywhere

## Installation

```bash
# Download binary
curl -L https://github.com/crashappsec/zero/releases/download/v1.0.0/zero-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o zero
chmod +x zero

# Or use Docker
docker pull ghcr.io/crashappsec/zero:1.0.0
```

## Documentation

- [Getting Started](docs/GETTING_STARTED.md)
- [Full Changelog](CHANGELOG.md)

## Thanks

Thanks to all contributors who made this release possible!
```

### Channels

- GitHub Releases (automatic)
- Project README badge update
- Documentation site (if applicable)
