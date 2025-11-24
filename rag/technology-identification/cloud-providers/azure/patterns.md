# Microsoft Azure

**Category**: cloud-providers
**Description**: Microsoft Azure SDK and services
**Homepage**: https://azure.microsoft.com

## Package Detection

### NPM
*Azure Node.js SDKs*

- `@azure/storage-blob`
- `@azure/identity`
- `@azure/cosmos`
- `@azure/service-bus`
- `@azure/functions`

### PYPI
*Azure Python SDKs*

- `azure-storage-blob`
- `azure-identity`
- `azure-cosmos`
- `azure-servicebus`
- `azure-functions`

### NUGET
*Azure .NET SDKs*

- `Azure.Storage.Blobs`
- `Azure.Identity`
- `Microsoft.Azure.Cosmos`

### MAVEN
*Azure Java SDKs*

- `com.azure:azure-storage-blob`
- `com.azure:azure-identity`

## Import Detection

### Javascript

**Pattern**: `from\s+['"]@azure/`
- Type: esm_import

**Pattern**: `require\(['"]@azure/`
- Type: commonjs_require

### Python

**Pattern**: `from\s+azure\.`
- Type: python_import

**Pattern**: `import\s+azure\.`
- Type: python_import

### Csharp

**Pattern**: `using\s+Azure\.`
- Type: csharp_using

**Pattern**: `using\s+Microsoft\.Azure\.`
- Type: csharp_using

## Environment Variables

*Azure subscription ID*

*Azure tenant ID*

*Azure client/app ID*

*Azure client secret*

*Azure Storage connection*

*Azure Storage account name*


## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Import Detection**: 90% (HIGH)
- **Environment Variable Detection**: 85% (MEDIUM)
- **API Endpoint Detection**: 80% (MEDIUM)
