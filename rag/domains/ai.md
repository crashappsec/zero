# AI/ML Security Domain Knowledge

## Overview

This document consolidates domain knowledge for the AI super scanner, covering ML model security, ML-BOM generation, AI framework analysis, and AI governance.

## ML-BOM (Machine Learning Bill of Materials)

### What is ML-BOM?

ML-BOM extends traditional SBOM concepts to AI/ML systems, documenting:
- **Models**: Name, version, source, architecture, license
- **Datasets**: Training data provenance and licensing
- **Frameworks**: AI/ML libraries and versions
- **Lineage**: Base models, fine-tuning chains
- **Governance**: Model cards, intended use, limitations

### CycloneDX ML-BOM Format

CycloneDX 1.5+ supports ML-BOM with the `modelCard` component:

```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.6",
  "components": [{
    "type": "machine-learning-model",
    "name": "bert-base-uncased",
    "version": "1.0",
    "supplier": { "name": "Google" },
    "licenses": [{ "license": { "id": "Apache-2.0" }}],
    "modelCard": {
      "modelParameters": {
        "architecture": { "family": "transformer" },
        "datasets": [{ "ref": "wikipedia", "type": "training" }]
      },
      "considerations": {
        "environmentalConsiderations": { ... }
      }
    }
  }]
}
```

## Model File Format Security

### High Risk: Pickle-Based Formats

**Extensions**: `.pt`, `.pth`, `.pkl`, `.pickle`, `.bin`

Pickle is Python's native serialization format. It can execute arbitrary code during deserialization:

```python
# Malicious pickle payload example (DO NOT USE)
import pickle
class Malicious:
    def __reduce__(self):
        return (os.system, ('rm -rf /',))

# Loading this executes the command
pickle.load(malicious_file)  # DANGEROUS
```

**Prevalence**: ~45% of ML model repositories contain pickle files.

**Mitigation**:
- Convert to SafeTensors format
- Use `torch.load(path, weights_only=True)` (PyTorch 2.0+)
- Never load pickle files from untrusted sources

### Medium Risk: ONNX and Keras

**ONNX** (`.onnx`):
- Custom operators can execute native code (C++/CUDA)
- Complex control flow (If/Loop nodes) can have unexpected behavior
- No observed real-world attacks, but theoretically vulnerable

**Keras** (`.h5`, `.keras`):
- Lambda layers can contain arbitrary Python code
- Custom layers may execute code during loading
- Prefer functional API over Lambda layers

### Low Risk: Secure Formats

**SafeTensors** (`.safetensors`):
- Developed by HuggingFace as pickle replacement
- Only stores tensor data, no code execution
- Zero-copy loading for performance
- Independently security audited

**GGUF/GGML** (`.gguf`, `.ggml`):
- Inference-only format for llama.cpp
- Contains weights and tokenizer, no code
- Designed for efficient local inference

**TensorFlow Lite** (`.tflite`):
- Mobile inference format
- Limited operation set, no custom ops
- Sandboxed execution

## AI Framework Security

### Deep Learning Frameworks

| Framework | Package | Key Security Concerns |
|-----------|---------|----------------------|
| PyTorch | `torch` | Pickle model loading, JIT compilation |
| TensorFlow | `tensorflow` | SavedModel custom ops, TF1 graph execution |
| JAX | `jax` | JIT compilation, XLA vulnerabilities |

### LLM Frameworks

| Framework | Package | Key Security Concerns |
|-----------|---------|----------------------|
| LangChain | `langchain` | Prompt injection, arbitrary tool execution |
| LlamaIndex | `llama_index` | Document parsing vulnerabilities |
| HuggingFace | `transformers` | Model download from untrusted repos |

### MLOps Tools

| Tool | Package | Security Notes |
|------|---------|----------------|
| MLflow | `mlflow` | Model registry authentication |
| Weights & Biases | `wandb` | API key management |
| DVC | `dvc` | Remote storage credentials |

## LLM API Security

### API Key Patterns

| Provider | Pattern | Example |
|----------|---------|---------|
| OpenAI | `sk-[a-zA-Z0-9]{48}` | `sk-abc123...` |
| Anthropic | `sk-ant-[a-zA-Z0-9-]+` | `sk-ant-api01-...` |
| HuggingFace | `hf_[a-zA-Z0-9]{34}` | `hf_abc123...` |
| Cohere | `[a-zA-Z0-9]{40}` | (generic pattern) |
| Google AI | `AIza[a-zA-Z0-9_-]{35}` | `AIzaSy...` |

### Secure API Key Handling

```python
# WRONG - hardcoded key
client = OpenAI(api_key="sk-abc123...")

# CORRECT - environment variable
client = OpenAI(api_key=os.environ["OPENAI_API_KEY"])

# BETTER - secrets manager
from secretmanager import get_secret
client = OpenAI(api_key=get_secret("openai-api-key"))
```

## Model Provenance and Supply Chain

### Model Registries

Zero's AI scanner supports detection and metadata enrichment from multiple model registries:

| Registry | Detection | API | Description |
|----------|-----------|-----|-------------|
| HuggingFace Hub | ✅ | ✅ | Largest open ML model repository (400k+ models) |
| TensorFlow Hub | ✅ | ❌ | Google's repository for reusable TensorFlow models |
| PyTorch Hub | ✅ | ❌ | Official PyTorch model repository |
| Replicate | ✅ | ✅ | Cloud ML platform with versioned models |
| Weights & Biases | ✅ | ✅ | MLOps platform with model artifacts |
| MLflow | ✅ | ❌ | Self-hosted MLOps model registry |
| Civitai | ✅ | ✅ | Community platform for Stable Diffusion models |
| Kaggle Models | ✅ | ❌ | Kaggle's ML model repository |
| Ollama | ✅ | ❌ | Local LLM model library |
| NVIDIA NGC | ✅ | ✅ | NVIDIA's GPU-optimized model catalog |
| AWS SageMaker JumpStart | ✅ | ❌ | AWS ML model catalog |
| Azure ML Model Catalog | ✅ | ❌ | Microsoft Azure ML repository |

### Model Sources Trust Levels

| Source | Trust Level | Verification |
|--------|-------------|--------------|
| HuggingFace Hub | Medium | Check author, downloads, model card |
| TensorFlow Hub | High | Google-maintained |
| PyTorch Hub | Medium | Check repository trust |
| Replicate | Medium | Check author, run count, version history |
| NVIDIA NGC | High | NVIDIA-curated models |
| AWS JumpStart | High | AWS-curated models |
| Azure ML | High | Microsoft-curated models |
| Civitai | Low | Community-uploaded, high variation |
| Ollama | Medium | Pre-converted GGUF models |
| Local files | Unknown | Must verify origin |
| API (OpenAI, etc.) | High | Closed source, audit logs |

### Supply Chain Risks

1. **Model Poisoning**: Backdoored models with hidden behaviors
2. **Data Poisoning**: Compromised training data affecting outputs
3. **Pickle RCE**: Malicious code in model files
4. **Dependency Confusion**: Typosquatted model names
5. **License Violations**: Models trained on copyrighted data

### Verification Checklist

- [ ] Model source is trusted (official repos, verified authors)
- [ ] Model card exists with training details
- [ ] License is compatible with intended use
- [ ] Model format is secure (safetensors preferred)
- [ ] Base model lineage is documented
- [ ] No known vulnerabilities in version

## AI Governance

### Model Card Requirements

Per Google's Model Card framework:
1. **Model Details**: Name, version, type, license
2. **Intended Use**: Primary use cases, out-of-scope uses
3. **Training Data**: Datasets used, preprocessing
4. **Evaluation**: Metrics, benchmarks, limitations
5. **Ethical Considerations**: Bias, fairness, privacy
6. **Environmental Impact**: Training compute, carbon footprint

### License Considerations

| License | Commercial Use | Derivatives | Common Models |
|---------|----------------|-------------|---------------|
| Apache-2.0 | Yes | Yes | BERT, T5, DistilBERT |
| MIT | Yes | Yes | Many research models |
| CC-BY-4.0 | Yes | Yes w/ attribution | LLaMA 2 (some) |
| CC-BY-NC-4.0 | No | No commercial | Many research models |
| OpenRAIL | Yes | Yes w/ restrictions | Stable Diffusion |
| Proprietary | Varies | No | GPT-4, Claude |

### Regulatory Landscape

- **EU AI Act**: Risk-based regulation, transparency requirements
- **US Executive Orders**: Federal AI procurement standards
- **NIST AI RMF**: Risk management framework
- **ISO/IEC 42001**: AI management system standard

## Security Findings Reference

### MLSEC-001: Unsafe Pickle Model File

**Severity**: High
**Category**: pickle_rce
**CWE**: CWE-502 (Deserialization of Untrusted Data)

**Description**: Model file uses pickle format which allows arbitrary code execution during loading.

**Remediation**:
```python
# Convert PyTorch model to SafeTensors
from safetensors.torch import save_file
save_file(model.state_dict(), "model.safetensors")
```

### MLSEC-002: Unsafe torch.load() Usage

**Severity**: High
**Category**: unsafe_loading
**CWE**: CWE-502

**Description**: `torch.load()` called without `weights_only=True` parameter.

**Remediation**:
```python
# Add weights_only parameter
model = torch.load("model.pt", weights_only=True)

# Or better, use safetensors
from safetensors.torch import load_file
model = load_file("model.safetensors")
```

### MLSEC-003: Hardcoded API Key

**Severity**: Critical
**Category**: api_key_exposure
**CWE**: CWE-798 (Hardcoded Credentials)

**Description**: LLM API key found hardcoded in source code.

**Remediation**:
```python
# Use environment variables
import os
api_key = os.environ.get("OPENAI_API_KEY")

# Or use python-dotenv
from dotenv import load_dotenv
load_dotenv()
```

## Tools and Resources

### Open Source Tools

- **cdxgen**: CycloneDX SBOM generator with ML profile
- **OWASP AIBOM Generator**: HuggingFace model card to CycloneDX
- **Safetensors**: Secure model serialization
- **ModelScan**: Detect malicious ML models

### Standards and Specifications

- **CycloneDX 1.6**: ML-BOM specification
- **SPDX AI Profile**: Software Package Data Exchange for AI
- **Model Cards**: Google's model documentation framework
- **Hugging Face Model Hub**: Model hosting and documentation

### Research References

- "PickleBall: Secure Deserialization of Pickle-based ML Models" (2024)
- "Atlas: ML Lifecycle Provenance & Transparency" (2025)
- "TAIBOM: Trustworthiness for AI-Enabled Systems" (2025)
