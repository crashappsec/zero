# Supply Chain Enhancement Implementation Planning Prompt

## Purpose

This prompt guides Claude to generate detailed implementation plans for supply chain security enhancements. Use this when planning new features or modules for the supply chain scanner.

## Prompt Template

```
You are a senior software architect planning the implementation of supply chain security features for the Gibson Powers toolkit. Gibson Powers is an open-source software analysis toolkit that provides deep insights into what software is made of, how it's built, and its security posture.

## Context

Current supply chain scanner capabilities:
- Vulnerability analysis (OSV.dev, CISA KEV integration)
- Provenance analysis (SLSA verification, npm provenance, sigstore)
- Package health analysis (deps.dev, OpenSSF Scorecard, deprecation detection)
- Legal compliance analysis (license checking, secret detection)

Architecture pattern:
- Modular bash scripts with library functions in lib/ directories
- RAG knowledge base in rag/supply-chain/ for Claude-enhanced analysis
- Integration with Anthropic Claude API for intelligent analysis
- Output formats: Markdown, JSON, SARIF
- CLI interface with --flags for configuration

## Feature to Plan: {FEATURE_NAME}

### Description
{FEATURE_DESCRIPTION}

### Requirements
{REQUIREMENTS}

## Planning Output Required

Generate a comprehensive implementation plan with the following sections:

### 1. Architecture Overview
- Where does this feature fit in the existing module structure?
- What new files/directories are needed?
- How does it integrate with existing components?

### 2. Implementation Phases
Break down into phases with:
- Phase name and objective
- Specific deliverables
- Dependencies on other phases
- Estimated complexity (Low/Medium/High)

### 3. File Structure
```
utils/supply-chain/
├── [new directories and files]
└── lib/
    └── [new library files]
```

### 4. Key Functions
For each major function:
- Function name and purpose
- Input parameters
- Output format
- Error handling approach

### 5. External Dependencies
- APIs to integrate (with rate limits and authentication)
- External tools (optional vs required)
- Data sources

### 6. Claude AI Integration
- What RAG documents are needed?
- What analysis prompts are required?
- How should Claude enhance the results?

### 7. CLI Interface
- New flags to add
- Example usage commands
- Integration with existing flags

### 8. Testing Strategy
- Unit test approach
- Integration test scenarios
- Test data requirements

### 9. Success Metrics
- How do we measure the feature works correctly?
- Performance targets
- Accuracy targets

### 10. Rollout Plan
- Phased rollout approach
- Feature flag considerations
- Documentation requirements
```

## Usage

Replace the placeholders:
- `{FEATURE_NAME}`: Name of the feature (e.g., "Library Recommendation Engine")
- `{FEATURE_DESCRIPTION}`: Detailed description of what the feature does
- `{REQUIREMENTS}`: Specific requirements and constraints

## Example Usage

```
Feature to Plan: Library Recommendation Engine

Description:
An AI-powered system that recommends improved library alternatives based on
security record, performance characteristics, maintenance quality, and
community health. Given a list of dependencies, it suggests better alternatives
where available.

Requirements:
- Must support npm, PyPI, and Go ecosystems
- Should provide rationale for each recommendation
- Must consider: security vulnerabilities, maintenance activity, license compatibility
- Should integrate with existing package-health-analyser.sh
- Must cache recommendations to reduce API calls
- Should output in both human-readable and machine-parseable formats
```
