# AWS

**Category**: cloud-providers
**Description**: Amazon Web Services cloud platform SDKs and tools
**Homepage**: https://aws.amazon.com

## Package Detection

### NPM
*AWS JavaScript SDK v2 and v3*

- `aws-sdk`
- `@aws-sdk/client-s3`
- `@aws-sdk/client-dynamodb`
- `@aws-sdk/client-lambda`
- `@aws-sdk/client-sqs`
- `@aws-sdk/client-sns`
- `@aws-sdk/client-ec2`
- `@aws-sdk/client-iam`
- `@aws-sdk/client-sts`
- `@aws-sdk/client-secrets-manager`
- `@aws-sdk/client-cloudwatch`

### PYPI
*AWS Python SDK*

- `boto3`
- `botocore`
- `aioboto3`
- `aiobotocore`

### GO
*AWS Go SDK*

- `github.com/aws/aws-sdk-go`
- `github.com/aws/aws-sdk-go-v2`

### MAVEN
*AWS Java SDK*

- `software.amazon.awssdk:s3`
- `software.amazon.awssdk:dynamodb`
- `com.amazonaws:aws-java-sdk`

### RUBYGEMS
*AWS Ruby SDK*

- `aws-sdk`
- `aws-sdk-core`
- `aws-sdk-s3`

## Import Detection

### Javascript

**Pattern**: `from\s+['"]aws-sdk['"]`

**Pattern**: `require\(['"]aws-sdk['"]\)`

**Pattern**: `from\s+['"]@aws-sdk/`

### Python

**Pattern**: `import\s+boto3`

**Pattern**: `from\s+boto3`

**Pattern**: `import\s+botocore`

### Go

**Pattern**: `"github.com/aws/aws-sdk-go`

## Environment Variables


## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Import Detection**: 90% (HIGH)
- **Environment Variable Detection**: 85% (MEDIUM)
