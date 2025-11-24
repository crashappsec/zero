# Anthropic Claude

**Category**: ai-ml/apis
**Description**: Anthropic's Claude AI models - Advanced language models for analysis, content creation, and conversation

## Package Detection

### NPM
- `@anthropic-ai/sdk`

### PYPI
- `anthropic`

### Related Packages
- `@langchain/anthropic`

## Import Detection

### Python
File extensions: .py

**Pattern**: `from anthropic import|import anthropic`
- Anthropic SDK import
- Example: `from anthropic import Anthropic`

**Pattern**: `Anthropic\(|anthropic\.Anthropic`
- Anthropic client initialization
- Example: `client = Anthropic(api_key=api_key)`

### Javascript
File extensions: .js, .ts

**Pattern**: `from ['"]@anthropic-ai/sdk['"]|require\(['"]@anthropic-ai/sdk['"]\)`
- Anthropic SDK import
- Example: `import Anthropic from '@anthropic-ai/sdk';`

### Common Imports
- `anthropic`
- `@anthropic-ai/sdk`

## Environment Variables

*Anthropic API environment variables*

- `ANTHROPIC_API_KEY`
- `ANTHROPIC_MODEL`
- `CLAUDE_API_KEY`

## Detection Notes

- Claude models via Anthropic API
- Supports Claude 3 family (Opus, Sonnet, Haiku)
- Multi-modal capabilities

## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Import Detection**: 90% (HIGH)
- **Environment Variable Detection**: 85% (MEDIUM)
- **API Endpoint Detection**: 70% (MEDIUM)
