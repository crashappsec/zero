# Technology Identification RAG Database

This directory contains pattern definitions for detecting technologies in codebases using a multi-layer detection architecture.

## Overview

- **112 technologies** across 15 categories
- **431 pattern files** with multi-layer detection
- **Confidence scoring** for accurate identification

## Categories

| Category | Technologies | Description |
|----------|-------------|-------------|
| `ai-ml/apis` | OpenAI, Anthropic, Cohere, Google AI, AI21, Mistral, Perplexity, Replicate, Together-AI | AI/ML API providers |
| `ai-ml/vectordb` | Pinecone, Weaviate, Qdrant, ChromaDB | Vector databases for embeddings |
| `ai-ml/mlops` | Hugging Face, Weights & Biases | ML operations platforms |
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
| `web-frameworks` | React, Vue, Angular, Django, Rails, etc. | Web frameworks |
| `cryptographic-libraries` | OpenSSL, etc. | Cryptography |

## Pattern File Structure

Each technology has up to 6 pattern files:

```
technology-name/
├── package-patterns.json    # Package manager detection (npm, pypi, maven, go, rubygems)
├── import-patterns.json     # Import statement patterns by language
├── env-patterns.json        # Environment variable patterns
├── config-patterns.json     # Configuration file patterns
├── api-patterns.json        # API endpoint and usage patterns
└── versions.json            # Version tracking and compatibility
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
./utils/technology-identification/technology-identification-analyser.sh --repo /path/to/repo
```

## Contributing

To add a new technology:

1. Create a directory under the appropriate category
2. Add pattern files following the existing structure
3. Include confidence scores based on pattern specificity
4. Test against real-world repositories

## License

GPL-3.0 - See [LICENSE](../../LICENSE) for details.
