# Zero Docker Distribution

Run Zero in a Docker container for consistent, dependency-free execution.

## Quick Start

```bash
# Pull the image
docker pull ghcr.io/crashappsec/zero:latest

# Analyze a repository
docker run -v ~/.zero:/home/zero/.zero \
  -e GITHUB_TOKEN \
  ghcr.io/crashappsec/zero hydrate expressjs/express

# Generate report
docker run -v ~/.zero:/home/zero/.zero \
  ghcr.io/crashappsec/zero report expressjs/express
```

## Setup

### Create an Alias (Recommended)

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
alias zero='docker run -v ~/.zero:/home/zero/.zero -v ~/.gitconfig:/root/.gitconfig:ro -e GITHUB_TOKEN -e ANTHROPIC_API_KEY ghcr.io/crashappsec/zero'
```

Then use Zero naturally:

```bash
zero hydrate expressjs/express
zero report expressjs/express
zero status
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | For private repos | GitHub personal access token |
| `ANTHROPIC_API_KEY` | For agents | Claude API key for AI analysis |

## Commands

### Hydrate (Clone + Scan)

```bash
# Public repository
zero hydrate expressjs/express

# With specific profile
zero hydrate expressjs/express --profile security

# Quick scan
zero hydrate expressjs/express --profile quick
```

### Generate Reports

```bash
# Generate and view report (starts HTTP server, press Ctrl+C to stop)
zero report expressjs/express

# Generate without opening browser
zero report expressjs/express --open=false

# Force regenerate
zero report expressjs/express --regenerate

# Start dev server for live exploration (hot reload)
docker run -v ~/.zero:/home/zero/.zero \
  -p 3000:3000 \
  ghcr.io/crashappsec/zero report expressjs/express --serve
# Then open http://localhost:3000
```

**Note:** Reports require an HTTP server to render (JavaScript loads data). The `zero report` command automatically starts a local server and opens your browser.

### Agent Mode (Interactive)

```bash
# Chat with Zero (requires TTY)
docker run -it \
  -v ~/.zero:/home/zero/.zero \
  -e ANTHROPIC_API_KEY \
  ghcr.io/crashappsec/zero agent
```

### Status & Refresh

```bash
# List hydrated repositories
zero status

# Refresh stale scans
zero refresh

# Force refresh specific repo
zero refresh expressjs/express --force
```

## Data Persistence

All scan data is stored in `~/.zero` on your host machine:

```
~/.zero/
├── repos/
│   └── expressjs/
│       └── express/
│           ├── repo/           # Cloned repository
│           ├── analysis/       # Scanner JSON output
│           └── report/         # Generated HTML report
└── config/                     # Configuration files
```

Mount this directory to persist data between container runs:

```bash
-v ~/.zero:/home/zero/.zero
```

## Building Locally

```bash
# Build the image
docker build -t zero:local .

# Run local build
docker run -v ~/.zero:/home/zero/.zero zero:local hydrate owner/repo
```

## Multi-Architecture Support

The official image supports:
- `linux/amd64` (Intel/AMD)
- `linux/arm64` (Apple Silicon, ARM servers)

## CI/CD Integration

### GitHub Actions

```yaml
name: Security Scan
on: [push]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Zero Scan
        run: |
          docker run \
            -v ${{ github.workspace }}:/repo \
            -v ~/.zero:/home/zero/.zero \
            ghcr.io/crashappsec/zero hydrate . --local /repo

      - name: Generate Report
        run: |
          docker run \
            -v ~/.zero:/home/zero/.zero \
            ghcr.io/crashappsec/zero report .

      - name: Upload Report
        uses: actions/upload-artifact@v4
        with:
          name: security-report
          path: ~/.zero/repos/*/report/
```

### GitLab CI

```yaml
security-scan:
  image: ghcr.io/crashappsec/zero:latest
  script:
    - zero hydrate $CI_PROJECT_PATH
    - zero report $CI_PROJECT_PATH
  artifacts:
    paths:
      - ~/.zero/repos/*/report/
```

## Troubleshooting

### Permission Denied

If you get permission errors, ensure the `.zero` directory is writable:

```bash
mkdir -p ~/.zero
chmod 755 ~/.zero
```

### No GitHub Token

For private repositories, set your GitHub token:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
docker run -e GITHUB_TOKEN ...
```

### Report Server Not Accessible

When using `--serve`, expose port 3000:

```bash
docker run -p 3000:3000 ... report owner/repo --serve
```

## Image Tags

| Tag | Description |
|-----|-------------|
| `latest` | Latest stable release |
| `v3.6.0` | Specific version |
| `sha-abc123` | Specific commit |
| `edge` | Latest development build |
