# Model Loading Detection Patterns

Code patterns for detecting model loading across different ML frameworks and registries.

## HuggingFace

### Transformers from_pretrained

```regex
(?:AutoModel|AutoTokenizer|AutoProcessor|AutoFeatureExtractor|AutoConfig|pipeline)\s*\.?\s*from_pretrained\s*\(\s*["']([^"']+)["']
```

Captures model name from HuggingFace Transformers `from_pretrained` calls.

### Generic from_pretrained

```regex
from_pretrained\s*\(\s*["']([^"']+)["']
```

Captures any `from_pretrained` call.

### HuggingFace Hub Download

```regex
hf_hub_download\s*\([^)]*repo_id\s*=\s*["']([^"']+)["']
```

Captures repository ID from Hub downloads.

### Sentence Transformers

```regex
SentenceTransformer\s*\(\s*["']([^"']+)["']
```

Captures model name from Sentence Transformers.

### vLLM

```regex
LLM\s*\(\s*(?:model\s*=\s*)?["']([^"']+)["']
```

Captures model from vLLM initialization.

### Text Generation Inference

```regex
--model-id\s+["']?([^\s"']+)
```

Captures model ID from TGI CLI arguments.

## PyTorch Hub

```regex
torch\.hub\.load\s*\(\s*["']([^"']+)["']\s*,\s*["']([^"']+)["']
```

Captures repository and model name, joined with `/`.

## TensorFlow Hub

### KerasLayer/Load

```regex
hub\.(?:KerasLayer|load)\s*\(\s*["']([^"']+)["']
```

Captures TensorFlow Hub model URL.

### TFHub URL

```regex
["'](https?://tfhub\.dev/[^"']+)["']
```

Captures direct TFHub URLs.

## Replicate

### Run

```regex
replicate\.run\s*\(\s*["']([^"']+)["']
```

Captures model version from Replicate run.

### Client Run

```regex
Replicate\s*\([^)]*\)\.run\s*\(\s*["']([^"']+)["']
```

Captures model from Replicate client.

## Weights & Biases

### Use Artifact

```regex
wandb\.use_artifact\s*\(\s*["']([^"']+)["']
```

Captures W&B artifact reference.

### API Artifact

```regex
wandb\.Api\s*\([^)]*\)\.artifact\s*\(\s*["']([^"']+)["']
```

Captures artifact from W&B API.

## MLflow

### Load Model

```regex
mlflow\.(?:pyfunc|sklearn|pytorch|tensorflow|keras)\.load_model\s*\(\s*["']([^"']+)["']
```

Captures MLflow model URI.

### Models URI

```regex
["'](models:/[^"']+)["']
```

Captures MLflow `models:/` URIs.

### Runs URI

```regex
["'](runs:/[^"']+)["']
```

Captures MLflow `runs:/` URIs.

## Kaggle

### Model Get

```regex
kaggle\.api\.model_get\s*\([^)]*model\s*=\s*["']([^"']+)["']
```

Captures Kaggle model reference.

### KaggleHub Download

```regex
kagglehub\.model_download\s*\(\s*["']([^"']+)["']
```

Captures model from KaggleHub.

## Civitai

```regex
["'](https?://civitai\.com/(?:api/download/)?models/[^"']+)["']
```

Captures Civitai model URLs.

## NVIDIA NGC

### Container Reference

```regex
["'](nvcr\.io/[^"']+)["']
```

Captures NGC container references.

### CLI Download

```regex
ngc\s+(?:registry\s+)?model\s+download[^"']*["']([^"']+)["']
```

Captures model from NGC CLI.

## AWS SageMaker JumpStart

### JumpStart Model

```regex
JumpStartModel\s*\(\s*model_id\s*=\s*["']([^"']+)["']
```

Captures JumpStart model ID.

### Model URIs

```regex
sagemaker\.(?:model_uris|image_uris)\.[^(]+\([^)]*model_id\s*=\s*["']([^"']+)["']
```

Captures model from SageMaker URIs.

## Azure ML

### Model

```regex
Model\s*\([^)]*name\s*=\s*["']([^"']+)["'][^)]*workspace
```

Captures Azure ML model name.

### Registry

```regex
azureml://registries/[^/]+/models/([^/"']+)
```

Captures model from Azure ML registry URI.

## Ollama

```regex
ollama\.(?:chat|generate|embeddings)\s*\([^)]*model\s*=\s*["']([^"']+)["']
```

Captures Ollama model name.

## Roboflow

### Project

```regex
roboflow\.Roboflow\s*\([^)]*\)\.workspace\([^)]*\)\.project\([^)]*["']([^"']+)["']
```

Captures Roboflow project.

### Inference

```regex
inference\.get_model\s*\([^)]*model_id\s*=\s*["']([^"']+)["']
```

Captures Roboflow inference model.
