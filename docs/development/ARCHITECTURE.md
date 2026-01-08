# Architecture Overview

> System design and component overview for Zero.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              User Interface                              │
├──────────────────┬──────────────────┬───────────────────────────────────┤
│    CLI (zero)    │    Web UI        │    MCP Server (IDE)               │
│    cmd/zero/     │    web/          │    pkg/mcp/                       │
└────────┬─────────┴────────┬─────────┴───────────────┬───────────────────┘
         │                  │                         │
         ▼                  ▼                         ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            API Server                                    │
│                          pkg/api/                                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │  REST API   │  │  WebSocket  │  │    SSE      │  │   Agent     │    │
│  │  handlers   │  │    Hub      │  │  Streaming  │  │  Handler    │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
└────────┬────────────────┬────────────────┬────────────────┬─────────────┘
         │                │                │                │
         ▼                ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           Core Services                                  │
├─────────────────┬─────────────────┬─────────────────┬───────────────────┤
│    Scanners     │     Agents      │    Workflow     │     Storage       │
│  pkg/scanner/   │   pkg/agent/    │  pkg/workflow/  │   pkg/storage/    │
└─────────────────┴─────────────────┴─────────────────┴───────────────────┘
         │                │                │                │
         ▼                ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Foundation Layer                                │
│                            pkg/core/                                     │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │ Config  │ │ Findings│ │   RAG   │ │  Rules  │ │ GitHub  │           │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘           │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Component Details

### CLI (`cmd/zero/`)

Entry point for command-line interface.

```
cmd/zero/
├── main.go           # Entry point, cobra root command
├── commands/
│   ├── hydrate.go    # Clone and scan repos
│   ├── status.go     # Show project status
│   ├── report.go     # Generate reports
│   ├── agent.go      # Interactive agent mode
│   ├── serve.go      # Start API server
│   └── feeds.go      # Sync security feeds
```

**Key Commands:**
- `zero hydrate <repo>` - Clone and analyze repository
- `zero agent` - Start interactive agent chat
- `zero serve` - Start API server
- `zero report <repo>` - Generate HTML report

### API Server (`pkg/api/`)

REST API and real-time communication layer.

```
pkg/api/
├── server.go         # HTTP server setup, middleware
├── routes.go         # Route definitions
├── handlers/         # REST endpoint handlers
│   ├── repos.go      # Repository CRUD
│   ├── scans.go      # Scan management
│   ├── analysis.go   # Analysis results
│   └── settings.go   # Configuration
├── agent/            # Agent chat handler
│   └── handler.go    # SSE streaming for chat
└── ws/               # WebSocket hub
    └── hub.go        # Real-time updates
```

**Endpoints:**
- `GET /api/repos` - List analyzed repositories
- `POST /api/scans` - Start new scan
- `POST /api/chat/stream` - SSE chat with agent
- `WS /ws/scan/:id` - Real-time scan progress

### Scanners (`pkg/scanner/`)

Security analysis engines.

```
pkg/scanner/
├── interface.go              # Scanner interface definition
├── runner.go                 # Concurrent scanner execution
├── code-packages/            # SBOM + dependency analysis
│   ├── code-packages.go      # Main scanner
│   ├── sbom.go               # SBOM generation
│   ├── vulns.go              # Vulnerability scanning
│   └── malcontent.go         # Supply chain detection
├── code-security/            # SAST + secrets
│   ├── security.go           # Main scanner
│   ├── semgrep.go            # Semgrep integration
│   ├── secrets.go            # Secret detection
│   └── crypto.go             # Cryptography analysis
├── code-quality/             # Quality metrics
├── devops/                   # DevOps analysis
├── technology-identification/# Tech detection
├── code-ownership/           # Contributor analysis
└── developer-experience/     # DX metrics
```

**Scanner Interface:**
```go
type Scanner interface {
    Name() string
    Description() string
    Features() []string
    Run(ctx context.Context, opts *ScanOptions) (*ScanResult, error)
}
```

### Agents (`pkg/agent/`)

AI-powered analysis specialists.

```
pkg/agent/
├── runtime.go        # Agent execution runtime
├── prompt.go         # Prompt building
├── tools.go          # Tool definitions
├── system_info.go    # Self-awareness tools
└── delegation.go     # Agent-to-agent delegation

agents/                # Agent definitions (markdown)
├── orchestrator/     # Zero - master orchestrator
├── supply-chain/     # Cereal - supply chain
├── security/         # Razor - code security
├── crypto/           # Gill - cryptography
├── ai-security/      # Hal - AI/ML security
└── ...               # 12 specialist agents
```

**Agent Architecture:**
```
User Query
    │
    ▼
┌─────────────────┐
│  Zero (Router)  │ ← Orchestrator agent
└────────┬────────┘
         │ Delegates based on query type
         ▼
┌────────────────────────────────────────┐
│  Specialist Agents                      │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐  │
│  │ Cereal  │ │ Razor   │ │ Gill    │  │
│  │ Supply  │ │ Security│ │ Crypto  │  │
│  └─────────┘ └─────────┘ └─────────┘  │
└────────────────────────────────────────┘
         │
         ▼ Uses tools
┌────────────────────────────────────────┐
│  Tools: GetAnalysis, Read, Grep,       │
│         WebSearch, DelegateAgent       │
└────────────────────────────────────────┘
```

### Workflow (`pkg/workflow/`)

High-level operations.

```
pkg/workflow/
├── hydrate/          # Clone + scan workflow
│   └── hydrate.go    # Repository hydration
├── automation/       # Watch mode, auto-refresh
│   └── watcher.go    # File system watcher
├── freshness/        # Staleness tracking
│   └── freshness.go  # Age calculations
└── diff/             # Scan comparison
    └── delta.go      # Diff generation
```

### Core (`pkg/core/`)

Foundation utilities.

```
pkg/core/
├── config/           # Configuration loading
├── findings/         # Finding types and severity
├── sarif/            # SARIF export format
├── rag/              # RAG pattern system
│   ├── loader.go     # Pattern loading
│   └── validator.go  # Pattern validation
├── rules/            # Semgrep rule management
├── github/           # GitHub API client
├── liveapi/          # Live API queries (OSV)
└── feeds/            # Security feed sync
```

### Web UI (`web/`)

Next.js dashboard.

```
web/
├── src/
│   ├── app/              # Next.js app router
│   │   ├── page.tsx      # Dashboard
│   │   ├── repos/        # Repository views
│   │   ├── chat/         # Agent chat
│   │   └── settings/     # Configuration
│   ├── components/       # React components
│   │   ├── ui/           # Base components
│   │   └── layout/       # Layout components
│   ├── hooks/            # Custom React hooks
│   │   └── useApi.ts     # API integration
│   └── lib/              # Utilities
│       ├── api.ts        # API client
│       └── types.ts      # TypeScript types
└── next.config.js        # Next.js config
```

---

## Data Flow

### Scan Flow

```
1. User: zero hydrate owner/repo
         │
         ▼
2. Hydrate Workflow
   ├── Clone repository to .zero/repos/<project>/repo/
   │
   ├── Load scan profile (scanners + features)
   │
   └── Execute scanners concurrently
         │
         ▼
3. Scanner Execution
   ├── code-packages: SBOM → vulns → malcontent → ...
   ├── code-security: semgrep → secrets → crypto → ...
   ├── devops: iac → containers → actions → ...
   └── ... (7 scanners)
         │
         ▼
4. Results Storage
   └── .zero/repos/<project>/analysis/
       ├── code-packages.json
       ├── code-security.json
       └── ...
         │
         ▼
5. Freshness Tracking
   └── .zero/repos/<project>/freshness.json
```

### Agent Chat Flow

```
1. User Message
         │
         ▼
2. API Server (POST /api/chat/stream)
         │
         ▼
3. Agent Runtime
   ├── Load agent definition
   ├── Build system prompt
   ├── Include project context (if selected)
   └── Call Claude API
         │
         ▼
4. Claude Response (streaming)
   ├── Text chunks → SSE to client
   └── Tool calls → Execute tools
         │
         ▼
5. Tool Execution
   ├── GetAnalysis → Read cached JSON
   ├── Read/Grep → File operations
   ├── WebSearch → External search
   └── DelegateAgent → Call specialist
         │
         ▼
6. Continue conversation until done
```

---

## Storage Layout

```
.zero/                          # Zero home directory
├── config/                     # Configuration
│   └── zero.config.json        # Scanner config
├── cache/                      # Cached data
│   ├── semgrep-rules/          # Synced Semgrep rules
│   └── rag-rules/              # Generated RAG rules
└── repos/                      # Analyzed repositories
    └── <owner>/<repo>/
        ├── repo/               # Cloned repository
        ├── analysis/           # Scanner results
        │   ├── code-packages.json
        │   ├── code-security.json
        │   ├── sbom.cdx.json
        │   └── ...
        ├── freshness.json      # Scan metadata
        └── report/             # Generated reports
```

---

## External Dependencies

### Required Tools
- **Semgrep** - SAST scanning
- **cdxgen** - SBOM generation
- **Trivy** - Container/IaC scanning
- **Checkov** - IaC security

### APIs
- **Claude API** - AI agent capabilities
- **GitHub API** - Repository access
- **OSV.dev** - Vulnerability data (live queries)
- **deps.dev** - Package health scores

---

## Security Considerations

### API Security
- CORS restricted to configured origins
- No authentication required (local tool)
- WebSocket origin validation needed

### Data Security
- Credentials stored in `.env` (gitignored)
- No sensitive data in scan results
- API keys never logged

### Scanning Security
- Sandboxed tool execution
- No code execution from scanned repos
- Read-only repository access
