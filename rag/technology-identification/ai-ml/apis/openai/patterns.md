# OpenAI

**Category**: ai-ml/apis
**Description**: OpenAI API client library for GPT-4, ChatGPT, DALL-E, and other AI models
**Homepage**: https://openai.com

## Package Detection

### NPM
*OpenAI Node.js client*

- `openai`

### PYPI
*OpenAI Python client*

- `openai`

### Related Packages
- `@langchain/openai`
- `openai-edge`
- `gpt-tokenizer`
- `tiktoken`

## Import Detection

### Javascript
File extensions: .js, .ts

**Pattern**: `import.*from ['"]openai['"]`
- OpenAI client import
- Example: `import OpenAI from 'openai';`

**Pattern**: `new OpenAI\(`
- OpenAI client instantiation
- Example: `const openai = new OpenAI({ apiKey: process.env.OPENAI_API_KEY });`

### Python
File extensions: .py

**Pattern**: `^import openai`
- OpenAI Python import
- Example: `import openai`

**Pattern**: `^from openai import`
- OpenAI specific imports
- Example: `from openai import OpenAI`

### Common Imports
- `openai`
- `openai.OpenAI`
- `openai.ChatCompletion`

## Environment Variables

*OpenAI API configuration*

- `OPENAI_API_KEY`
- `OPENAI_ORG_ID`
- `OPENAI_MODEL`

## Detection Notes

- Check for OpenAI API key in environment (OPENAI_API_KEY)
- Look for GPT model names in code (gpt-4, gpt-3.5-turbo)
- Common with LangChain and AI frameworks

## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Import Detection**: 90% (HIGH)
- **Environment Variable Detection**: 85% (MEDIUM)
- **API Endpoint Detection**: 80% (MEDIUM)
