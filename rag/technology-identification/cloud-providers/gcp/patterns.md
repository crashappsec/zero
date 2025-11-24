# Google Cloud Platform

**Category**: cloud-providers
**Description**: Google Cloud Platform SDK and services
**Homepage**: https://cloud.google.com

## Package Detection

### NPM
*GCP Node.js client libraries*

- `@google-cloud/storage`
- `@google-cloud/pubsub`
- `@google-cloud/firestore`
- `@google-cloud/bigquery`
- `@google-cloud/functions-framework`

### PYPI
*GCP Python client libraries*

- `google-cloud-storage`
- `google-cloud-pubsub`
- `google-cloud-firestore`
- `google-cloud-bigquery`
- `google-api-python-client`

### MAVEN
*GCP Java client libraries*

- `com.google.cloud:google-cloud-storage`
- `com.google.cloud:google-cloud-pubsub`

### GO
*GCP Go client libraries*

- `cloud.google.com/go/storage`
- `cloud.google.com/go/pubsub`
- `cloud.google.com/go/firestore`

## Import Detection

### Javascript

**Pattern**: `from\s+['"]@google-cloud/`
- Type: esm_import

**Pattern**: `require\(['"]@google-cloud/`
- Type: commonjs_require

### Python

**Pattern**: `from\s+google\.cloud`
- Type: python_import

**Pattern**: `import\s+google\.cloud`
- Type: python_import

### Go

**Pattern**: `"cloud\.google\.com/go/`
- Type: go_import

## Environment Variables

*Path to service account JSON*

*GCP project ID*

*GCP project ID*

*GCP project ID*

*GCP region*

*Google Cloud Storage bucket*


## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Import Detection**: 90% (HIGH)
- **Environment Variable Detection**: 85% (MEDIUM)
- **API Endpoint Detection**: 80% (MEDIUM)
