# LLM API Provider Detection Patterns

Patterns for detecting Large Language Model API provider usage in codebases.

## Providers

### OpenAI

- **Name**: OpenAI
- **Environment Variables**: `OPENAI_API_KEY`
- **Packages**: `openai`
- **API Key Pattern**: `sk-[a-zA-Z0-9]{20,}`
- **Model Pattern**: `model\s*[=:]\s*["'](gpt-4[^"']*|gpt-3\.5[^"']*|text-embedding[^"']*|dall-e[^"']*|whisper[^"']*|o1[^"']*)["']`

### Anthropic

- **Name**: Anthropic
- **Environment Variables**: `ANTHROPIC_API_KEY`
- **Packages**: `anthropic`
- **API Key Pattern**: `sk-ant-[a-zA-Z0-9-]{20,}`
- **Model Pattern**: `model\s*[=:]\s*["'](claude-[^"']+)["']`

### Google AI

- **Name**: Google AI
- **Environment Variables**: `GOOGLE_API_KEY`, `GEMINI_API_KEY`
- **Packages**: `google-generativeai`, `vertexai`
- **API Key Pattern**: `AIza[a-zA-Z0-9_-]{35}`
- **Model Pattern**: `model\s*[=:]\s*["'](gemini-[^"']+)["']`

### Cohere

- **Name**: Cohere
- **Environment Variables**: `COHERE_API_KEY`
- **Packages**: `cohere`
- **API Key Pattern**: `[a-zA-Z0-9]{40}`
- **Model Pattern**: `model\s*[=:]\s*["'](command[^"']*|embed[^"']*)["']`

### Mistral

- **Name**: Mistral
- **Environment Variables**: `MISTRAL_API_KEY`
- **Packages**: `mistralai`
- **API Key Pattern**: `[a-zA-Z0-9]{32}`
- **Model Pattern**: `model\s*[=:]\s*["'](mistral-[^"']+|open-mistral[^"']+|open-mixtral[^"']+)["']`

### Replicate

- **Name**: Replicate
- **Environment Variables**: `REPLICATE_API_TOKEN`
- **Packages**: `replicate`
- **API Key Pattern**: `r8_[a-zA-Z0-9]{40}`

### Together AI

- **Name**: Together AI
- **Environment Variables**: `TOGETHER_API_KEY`
- **Packages**: `together`
- **API Key Pattern**: `[a-f0-9]{64}`
- **Model Pattern**: `together\.Complete[^(]*\([^)]*model\s*=\s*["']([^"']+)["']`

### Groq

- **Name**: Groq
- **Environment Variables**: `GROQ_API_KEY`
- **Packages**: `groq`
- **API Key Pattern**: `gsk_[a-zA-Z0-9]{52}`
- **Model Pattern**: `Groq\s*\([^)]*\)\.chat[^(]*\([^)]*model\s*=\s*["']([^"']+)["']`

### Perplexity

- **Name**: Perplexity
- **Environment Variables**: `PERPLEXITY_API_KEY`
- **Packages**: `openai` (uses OpenAI SDK)
- **API Key Pattern**: `pplx-[a-zA-Z0-9]{48}`
- **Model Pattern**: `model\s*[=:]\s*["'](llama-[^"']*-sonar[^"']*|pplx-[^"']+)["']`

### Fireworks AI

- **Name**: Fireworks AI
- **Environment Variables**: `FIREWORKS_API_KEY`
- **Packages**: `fireworks-ai`
- **API Key Pattern**: `fw_[a-zA-Z0-9]{43}`
- **Model Pattern**: `model\s*[=:]\s*["'](accounts/fireworks/models/[^"']+)["']`

### Anyscale

- **Name**: Anyscale
- **Environment Variables**: `ANYSCALE_API_KEY`
- **Packages**: `openai` (uses OpenAI SDK)
- **API Key Pattern**: `esecret_[a-zA-Z0-9]{32}`
- **Base URL Pattern**: `base_url\s*=\s*["']https://api\.endpoints\.anyscale\.com`

### DeepInfra

- **Name**: DeepInfra
- **Environment Variables**: `DEEPINFRA_API_TOKEN`
- **Packages**: `openai` (uses OpenAI SDK)
- **API Key Pattern**: `[a-zA-Z0-9]{43}`
- **Base URL Pattern**: `base_url\s*=\s*["']https://api\.deepinfra\.com`

## LangChain Integration Patterns

Pattern for detecting models via LangChain wrappers:

```regex
Chat(?:OpenAI|Anthropic|Google|VertexAI|Cohere|Mistral|Groq|Fireworks|Together|Perplexity)\s*\([^)]*model(?:_name)?\s*=\s*["']([^"']+)["']
```

This pattern captures the model name from LangChain chat model constructors.

## Security Notes

1. **Never commit API keys** - Use environment variables or secrets management
2. **Rotate exposed keys immediately** - If a key is found in code, rotate it
3. **Use least-privilege keys** - Prefer read-only or limited scope API keys
4. **Monitor usage** - Enable usage alerts with API providers
