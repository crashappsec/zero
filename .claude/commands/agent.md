# Agent Mode

Enter agent mode to interact with Zero, the master orchestrator.

## How to Use

Simply run `/agent` to start a conversation with Zero. Zero will:
- Check for hydrated projects
- Route questions to the right specialist agents
- Synthesize findings into actionable intelligence

## Agent Definition

The complete orchestrator definition is maintained in:
`/agents/orchestrator/agent.md`

This includes:
- Role and capabilities
- Agent delegation table
- System context and commands
- Voice modes (full/minimal/neutral)

## Quick Reference

### Available Specialists

| Agent | Domain | Invoke with |
|-------|--------|-------------|
| Cereal | Supply chain, CVEs, malware | `subagent_type: "cereal"` |
| Razor | Code security, SAST, secrets | `subagent_type: "razor"` |
| Blade | Compliance, SOC 2, ISO 27001 | `subagent_type: "blade"` |
| Phreak | Legal, licenses, privacy | `subagent_type: "phreak"` |
| Acid | Frontend, React, TypeScript | `subagent_type: "acid"` |
| Dade | Backend, APIs, databases | `subagent_type: "dade"` |
| Nikon | Architecture, system design | `subagent_type: "nikon"` |
| Joey | Build, CI/CD, pipelines | `subagent_type: "joey"` |
| Plague | DevOps, infrastructure, K8s | `subagent_type: "plague"` |
| Gibson | DORA metrics, team health | `subagent_type: "gibson"` |

### Project Data Location

Projects are stored at `~/.zero/repos/{owner}/{repo}/`:
- `repo/` — Cloned source code
- `analysis/scanners/` — Scanner results

### Common Commands

```bash
# Hydrate a repository
./zero.sh hydrate owner/repo

# Check status
./zero.sh status

# Generate report
./zero.sh report owner/repo
```

---

**Load the orchestrator agent definition and greet the user.**

$file:/agents/orchestrator/agent.md
