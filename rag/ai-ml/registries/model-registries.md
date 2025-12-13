# Model Registries

Reference guide for ML model registries, their trust levels, and verification methods.

## Trust Levels

- **High**: Curated by major cloud providers (AWS, Azure, Google, NVIDIA)
- **Medium**: Popular platforms with community vetting (HuggingFace, PyTorch Hub)
- **Low**: Community platforms with minimal vetting (Civitai)
- **Internal**: Self-hosted registries (MLflow)

## Public Registries

### HuggingFace Hub

| Property | Value |
|----------|-------|
| Name | HuggingFace Hub |
| Base URL | https://huggingface.co |
| API URL | https://huggingface.co/api/models |
| Has API | Yes |
| Trust Level | Medium |
| Description | Largest open ML model repository with 400k+ models |
| Verification | Check author, downloads, model card |
| Model URL | `https://huggingface.co/{model_name}` |

### TensorFlow Hub

| Property | Value |
|----------|-------|
| Name | TensorFlow Hub |
| Base URL | https://tfhub.dev |
| Has API | No |
| Trust Level | High |
| Description | Google's repository for reusable TensorFlow models |
| Verification | Google-maintained |
| Model URL | `https://tfhub.dev/{model_name}` |

### PyTorch Hub

| Property | Value |
|----------|-------|
| Name | PyTorch Hub |
| Base URL | https://pytorch.org/hub |
| Has API | No |
| Trust Level | Medium |
| Description | Official PyTorch model repository |
| Verification | Check repository trust |
| Model URL | `https://github.com/{owner}/{repo}` |

### Replicate

| Property | Value |
|----------|-------|
| Name | Replicate |
| Base URL | https://replicate.com |
| API URL | https://api.replicate.com/v1/models |
| Has API | Yes |
| Trust Level | Medium |
| Description | Cloud ML platform with versioned models |
| Verification | Check author, run count, version history |
| Model URL | `https://replicate.com/{model_name}` |

### Weights & Biases

| Property | Value |
|----------|-------|
| Name | Weights & Biases |
| Base URL | https://wandb.ai |
| API URL | https://api.wandb.ai/artifacts |
| Has API | Yes |
| Trust Level | Medium |
| Description | MLOps platform with model artifacts |
| Verification | Check organization, artifact history |
| Model URL | `https://wandb.ai/artifacts/{model_name}` |

### Kaggle Models

| Property | Value |
|----------|-------|
| Name | Kaggle Models |
| Base URL | https://kaggle.com/models |
| Has API | No |
| Trust Level | Medium |
| Description | Kaggle's ML model repository |
| Verification | Check author reputation, competition results |
| Model URL | `https://kaggle.com/models/{model_name}` |

### Ollama Library

| Property | Value |
|----------|-------|
| Name | Ollama Library |
| Base URL | https://ollama.com/library |
| Has API | No |
| Trust Level | Medium |
| Description | Local LLM model library for ollama |
| Verification | Pre-converted GGUF models, check source model |
| Model URL | `https://ollama.com/library/{model_name}` |

### Civitai

| Property | Value |
|----------|-------|
| Name | Civitai |
| Base URL | https://civitai.com |
| API URL | https://civitai.com/api/v1/models |
| Has API | Yes |
| Trust Level | Low |
| Description | Community platform for Stable Diffusion models |
| Verification | Community-uploaded, high variation - scan carefully |
| Model URL | `https://civitai.com/models/{model_id}` |

### Roboflow Universe

| Property | Value |
|----------|-------|
| Name | Roboflow Universe |
| Base URL | https://universe.roboflow.com |
| API URL | https://api.roboflow.com |
| Has API | Yes |
| Trust Level | Medium |
| Description | Computer vision model and dataset platform |
| Verification | Check workspace, model metrics |
| Model URL | `https://universe.roboflow.com/{workspace}/{project}` |

### ModelHub.ai

| Property | Value |
|----------|-------|
| Name | ModelHub.ai |
| Base URL | https://modelhub.ai |
| Has API | No |
| Trust Level | Medium |
| Description | Deep learning model sharing platform |
| Verification | Check model documentation |
| Model URL | `https://modelhub.ai/{model_name}` |

## Cloud Provider Registries

### NVIDIA NGC

| Property | Value |
|----------|-------|
| Name | NVIDIA NGC |
| Base URL | https://catalog.ngc.nvidia.com |
| API URL | https://api.ngc.nvidia.com/v2/models |
| Has API | Yes |
| Trust Level | High |
| Description | NVIDIA's GPU-optimized model catalog |
| Verification | NVIDIA-curated models |
| Model URL | `https://catalog.ngc.nvidia.com/orgs/nvidia/models/{model_name}` |

### AWS SageMaker JumpStart

| Property | Value |
|----------|-------|
| Name | AWS SageMaker JumpStart |
| Base URL | https://aws.amazon.com/sagemaker/jumpstart |
| Has API | No |
| Trust Level | High |
| Description | AWS ML model catalog for SageMaker |
| Verification | AWS-curated models |

### Azure ML Model Catalog

| Property | Value |
|----------|-------|
| Name | Azure ML Model Catalog |
| Base URL | https://ml.azure.com |
| Has API | No |
| Trust Level | High |
| Description | Microsoft Azure ML model repository |
| Verification | Microsoft-curated models |

### Google Vertex AI Model Garden

| Property | Value |
|----------|-------|
| Name | Google Vertex AI Model Garden |
| Base URL | https://cloud.google.com/vertex-ai/docs/model-garden |
| Has API | No |
| Trust Level | High |
| Description | Google Cloud's curated model catalog |
| Verification | Google-curated models |

## Self-Hosted Registries

### MLflow Model Registry

| Property | Value |
|----------|-------|
| Name | MLflow Model Registry |
| Has API | No |
| Trust Level | Internal |
| Description | Self-hosted MLOps model registry |
| Verification | Internal deployment - verify source |
| Documentation | https://mlflow.org/docs/latest/model-registry.html |

## Security Recommendations

1. **Verify Model Sources**: Always check model provenance before deployment
2. **Prefer High Trust**: Use cloud provider registries for production
3. **Scan Community Models**: Extra scrutiny for Civitai and user-uploaded models
4. **Check Model Cards**: Review model documentation and intended use
5. **Monitor Downloads**: Verify download counts and community adoption
6. **Use SafeTensors**: Prefer safe formats over pickle-based models
