# Technology Identification RAG Database

This directory contains pattern definitions for detecting technologies in codebases using a multi-layer detection architecture.

## Overview

- **119 technologies** across 20+ categories
- **Markdown-based patterns** for easy reading and editing
- **Confidence scoring** for accurate identification

## Categories

| Category | Technologies | Description |
|----------|-------------|-------------|
| `languages` | Python, JavaScript, TypeScript, Go, Rust, Java, C#, Ruby, PHP, Kotlin, Swift, Scala, Elixir, Clojure, Haskell, Nim | Programming languages |
| `ai-ml/apis` | OpenAI, Anthropic, Cohere, Google AI, AI21, Mistral, Perplexity, Replicate, Together-AI | AI/ML API providers |
| `ai-ml/vectordb` | Pinecone, Weaviate, Qdrant, ChromaDB | Vector databases for embeddings |
| `ai-ml/mlops` | Hugging Face, Weights & Biases | ML operations platforms |
| `ai-ml/frameworks` | LangChain | AI/ML frameworks |
| `authentication` | Auth0, Okta, AWS Cognito | Authentication providers |
| `cloud-providers` | AWS, GCP, Azure, Cloudflare | Cloud infrastructure |
| `databases` | PostgreSQL, MongoDB, Redis, MySQL, Elasticsearch, DynamoDB, Supabase, SQLite | Database systems |
| `messaging` | Kafka, RabbitMQ, SQS | Message queues |
| `monitoring` | Datadog, Sentry, New Relic | Observability platforms |
| `business-tools/payment` | Stripe, PayPal, Square, Braintree | Payment processors |
| `business-tools/email` | SendGrid, Mailgun, Resend, Postmark | Email services |
| `business-tools/analytics` | Segment, Mixpanel, Amplitude, PostHog | Analytics platforms |
| `business-tools/cms` | Contentful, Sanity, Strapi, Prismic | Content management |
| `developer-tools/testing` | Jest, Pytest, Cypress, Playwright, Vitest, Mocha | Testing frameworks |
| `developer-tools/cicd` | GitHub Actions, GitLab CI, CircleCI, Jenkins | CI/CD platforms |
| `developer-tools/feature-flags` | LaunchDarkly, Unleash, Flagsmith, GrowthBook | Feature flag services |
| `developer-tools/containers` | Docker, Kubernetes | Container platforms |
| `developer-tools/infrastructure` | Terraform, Pulumi, Ansible, CloudFormation | Infrastructure as Code |
| `developer-tools/linting` | ESLint, Prettier, RuboCop, Pylint, golangci-lint | Code linters/formatters |
| `developer-tools/bundlers` | Webpack, Vite, Rollup, esbuild, Parcel | JavaScript bundlers |
| `web-frameworks/frontend` | React, Vue, Angular, Svelte, Next.js, Nuxt, Astro, Remix | Frontend frameworks |
| `web-frameworks/backend` | Express, FastAPI, Django, Flask, Rails, Spring Boot, Laravel, NestJS, Fastify, Phoenix, ASP.NET Core | Backend frameworks |
| `cryptographic-libraries` | OpenSSL | Cryptography |

## Pattern File Structure

Each technology has a single `patterns.md` file containing all detection patterns:

```
technology-name/
└── patterns.md    # All detection patterns in Markdown format
```

### Pattern File Format

```markdown
# Technology Name

**Category**: category/subcategory
**Description**: Brief description
**Homepage**: https://example.com

## Package Detection

### NPM
- `package-name`
- `@scope/package-name`

### PYPI
- `python-package`

## Configuration Files

- `config-file.json`
- `*.config.js`

## File Extensions

- `.ext`

## Import Detection

### Python
**Pattern**: `import example`
- Description of pattern
- Example: `import example`

## Environment Variables

- `EXAMPLE_API_KEY`
- `EXAMPLE_SECRET`

## Detection Notes

- Important detection considerations
- Common usage patterns

## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Configuration File Detection**: 95% (HIGH)
- **Import Detection**: 90% (HIGH)
- **Environment Variable Detection**: 85% (MEDIUM)
```

## Multi-Layer Detection Architecture

1. **Layer 1a**: SBOM package detection (from Syft/osv-scanner)
2. **Layer 1b**: Manifest file analysis (package.json, requirements.txt, etc.)
3. **Layer 2**: Configuration file patterns
4. **Layer 3**: Import statement analysis (code parsing)
5. **Layer 4**: API endpoint detection
6. **Layer 5**: Environment variable patterns
7. **Layer 6**: Bayesian confidence aggregation across all layers

## Confidence Scoring

Each pattern includes a confidence score (0-100):
- **95+**: High confidence (unique identifiers)
- **85-94**: Strong confidence (specific patterns)
- **70-84**: Moderate confidence (common patterns)
- **<70**: Low confidence (generic patterns)

## Usage

These patterns are used by the `technology-identification-analyser.sh` script:

```bash
# Single repository
./utils/technology-identification/technology-identification-analyser.sh --repo owner/repo

# Organization scan (prints summary after each repo)
./utils/technology-identification/technology-identification-analyser.sh --org org-name

# Local directory
./utils/technology-identification/technology-identification-analyser.sh /path/to/repo
```

## Contributing

To add a new technology:

1. Create a directory under the appropriate category
2. Create a `patterns.md` file following the format above
3. Include confidence scores based on pattern specificity
4. Test against real-world repositories

## License

GPL-3.0 - See [LICENSE](../../LICENSE) for details.
