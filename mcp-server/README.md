# Zero MCP Server

MCP (Model Context Protocol) server for Zero - provides a data layer for repository analysis.

## Overview

This MCP server exposes analysis data from Zero's scanners to Claude Code and other MCP clients. Instead of Claude reading raw JSON files, it can use semantic queries like "get critical vulnerabilities" or "search for crypto behaviors".

## Available Tools

| Tool | Description |
|------|-------------|
| `list_projects` | List all hydrated projects with available analyses |
| `get_project_summary` | Get summary stats for a project (vuln counts, risk levels) |
| `get_malcontent` | Get malware/suspicious behavior findings |
| `get_vulnerabilities` | Get CVE/vulnerability findings |
| `get_technologies` | Get detected technologies and frameworks |
| `get_package_health` | Get dependency health scores |
| `get_licenses` | Get license information |
| `get_code_security` | Get static analysis (Semgrep) findings |
| `search_findings` | Search across all findings for a keyword |
| `get_analysis_raw` | Get raw JSON for any analysis type |

## Installation

```bash
cd mcp-server
npm install
npm run build
```

## Configuration

### For Claude Code

Add to `.claude/mcp.json` in your project:

```json
{
  "mcpServers": {
    "gibson": {
      "command": "node",
      "args": ["/path/to/gibson-powers/mcp-server/dist/index.js"],
      "env": {
        "PHANTOM_HOME": "~/.phantom/projects"
      }
    }
  }
}
```

### For Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "gibson": {
      "command": "node",
      "args": ["/path/to/gibson-powers/mcp-server/dist/index.js"]
    }
  }
}
```

## Usage Examples

Once configured, Claude can use the tools naturally:

**User:** "What projects have we analyzed?"
**Claude:** Uses `list_projects` tool

**User:** "Show me the critical malcontent findings in express"
**Claude:** Uses `get_malcontent` with `project: "expressjs/express"` and `min_risk: "CRITICAL"`

**User:** "Search for crypto patterns across all projects"
**Claude:** Uses `search_findings` with `query: "crypto"`

## Development

```bash
# Run in development mode
npm run dev

# Build
npm run build

# Type check
npm run typecheck
```

## Architecture

```
Claude Code / Claude Desktop
         │
    MCP Protocol (stdio)
         │
    ┌────▼────┐
    │  Zero   │
    │   MCP   │
    │ Server  │
    └────┬────┘
         │
    File System
         │
~/.phantom/projects/
  ├── owner/repo/analysis/
  │   ├── malcontent.json
  │   ├── vulnerabilities.json
  │   ├── technology.json
  │   └── ...
  └── ...
```

## License

GPL-3.0 - Crash Override Inc.
