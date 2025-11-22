<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Chalk Build Analyzer

**Status**: 🔬 Experimental

Analyzes build artifacts marked with Chalk to extract and validate supply chain metadata.

## ⚠️ Development Status

This utility is in **early development** and is not yet ready for Beta or production use. It provides basic Chalk metadata analysis but lacks the comprehensive testing, documentation, and features of the Beta supply chain analyzer.

### What Works
- ✅ Chalk metadata extraction
- ✅ Build artifact analysis
- ✅ Supply chain metadata insights
- ✅ AI-enhanced analysis with Claude

### What's Missing
- ❌ Configuration system integration
- ❌ Multi-artifact scanning
- ❌ Output format options (JSON, markdown)
- ❌ Policy enforcement
- ❌ Comprehensive testing
- ❌ Complete documentation

**Use at your own risk**. For Beta-quality supply chain analysis, use the [Supply Chain Security Analyzer](../supply-chain/).

## Overview

[Chalk](https://crashoverride.com/chalk) is a tool by Crash Override for embedding metadata into software artifacts during the build process. This analyzer extracts and validates that metadata to provide insights into:

- **Build Context**: When, where, and how artifacts were built
- **Source Information**: Git commit, branch, repository details
- **Environment**: Build environment configuration and tools
- **Attestations**: Build provenance and signatures
- **Supply Chain**: Complete build supply chain metadata

## Quick Start

### Prerequisites

```bash
# Install Chalk
# See: https://crashoverride.com/chalk

# Verify installation
chalk --version

# Optional: For JSON processing
brew install jq
```

### Basic Usage

```bash
# Analyze chalked artifact
./chalk-build-analyzer.sh /path/to/artifact

# AI-enhanced analysis
export ANTHROPIC_API_KEY="your-key"
./chalk-build-analyzer-claude.sh /path/to/artifact

# Compare base vs Claude analysis
./compare-analyzers.sh /path/to/artifact
```

## Available Scripts

### chalk-build-analyzer.sh

Base analyzer that extracts and displays Chalk metadata.

**Features**:
- Chalk metadata extraction
- Build information display
- Source repository details
- Environment configuration
- Timestamp and versioning

**Usage**:
```bash
# Analyze artifact
./chalk-build-analyzer.sh myapp

# Analyze Docker image
./chalk-build-analyzer.sh my-image:tag

# Analyze binary
./chalk-build-analyzer.sh /usr/local/bin/myapp
```

**Output**:
```
===================================
Chalk Build Analysis
===================================
Artifact: myapp
Analysis Date: 2024-11-21

Build Information:
  Build Time: 2024-11-20 15:30:45 UTC
  Build Host: build-server-01
  Build User: ci-bot

Source Information:
  Repository: github.com/owner/repo
  Commit: abc123...
  Branch: main
  Tag: v1.2.3

Environment:
  OS: Linux 5.15
  Compiler: gcc 11.2
  Dependencies: [list of deps]

Attestations:
  Signed: Yes
  Signature Valid: Yes
  SLSA Level: 2
```

### chalk-build-analyzer-claude.sh

AI-enhanced analyzer with security insights and recommendations.

**Features**:
- All base analyzer features
- Supply chain risk assessment
- Build environment analysis
- Security recommendations
- Compliance checking

**Requires**: `ANTHROPIC_API_KEY` environment variable

**Usage**:
```bash
export ANTHROPIC_API_KEY="your-key"
./chalk-build-analyzer-claude.sh /path/to/artifact
```

### compare-analyzers.sh

Compare base and AI-enhanced analysis side-by-side.

**Usage**:
```bash
./compare-analyzers.sh /path/to/artifact
```

## Chalk Metadata Fields

### Build Context
- `CHALK_ID`: Unique identifier for this chalk instance
- `BUILD_TIME`: When the artifact was built
- `BUILD_HOST`: Where the artifact was built
- `BUILD_USER`: Who built the artifact

### Source Information
- `GIT_REPO`: Source repository URL
- `GIT_COMMIT`: Full commit SHA
- `GIT_BRANCH`: Branch name
- `GIT_TAG`: Tag if present
- `GIT_DIRTY`: Whether working tree was clean

### Environment
- `OS_NAME`: Operating system
- `OS_VERSION`: OS version
- `COMPILER`: Compiler/build tool used
- `COMPILER_VERSION`: Tool version

### Attestations
- `SIGNATURE`: Digital signature
- `ATTESTATION`: Build attestation
- `SLSA_LEVEL`: SLSA compliance level

## Known Limitations

### Current Limitations

1. **Single Artifact Only**: No bulk scanning
2. **No Configuration System**: Cannot persist settings
3. **Limited Output Formats**: Text only
4. **No Policy Enforcement**: Cannot validate against policies
5. **Basic Validation**: Limited metadata validation
6. **No Comparison**: Cannot compare multiple artifacts

### Analysis Limitations

- Requires Chalk to be present in artifacts
- Limited to Chalk-supported metadata
- No historical tracking
- No artifact comparison
- Relies on Chalk tool availability

## Roadmap to Production

### Phase 1: Core Functionality (Current)
- [x] Basic Chalk extraction
- [x] Metadata display
- [x] AI-enhanced analysis
- [ ] Comprehensive error handling

### Phase 2: Integration
- [ ] Hierarchical configuration system
- [ ] Multi-artifact scanning
- [ ] Output format options (JSON, markdown)
- [ ] Policy validation

### Phase 3: Advanced Features
- [ ] Artifact comparison
- [ ] Historical tracking
- [ ] Policy-as-code
- [ ] Dashboard integration
- [ ] Alerting/notifications

### Phase 4: Production Ready
- [ ] Comprehensive testing
- [ ] Complete documentation
- [ ] CI/CD examples
- [ ] Performance optimization
- [ ] Enterprise features

## Development

### Architecture

```
chalk-build-analyzer/
├── chalk-build-analyzer.sh              # Base analyzer
├── chalk-build-analyzer-claude.sh       # AI-enhanced analyzer
└── compare-analyzers.sh                 # Comparison tool
```

### Adding Features

Priority development areas:

1. **Configuration Integration**: Add global config support
2. **Bulk Scanning**: Multiple artifacts, directory scanning
3. **Output Formats**: JSON, markdown, CSV
4. **Policy Validation**: Define and enforce metadata policies
5. **Testing**: Comprehensive test suite
6. **Documentation**: Usage guide and examples

## Use Cases

### Build Verification
Verify artifacts contain expected build metadata.

### Supply Chain Audit
Audit complete supply chain from source to deployment.

### Compliance Checking
Ensure builds meet compliance requirements (SLSA, etc.).

### Incident Response
Quickly identify artifact origins during incidents.

### Release Management
Track and verify release artifacts.

## Examples

### Example 1: Verify Build Origin

```bash
./chalk-build-analyzer.sh myapp | grep "GIT_COMMIT"
```

### Example 2: Check SLSA Level

```bash
./chalk-build-analyzer.sh myapp | grep "SLSA_LEVEL"
```

### Example 3: Audit Multiple Artifacts

```bash
for artifact in app1 app2 app3; do
  echo "Analyzing $artifact..."
  ./chalk-build-analyzer.sh "$artifact"
  echo ""
done
```

### Example 4: CI/CD Validation

```bash
#!/bin/bash
# Verify artifact was built from main branch
./chalk-build-analyzer.sh myapp | grep -q "Branch: main"
if [ $? -ne 0 ]; then
  echo "Artifact not from main branch!"
  exit 1
fi
```

## Related Documentation

- [Chalk Documentation](https://crashoverride.com/chalk)
- [Chalk Build Skill](../../skills/chalk-build-analyzer/)
- [Supply Chain Security](../supply-chain/)
- [Changelog](./CHANGELOG.md)

## Contributing

Contributions welcome! This utility needs significant work to reach production quality. See [CONTRIBUTING.md](../../CONTRIBUTING.md).

Priority areas:
- Configuration system integration
- Multi-artifact support
- Output format options
- Policy validation
- Comprehensive testing

## License

GPL-3.0 - See [LICENSE](../../LICENSE) for details.

## Version

Current version: 1.0.0 (Experimental)

See [CHANGELOG.md](./CHANGELOG.md) for version history.

### Test Organization

Test with the [Gibson Powers Test Organization](https://github.com/Gibson-Powers-Test-Org):

```bash
# Basic analysis
./$(basename $readme) [input]

# Claude AI analysis (when fully implemented)
./$(basename $readme) --claude [input]

# Get all options
./$(basename $readme) --help
```

### Claude AI Support

This tool supports `--claude` flag for AI-powered analysis (implementation in progress).
Set `ANTHROPIC_API_KEY` environment variable or use `-k` flag.
