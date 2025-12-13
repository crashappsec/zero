# AI Technology Categories

AI technology category definitions for AI Adoption reporting and governance.

## AI Categories

### LLM APIs (`ai-ml/apis`)

Large Language Model API providers.

- **Risk Level**: Medium
- **Examples**: OpenAI, Anthropic, Google AI, Cohere, Mistral, AI21, Replicate

### AI Frameworks (`ai-ml/frameworks`)

AI/ML application frameworks and orchestration.

- **Risk Level**: Low
- **Examples**: LangChain, LlamaIndex, Haystack

### Vector Databases (`ai-ml/vectordb`)

Vector storage for embeddings and RAG applications.

- **Risk Level**: Low
- **Examples**: Pinecone, Weaviate, Qdrant, ChromaDB, Milvus, pgvector, FAISS

### MLOps Tools (`ai-ml/mlops`)

Machine learning operations and model management.

- **Risk Level**: Low
- **Examples**: Hugging Face, Weights & Biases, MLflow, DVC

### AI Agents (`ai-ml/agents`)

Autonomous AI agent frameworks with tool use capabilities.

- **Risk Level**: High
- **Examples**: LangChain Agents, CrewAI, AutoGPT, Microsoft AutoGen, Claude Agent SDK
- **Security Notes**: Agents can execute code, access files, and make network requests autonomously

### Tool/Function Calling (`ai-ml/patterns/tool-calling`)

LLM tool use and function calling implementations.

- **Risk Level**: Medium
- **Examples**: OpenAI Functions, Claude Tools, Gemini Functions
- **Security Notes**: Tool inputs should be validated; avoid eval/exec with tool arguments

### RAG Implementations (`ai-ml/patterns/rag`)

Retrieval Augmented Generation architectures.

- **Risk Level**: Low
- **Examples**: LangChain RAG, LlamaIndex, Custom RAG pipelines
- **Security Notes**: Document sources should be validated; query injection risks

### Agent Orchestration (`ai-ml/patterns/orchestration`)

Multi-agent coordination and workflow systems.

- **Risk Level**: High
- **Examples**: LangGraph, CrewAI Crews, AutoGen GroupChat
- **Security Notes**: Multi-agent systems multiply attack surfaces; require strict resource limits

### AI Coding Assistants (`genai-tools`)

AI-powered code generation and assistance tools.

- **Risk Level**: Medium
- **Examples**: GitHub Copilot, Cursor, Codeium, Tabnine, Amazon CodeWhisperer

## Usage Patterns

### Simple API Call

Direct LLM API calls without tools or agents.

- **Maturity**: Experimental
- **Indicators**: Single API call, No tool definitions, Basic error handling

### Tool/Function Calling

LLM with defined tools for external actions.

- **Maturity**: Emerging
- **Indicators**: `tools=` parameter, function definitions, Tool response handling

### Basic RAG

Simple retrieval + generation pipeline.

- **Maturity**: Emerging
- **Indicators**: Vector store queries, Context assembly, Single retrieval step

### Advanced RAG

Sophisticated RAG with reranking, hybrid search.

- **Maturity**: Standardized
- **Indicators**: Multiple retrievers, Reranking, Query transformation

### Single Agent

Autonomous agent with tool loop.

- **Maturity**: Standardized
- **Indicators**: AgentExecutor, Agentic loop, Tool chain

### Multi-Agent System

Multiple coordinated agents.

- **Maturity**: Optimized
- **Indicators**: Agent orchestration, Message passing, Role specialization

### Autonomous System

Self-directed goal pursuit with minimal human oversight.

- **Maturity**: Strategic
- **Indicators**: Continuous execution, Self-improvement, Goal decomposition

## Maturity Levels

| Level | Name | Description | Indicators |
|-------|------|-------------|------------|
| 1 | Experimental | Individual developers testing AI APIs | Single-file AI usage, No error handling, Hardcoded keys |
| 2 | Emerging | Team-level AI adoption | Multiple repos with AI, Environment variables, Basic retry logic |
| 3 | Standardized | Organization-wide AI standards | Shared AI libraries, Proxy/gateway usage, Consistent error handling |
| 4 | Optimized | AI Center of Excellence | Custom abstractions, Caching layers, Usage analytics |
| 5 | Strategic | AI-native architecture | Multi-model routing, Agentic workflows, RAG infrastructure |

## Governance

### Categories Requiring Approval

- `ai-ml/apis`
- `ai-ml/agents`
- `ai-ml/patterns/orchestration`

### Categories Requiring Security Review

- `ai-ml/apis`
- `ai-ml/agents`
- `ai-ml/patterns/tool-calling`
- `genai-tools`

### Data Classification Notes

| Category | Classification |
|----------|---------------|
| `ai-ml/apis` | May process sensitive data via API calls |
| `ai-ml/agents` | Autonomous execution with external access capabilities |
| `ai-ml/patterns/rag` | Document embeddings may contain sensitive content |
| `genai-tools` | May expose code context to external services |

## Risk Mitigation

### AI Agents

1. Implement least-privilege tool access
2. Add rate limits and circuit breakers
3. Log all agent actions for audit
4. Implement kill switches for autonomous agents

### Tool/Function Calling

1. Validate all tool inputs
2. Avoid eval/exec with tool arguments
3. Implement rate limiting

### Agent Orchestration

1. Set iteration/time limits
2. Authenticate agent-to-agent communication
3. Monitor resource consumption
