# Bundle Analysis Scanner

Analyzes npm/JavaScript bundle sizes, identifies heavy packages, and detects tree-shaking compatibility.

## Features

- **Bundle Size Analysis**: Queries Bundlephobia API for accurate minified and gzipped sizes
- **Heavy Package Detection**: Flags packages over configurable threshold (default: 50KB gzipped)
- **Tree-Shaking Detection**: Identifies ESM support via npm registry metadata
- **Actionable Recommendations**: Suggests optimizations based on detected issues

## Usage

```bash
# Analyze local project
./bundle-analysis.sh /path/to/project

# Analyze with output file
./bundle-analysis.sh -o bundle-analysis.json /path/to/project

# Use cached repository from Zero
./bundle-analysis.sh --repo expressjs/express

# Custom heavy threshold (100KB)
./bundle-analysis.sh --threshold 100000 /path/to/project

# Enable API response caching
./bundle-analysis.sh --cache-dir /tmp/bundle-cache /path/to/project
```

## Options

| Option | Description |
|--------|-------------|
| `--local-path PATH` | Use pre-cloned repository |
| `--repo OWNER/REPO` | GitHub repository (from Zero cache) |
| `--org ORG` | GitHub org (first repo in Zero cache) |
| `-o, --output FILE` | Write JSON to file (default: stdout) |
| `--threshold BYTES` | Heavy package threshold in gzipped bytes (default: 50000) |
| `--cache-dir DIR` | Directory for API response caching |
| `-k, --keep-clone` | Keep cloned repository |
| `-h, --help` | Show help |

## Output Format

```json
{
  "analyzer": "bundle-analysis",
  "version": "1.0.0",
  "timestamp": "2025-12-08T12:00:00Z",
  "target": "/path/to/project",
  "summary": {
    "total_dependencies": 45,
    "analyzed": 42,
    "failed": 3,
    "total_size_bytes": 5242880,
    "total_gzip_bytes": 1048576,
    "heavy_packages": 5,
    "tree_shakeable": 38,
    "not_tree_shakeable": 4
  },
  "packages": [...],
  "top_largest": [...],
  "tree_shaking_issues": [...],
  "recommendations": [...]
}
```

## APIs Used

### Bundlephobia API
- Endpoint: `https://bundlephobia.com/api/size?package=<name>@<version>`
- Returns: `{ name, version, size, gzip, dependencyCount }`
- Rate limited: 100ms delay between requests

### npm Registry API
- Endpoint: `https://registry.npmjs.org/<package>`
- Used for: ESM detection (`module`, `exports`, `type` fields)
- Used for: Side effects detection (`sideEffects` field)

## Package Classification

| Rating | Gzipped Size |
|--------|--------------|
| Minimal | < 5 KB |
| Small | 5-25 KB |
| Medium | 25-50 KB |
| Large | 50-100 KB |
| Very Large | > 100 KB |

Packages rated "Large" or "Very Large" are flagged as heavy.

## Tree-Shaking Detection

ESM support is detected by checking npm registry metadata:
- `module` field - ESM entry point
- `exports` field with `import` condition
- `type: "module"` - Full ESM package

Tree-shaking effectiveness is also influenced by `sideEffects`:
- `sideEffects: false` - Fully tree-shakeable
- `sideEffects: true` - Limited tree-shaking
- Missing - Bundlers may not tree-shake safely

## Recommendations Generated

1. **Heavy Package**: Package over threshold without tree-shaking
2. **Tree Shaking**: CommonJS-only package that can't be tree-shaken

## Library Files

- `lib/bundlephobia-client.sh` - Bundlephobia API client with caching
- `lib/npm-registry-client.sh` - npm registry API client
- `lib/tree-shake-detector.sh` - Tree-shaking analysis logic

## Dependencies

- `bash` 4.0+
- `curl` - HTTP requests
- `jq` - JSON processing

## License

GPL-3.0 - Crash Override Inc.
