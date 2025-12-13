# Model File Format Security

Model file formats have varying levels of security risk based on their ability to execute arbitrary code during loading.

## Risk Levels

- **High Risk**: Can execute arbitrary code during deserialization
- **Medium Risk**: May contain executable components or custom operators
- **Low Risk**: Inference-only formats with no code execution capability

## Format Reference

### High Risk Formats (Pickle-based)

| Extension | Name | Risk | CWE | Remediation |
|-----------|------|------|-----|-------------|
| `.pt` | PyTorch Pickle | High | CWE-502 | Convert to SafeTensors format |
| `.pth` | PyTorch Pickle | High | CWE-502 | Convert to SafeTensors format |
| `.pkl` | Python Pickle | High | CWE-502 | Use safer serialization like JSON, or convert model to SafeTensors |
| `.pickle` | Python Pickle | High | CWE-502 | Use safer serialization like JSON, or convert model to SafeTensors |

### Medium Risk Formats

| Extension | Name | Risk | CWE | Remediation |
|-----------|------|------|-----|-------------|
| `.bin` | Binary Weights | Medium | CWE-502 | Verify format is weights-only, consider SafeTensors |
| `.onnx` | ONNX | Medium | CWE-94 | Review custom operators, use only trusted models |
| `.h5` | HDF5/Keras | Medium | CWE-94 | Convert Lambda layers to functional API |
| `.keras` | Keras | Medium | CWE-94 | Convert Lambda layers to functional API |
| `.pb` | TensorFlow SavedModel | Medium | CWE-94 | Review custom ops, use trusted models only |
| `.engine` | TensorRT Engine | Medium | CWE-94 | Only use engines compiled from trusted sources |
| `.plan` | TensorRT Plan | Medium | CWE-94 | Only use plans compiled from trusted sources |

### Low Risk Formats (Safe)

| Extension | Name | Risk | Notes |
|-----------|------|------|-------|
| `.safetensors` | SafeTensors | Low | Secure format, no code execution possible |
| `.gguf` | GGUF | Low | Inference-only format, no code execution |
| `.ggml` | GGML | Low | Inference-only format, no code execution |
| `.tflite` | TensorFlow Lite | Low | Mobile inference format with limited op set |
| `.mlmodel` | Core ML | Low | Apple inference format with sandbox |
| `.mlpackage` | Core ML Package | Low | Apple inference format with sandbox |

## Safe Loading Patterns

### PyTorch

```python
# Unsafe (allows arbitrary code execution)
torch.load(path)

# Safe (weights only)
torch.load(path, weights_only=True)

# Safest (use SafeTensors)
from safetensors.torch import load_file
model_weights = load_file(path)
```

### TensorFlow

```python
# Unsafe (may execute Lambda layers)
tf.keras.models.load_model(path)

# Safe (blocks unsafe operations)
tf.keras.models.load_model(path, safe_mode=True)
```

## Security Recommendations

1. **Prefer SafeTensors**: Convert models to `.safetensors` format when possible
2. **Verify Sources**: Only load models from trusted sources and verified authors
3. **Scan for Risks**: Use tools to detect pickle-based formats before deployment
4. **Isolate Loading**: Load untrusted models in sandboxed environments
5. **Review Custom Ops**: Audit any custom operators in ONNX/TensorFlow models
