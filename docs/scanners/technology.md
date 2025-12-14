# Technology Scanner

The Technology scanner provides comprehensive technology identification and AI/ML analysis, generating both a technology inventory and an **ML-BOM (Machine Learning Bill of Materials)**. It detects languages, frameworks, AI models, datasets, and infrastructure components using RAG-based patterns.

## Overview

| Property | Value |
|----------|-------|
| **Name** | `technology` |
| **Version** | 1.0.0 |
| **Output File** | `technology.json` |
| **Dependencies** | None (optionally uses SBOM for enrichment) |
| **Estimated Time** | 30-90 seconds |

## Features

### 1. Detection (`detection`)

Identifies technologies, frameworks, and libraries using tiered RAG-based detection.

**Configuration:**
```json
{
  "detection": {
    "enabled": true,
    "tier": "auto",
    "scan_extensions": true,
    "scan_config": true,
    "scan_sbom": true,
    "scan_imports": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable technology detection |
| `tier` | string | `"auto"` | Detection tier: `"quick"`, `"deep"`, or `"auto"` |
| `scan_extensions` | bool | `true` | Detect via file extensions |
| `scan_config` | bool | `true` | Detect via config files |
| `scan_sbom` | bool | `true` | Enrich from SBOM data |
| `scan_imports` | bool | `true` | Scan code import statements |

**Detection Tiers:**

| Tier | Speed | Method | Use Case |
|------|-------|--------|----------|
| Tier 1 (Quick) | ~5-10s | SBOM package analysis | CI/CD pipelines, bulk scans |
| Tier 2 (Deep) | ~30-60s | File scanning, imports, patterns | Detailed inventory |
| Tier 3 (Extract) | With Tier 2 | Config value extraction | Infrastructure mapping |

**RAG Pattern Categories (119+ Technologies):**

| Category | Count | Examples |
|----------|-------|----------|
| Languages | 16 | Python, JavaScript, Go, Rust, Java, TypeScript |
| Web Frameworks - Frontend | 8 | React, Vue, Angular, Svelte, Next.js, Nuxt |
| Web Frameworks - Backend | 11 | Express, FastAPI, Django, Flask, Rails, Spring Boot |
| AI/ML APIs | 10 | OpenAI, Anthropic, Cohere, Google AI, Mistral |
| AI/ML Frameworks | 5 | LangChain, TensorFlow, PyTorch, Hugging Face |
| Vector Databases | 5 | Pinecone, Weaviate, ChromaDB, Qdrant, Milvus |
| Cloud Providers | 5 | AWS, GCP, Azure, Cloudflare, DigitalOcean |
| Databases | 8 | PostgreSQL, MongoDB, Redis, MySQL, Elasticsearch |
| DevOps/IaC | 10 | Docker, Kubernetes, Terraform, GitHub Actions |
| Testing Frameworks | 6 | Jest, Pytest, Cypress, Playwright, Vitest |
| Authentication | 3 | Auth0, Okta, AWS Cognito |
| Monitoring | 5 | Datadog, Sentry, New Relic, Grafana |
| CNCF Projects | 20+ | Envoy, Istio, Prometheus, ArgoCD, Flux |

**Detection Evidence:**
```json
{
  "name": "React",
  "category": "web-frameworks/frontend",
  "confidence": 95,
  "detection_tier": 1,
  "detection_methods": ["package", "import"],
  "evidence": [
    {"type": "package", "source": "package.json", "value": "react@18.2.0"},
    {"type": "import", "file": "src/App.tsx", "line": 1}
  ]
}
```

### 2. AI Models (`models`)

Detects AI/ML models through file patterns, code analysis, and registry queries.

**Configuration:**
```json
{
  "models": {
    "enabled": true,
    "detect_model_files": true,
    "scan_code_patterns": true,
    "scan_configs": true,
    "query_registries": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable model detection |
| `detect_model_files` | bool | `true` | Scan for model files |
| `scan_code_patterns` | bool | `true` | Detect model loading in code |
| `scan_configs` | bool | `true` | Check config files for model refs |
| `query_registries` | bool | `true` | Query HuggingFace/Replicate APIs |

**Detected Model Formats:**

| Extension | Format | Risk Level |
|-----------|--------|------------|
| `.pt`, `.pth` | PyTorch | Low |
| `.h5`, `.hdf5` | HDF5/Keras | Low |
| `.onnx` | ONNX | Low |
| `.safetensors` | SafeTensors | Low (Safe) |
| `.pkl`, `.pickle` | Pickle | High (Arbitrary code execution) |
| `.pb`, `.pbtxt` | TensorFlow | Low |
| `.mlmodel` | Core ML | Low |
| `.tflite` | TensorFlow Lite | Low |
| `.gguf`, `.ggml` | GGML/llama.cpp | Low |

**Model Detection in Code:**

| Pattern | Source |
|---------|--------|
| `AutoModel.from_pretrained("...")` | Hugging Face Transformers |
| `pipeline("...", model="...")` | Hugging Face Pipeline |
| `replicate.run("...")` | Replicate API |
| `openai.ChatCompletion.create(model="...")` | OpenAI API |
| `anthropic.messages.create(model="...")` | Anthropic API |

**Registry Enrichment:**
When `query_registries` is enabled, fetches metadata from:
- **HuggingFace Hub**: Model cards, licenses, tags, downloads
- **Replicate**: Model versions, run counts

### 3. AI Frameworks (`frameworks`)

Detects AI/ML frameworks and libraries.

**Configuration:**
```json
{
  "frameworks": {
    "enabled": true,
    "detect_deep_learning": true,
    "detect_llm": true,
    "detect_mlops": true
  }
}
```

**Detected Framework Categories:**

| Category | Frameworks |
|----------|------------|
| Deep Learning | PyTorch, TensorFlow, Keras, JAX, MXNet |
| LLM/GenAI | transformers, langchain, llama-cpp, vllm, openai, anthropic |
| ML Libraries | scikit-learn, XGBoost, LightGBM, CatBoost |
| MLOps | MLflow, Kubeflow, DVC, Weights & Biases |
| Vector DBs | pinecone, weaviate, chroma, milvus, qdrant |

### 4. Datasets (`datasets`)

Detects dataset references and provenance.

**Configuration:**
```json
{
  "datasets": {
    "enabled": true,
    "detect_references": true,
    "check_provenance": true
  }
}
```

**Detected Patterns:**
- `load_dataset("...")` - Hugging Face Datasets
- `datasets/...` paths - HuggingFace Hub references
- Local data files (`.csv`, `.parquet`, `.jsonl`, `.arrow`)
- Cloud storage references (`s3://`, `gs://`)

### 5. AI Security (`ai_security`)

Detects AI/ML-specific security vulnerabilities.

**Configuration:**
```json
{
  "ai_security": {
    "enabled": true,
    "check_pickle_files": true,
    "detect_unsafe_loading": true,
    "check_api_key_exposure": true
  }
}
```

**Security Issues Detected:**

| Issue | Severity | CWE | Description |
|-------|----------|-----|-------------|
| Pickle Model | High | CWE-502 | Pickle files can execute arbitrary code |
| `torch.load()` without `weights_only` | High | CWE-502 | Unsafe PyTorch deserialization |
| Hardcoded API Key | Critical | CWE-798 | API keys in code |
| HuggingFace Token Exposure | High | CWE-798 | HF_TOKEN in code |
| OpenAI Key Exposure | Critical | CWE-798 | OPENAI_API_KEY exposed |

### 6. AI Governance (`ai_governance`)

Checks AI/ML governance compliance.

**Configuration:**
```json
{
  "ai_governance": {
    "enabled": true,
    "require_model_cards": true,
    "require_license": true,
    "require_dataset_info": false,
    "blocked_licenses": ["CC-BY-NC-4.0", "proprietary"]
  }
}
```

**Governance Checks:**
- Model card documentation (README.md, MODEL_CARD.md)
- License documentation and compliance
- Dataset cards for training data
- Blocked license detection

### 7. Infrastructure Extraction (`infrastructure`)

Extracts infrastructure configuration values (Tier 3).

**Configuration:**
```json
{
  "infrastructure": {
    "enabled": true,
    "extract_registries": true,
    "extract_cloud_accounts": true,
    "extract_clusters": true
  }
}
```

**Extracted Values:**
- Container registry URLs (ECR, GCR, ACR, DockerHub)
- AWS account IDs from ARNs
- Kubernetes cluster names and endpoints
- Cloud project/subscription IDs

## How It Works

### Technical Flow

1. **RAG Pattern Loading**: Loads patterns from `rag/technology-identification/`
2. **Tier 1 (Quick)**: Analyzes SBOM packages if available
3. **Tier 2 (Deep)**: Scans files for extensions, configs, imports
4. **Tier 3 (Extract)**: Parses DevOps configs for values
5. **AI Detection**: Model files, frameworks, datasets
6. **Security Scan**: Pickle files, unsafe loading, API keys
7. **Governance Check**: Model cards, licenses
8. **ML-BOM Generation**: Combines model/framework/dataset info

### Architecture

```
Repository
    │
    ├─► Detection Feature ─────► RAG Patterns ─────► Technology Inventory
    │       │
    │       ├─► Tier 1: SBOM packages
    │       ├─► Tier 2: Files, configs, imports
    │       └─► Tier 3: Infrastructure extraction
    │
    ├─► Models Feature ────────► File scan + Registry queries ───► Model Inventory
    │
    ├─► Frameworks Feature ────► Package detection ───► Framework List
    │
    ├─► Datasets Feature ──────► Reference detection ───► Dataset List
    │
    ├─► AI Security Feature ───► Vulnerability scan ───► Security Findings
    │
    ├─► AI Governance Feature ─► Compliance checks ───► Governance Issues
    │
    └─► Infrastructure Feature ► Config parsing ───► Infrastructure Map
```

### RAG Pattern Structure

Patterns are stored in `rag/technology-identification/`:

```
rag/technology-identification/
├── README.md                    # Overview and usage
├── confidence-config.md         # Confidence scoring rules
├── languages/
│   ├── python/patterns.md
│   ├── javascript/patterns.md
│   └── ...
├── web-frameworks/
│   ├── frontend/
│   │   ├── react/patterns.md
│   │   └── vue/patterns.md
│   └── backend/
│       ├── express/patterns.md
│       └── django/patterns.md
├── ai-ml/
│   ├── apis/
│   │   ├── openai/patterns.md
│   │   └── anthropic/patterns.md
│   ├── frameworks/
│   │   └── langchain/patterns.md
│   └── vectordb/
│       └── pinecone/patterns.md
└── ...
```

## Usage

### Command Line

```bash
# Run technology scanner only
./zero scan --scanner technology /path/to/repo

# Run with quick detection (Tier 1 only)
./zero scan --scanner technology --tier quick /path/to/repo

# Run with deep detection (all tiers)
./zero scan --scanner technology --tier deep /path/to/repo
```

### Configuration Profiles

| Profile | detection | models | frameworks | datasets | ai_security | ai_governance |
|---------|-----------|--------|------------|----------|-------------|---------------|
| `quick` | Tier 1 | - | - | - | - | - |
| `standard` | Tier 1+2 | Yes | Yes | - | - | - |
| `full` | All tiers | Yes | Yes | Yes | Yes | Yes |
| `ai-security` | Tier 1+2 | Yes | Yes | Yes | Yes | Yes |

## Output Format

```json
{
  "scanner": "technology",
  "version": "1.0.0",
  "metadata": {
    "features_run": ["detection", "models", "frameworks", "datasets", "ai_security", "ai_governance"]
  },
  "summary": {
    "detection": {
      "total_technologies": 45,
      "by_category": {
        "languages": 4,
        "web-frameworks": 5,
        "databases": 3,
        "ai-ml": 8,
        "devops": 6
      },
      "detection_tier": 2
    },
    "models": {
      "total_models": 5,
      "by_source": {"huggingface": 3, "local": 2},
      "by_format": {"safetensors": 2, "pytorch": 2, "pickle": 1}
    },
    "frameworks": {
      "total_frameworks": 8,
      "by_category": {"deep_learning": 2, "llm": 3, "mlops": 2, "vector_db": 1}
    },
    "datasets": {
      "total_datasets": 3,
      "by_source": {"huggingface": 2, "local": 1}
    },
    "ai_security": {
      "total_findings": 4,
      "critical": 1,
      "high": 2,
      "medium": 1
    },
    "ai_governance": {
      "models_missing_cards": 1,
      "blocked_license_violations": 0
    }
  },
  "findings": {
    "detection": {
      "technologies": [
        {
          "name": "Python",
          "category": "languages",
          "confidence": 95,
          "detection_tier": 1,
          "evidence": [...]
        },
        {
          "name": "LangChain",
          "category": "ai-ml/frameworks",
          "confidence": 90,
          "detection_tier": 1,
          "evidence": [{"type": "package", "source": "requirements.txt", "value": "langchain==0.1.0"}]
        }
      ]
    },
    "models": [...],
    "frameworks": [...],
    "datasets": [...],
    "ai_security": [...],
    "ai_governance": [...],
    "infrastructure": {
      "container_registries": [...],
      "cloud_accounts": [...],
      "kubernetes_clusters": [...]
    }
  },
  "ml_bom": {
    "models": [...],
    "frameworks": [...],
    "datasets": [...]
  }
}
```

## Prerequisites

No required external tools. Optional for enhanced detection:

| Service | Purpose | Required |
|---------|---------|----------|
| HuggingFace Hub | Model metadata enrichment | No |
| Replicate API | Model metadata enrichment | No |

**Environment Variables (optional):**
- `HF_TOKEN` or `HUGGING_FACE_HUB_TOKEN`: For private model access
- `REPLICATE_API_TOKEN`: For Replicate API access

## Related Scanners

- **sbom**: Provides package data for Tier 1 detection enrichment
- **packages**: Detects framework dependencies
- **code-security**: May detect overlapping secrets
- **health**: Uses technology data for project assessment

## See Also

- [RAG Technology Patterns](../../rag/technology-identification/README.md) - Pattern database
- [Packages Scanner](packages.md) - Dependency analysis
- [Code Security Scanner](code-security.md) - Secret detection
- [OWASP ML Security Top 10](https://owasp.org/www-project-machine-learning-security-top-10/)
- [HuggingFace Model Hub](https://huggingface.co/models)
