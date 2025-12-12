# MCP Server Integration

Zero includes an MCP (Model Context Protocol) server that exposes analysis data to AI assistants like Claude Code.

## Overview

The MCP server provides:
- Access to cached analysis data for hydrated repositories
- Real-time scanner execution
- Agent context loading
- Project status information

## Configuration

### Claude Code Integration

Zero's MCP server is configured in `.claude/mcp.json`:

```json
{
  "mcpServers": {
    "zero": {
      "command": "node",
      "args": [
        "./mcp-server/dist/index.js"
      ],
      "env": {
        "ZERO_HOME": "~/.zero"
      }
    }
  }
}
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ZERO_HOME` | `~/.zero` | Base directory for Zero data |
| `GITHUB_TOKEN` | - | GitHub API access (required for hydration) |
| `ANTHROPIC_API_KEY` | - | Claude API key (optional, for AI-enhanced analysis) |

## Available Tools

### Repository Management

#### `zero_hydrate`

Clone and analyze a repository:

```json
{
  "tool": "zero_hydrate",
  "arguments": {
    "repo": "expressjs/express",
    "profile": "security"
  }
}
```

**Parameters:**
- `repo` (required) - GitHub repository in `owner/repo` format
- `profile` (optional) - Scan profile: `quick`, `standard`, `security`, `deep`, `crypto`, `packages`

#### `zero_status`

Get status of hydrated projects:

```json
{
  "tool": "zero_status",
  "arguments": {
    "repo": "expressjs/express"
  }
}
```

**Response:**
```json
{
  "project_id": "expressjs/express",
  "status": "complete",
  "last_scan": "2025-12-12T10:30:00Z",
  "scanners_completed": ["package-sbom", "package-vulns", "code-secrets"],
  "summary": {
    "vulnerabilities": 3,
    "critical": 0,
    "high": 1,
    "secrets": 0
  }
}
```

### Analysis Data

#### `zero_get_findings`

Get analysis findings for a repository:

```json
{
  "tool": "zero_get_findings",
  "arguments": {
    "repo": "expressjs/express",
    "scanner": "package-vulns",
    "severity": "high"
  }
}
```

**Parameters:**
- `repo` (required) - Repository identifier
- `scanner` (optional) - Specific scanner: `package-vulns`, `code-secrets`, `crypto-ciphers`, etc.
- `severity` (optional) - Filter by severity: `critical`, `high`, `medium`, `low`

#### `zero_get_summary`

Get executive summary of all findings:

```json
{
  "tool": "zero_get_summary",
  "arguments": {
    "repo": "expressjs/express"
  }
}
```

### Agent Interaction

#### `zero_invoke_agent`

Invoke a specialist agent:

```json
{
  "tool": "zero_invoke_agent",
  "arguments": {
    "agent": "cereal",
    "repo": "expressjs/express",
    "prompt": "Analyze the CVE findings and prioritize remediation."
  }
}
```

**Parameters:**
- `agent` (required) - Agent name: `cereal`, `razor`, `gill`, `blade`, etc.
- `repo` (required) - Repository to analyze
- `prompt` (required) - Investigation prompt

#### `zero_get_agent_context`

Load context for an agent (for custom analysis):

```json
{
  "tool": "zero_get_agent_context",
  "arguments": {
    "agent": "gill",
    "repo": "expressjs/express",
    "mode": "full"
  }
}
```

**Parameters:**
- `agent` (required) - Agent name
- `repo` (required) - Repository
- `mode` (optional) - Context mode: `summary`, `critical`, `full`

## Data Access

### File Paths

Analysis data is stored at:
```
~/.zero/repos/{owner}/{repo}/
├── repo/                    # Cloned repository
└── analysis/
    ├── manifest.json        # Scan metadata
    ├── package-sbom.json    # SBOM data
    ├── sbom.cdx.json        # CycloneDX SBOM
    ├── package-vulns.json   # Vulnerability findings
    ├── package-health.json  # Package health
    ├── code-secrets.json    # Secret detection
    ├── crypto-ciphers.json  # Cipher analysis
    ├── crypto-keys.json     # Key analysis
    ├── crypto-random.json   # RNG analysis
    ├── crypto-tls.json      # TLS analysis
    └── ...
```

### Direct File Access

The MCP server also exposes tools for direct file access:

```json
{
  "tool": "read_file",
  "arguments": {
    "path": "~/.zero/repos/expressjs/express/analysis/package-vulns.json"
  }
}
```

## Usage Examples

### Complete Security Assessment

```javascript
// 1. Hydrate the repository
await mcpClient.callTool("zero_hydrate", {
  repo: "expressjs/express",
  profile: "security"
});

// 2. Get summary
const summary = await mcpClient.callTool("zero_get_summary", {
  repo: "expressjs/express"
});

// 3. Get critical findings
const critical = await mcpClient.callTool("zero_get_findings", {
  repo: "expressjs/express",
  severity: "critical"
});

// 4. Invoke specialist for investigation
const analysis = await mcpClient.callTool("zero_invoke_agent", {
  agent: "cereal",
  repo: "expressjs/express",
  prompt: "Investigate the critical vulnerabilities. What's the blast radius?"
});
```

### Crypto Analysis Workflow

```javascript
// 1. Run crypto profile
await mcpClient.callTool("zero_hydrate", {
  repo: "myorg/myapp",
  profile: "crypto"
});

// 2. Get Gill's analysis
const cryptoAnalysis = await mcpClient.callTool("zero_invoke_agent", {
  agent: "gill",
  repo: "myorg/myapp",
  prompt: "Review all cryptographic implementations. What needs to be fixed?"
});
```

### Compliance Check

```javascript
// 1. Deep scan for compliance
await mcpClient.callTool("zero_hydrate", {
  repo: "myorg/myapp",
  profile: "deep"
});

// 2. Invoke Blade for compliance assessment
const compliance = await mcpClient.callTool("zero_invoke_agent", {
  agent: "blade",
  repo: "myorg/myapp",
  prompt: "Assess SOC 2 readiness. What control gaps exist?"
});
```

## Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `REPO_NOT_FOUND` | Repository not hydrated | Run `zero_hydrate` first |
| `SCANNER_NOT_FOUND` | Invalid scanner name | Check [Scanner Reference](../scanners/reference.md) |
| `AGENT_NOT_FOUND` | Invalid agent name | Check [Agent Reference](../agents/README.md) |
| `GITHUB_TOKEN_MISSING` | No GitHub token | Set `GITHUB_TOKEN` environment variable |

### Error Response Format

```json
{
  "error": {
    "code": "REPO_NOT_FOUND",
    "message": "Repository expressjs/express has not been hydrated",
    "suggestion": "Run zero_hydrate with repo='expressjs/express' first"
  }
}
```

## See Also

- [Scanner Reference](../scanners/reference.md) - Available scanners
- [Agent Reference](../agents/README.md) - Specialist agents
- [Output Formats](../scanners/output-formats.md) - JSON schemas
