# AWS SDK Versions

## JavaScript/Node.js

### AWS SDK v2 (Legacy - Maintenance Mode)
- **Package**: `aws-sdk`
- **Latest**: v2.1689.x
- **Status**: Maintenance mode (security updates only)
- **End of Support**: Expected 2025
- **Node.js**: 10.x and higher
- **Repository**: https://github.com/aws/aws-sdk-js

### AWS SDK v3 (Current)
- **Package**: `@aws-sdk/*`
- **Latest**: v3.x (modular packages)
- **Status**: Active development
- **Node.js**: 14.x and higher (18.x recommended)
- **Repository**: https://github.com/aws/aws-sdk-js-v3

#### Key v3 Packages
```json
"@aws-sdk/client-s3": "^3.x"
"@aws-sdk/client-dynamodb": "^3.x"
"@aws-sdk/client-lambda": "^3.x"
"@aws-sdk/client-sqs": "^3.x"
"@aws-sdk/client-sns": "^3.x"
"@aws-sdk/lib-dynamodb": "^3.x"
```

## Python

### boto3
- **Package**: `boto3`
- **Latest**: 1.x (1.35.x as of 2025)
- **Status**: Active development
- **Python**: 3.8+ (3.12 recommended)
- **Repository**: https://github.com/boto/boto3

### botocore
- **Package**: `botocore`
- **Latest**: 1.x (dependency of boto3)
- **Status**: Active development
- **Python**: 3.8+
- **Repository**: https://github.com/boto/botocore

### Legacy boto (EOL)
- **Package**: `boto`
- **Status**: End of Life (use boto3)
- **Python**: 2.6, 2.7, 3.3+

## Java

### AWS SDK for Java v1 (Maintenance Mode)
- **Package**: `com.amazonaws:aws-java-sdk`
- **Latest**: 1.12.x
- **Status**: Maintenance mode
- **Java**: 8+
- **Repository**: https://github.com/aws/aws-sdk-java

### AWS SDK for Java v2 (Current)
- **Package**: `software.amazon.awssdk:*`
- **Latest**: 2.x (2.28.x as of 2025)
- **Status**: Active development
- **Java**: 8+ (11 recommended)
- **Repository**: https://github.com/aws/aws-sdk-java-v2

#### pom.xml Example
```xml
<dependency>
    <groupId>software.amazon.awssdk</groupId>
    <artifactId>s3</artifactId>
    <version>2.28.0</version>
</dependency>
```

## .NET

### AWS SDK for .NET v3 (Current)
- **Package**: `AWSSDK.*`
- **Latest**: 3.7.x
- **Status**: Active development
- **.NET**: .NET Core 3.1+, .NET 5+, .NET Framework 4.6.1+
- **Repository**: https://github.com/aws/aws-sdk-net

#### Common Packages
```xml
<PackageReference Include="AWSSDK.S3" Version="3.7.*" />
<PackageReference Include="AWSSDK.DynamoDBv2" Version="3.7.*" />
<PackageReference Include="AWSSDK.Lambda" Version="3.7.*" />
```

## Go

### AWS SDK for Go v1
- **Package**: `github.com/aws/aws-sdk-go`
- **Latest**: v1.55.x
- **Status**: Maintenance mode
- **Go**: 1.15+
- **Repository**: https://github.com/aws/aws-sdk-go

### AWS SDK for Go v2 (Current)
- **Package**: `github.com/aws/aws-sdk-go-v2`
- **Latest**: v1.x (modular)
- **Status**: Active development
- **Go**: 1.20+ (1.21 recommended)
- **Repository**: https://github.com/aws/aws-sdk-go-v2

#### go.mod Example
```go
require (
    github.com/aws/aws-sdk-go-v2 v1.30.0
    github.com/aws/aws-sdk-go-v2/service/s3 v1.60.0
    github.com/aws/aws-sdk-go-v2/service/dynamodb v1.34.0
)
```

## Ruby

### AWS SDK for Ruby v3 (Current)
- **Package**: `aws-sdk-*`
- **Latest**: v3.x (3.2.x as of 2025)
- **Status**: Active development
- **Ruby**: 2.5+ (3.x recommended)
- **Repository**: https://github.com/aws/aws-sdk-ruby

#### Gemfile Example
```ruby
gem 'aws-sdk-s3', '~> 1.160'
gem 'aws-sdk-dynamodb', '~> 1.120'
gem 'aws-sdk-lambda', '~> 1.130'
```

### AWS SDK for Ruby v2 (Legacy)
- **Package**: `aws-sdk`
- **Status**: Deprecated (use v3)
- **Ruby**: 1.9.3+

## PHP

### AWS SDK for PHP v3 (Current)
- **Package**: `aws/aws-sdk-php`
- **Latest**: v3.x (3.320.x as of 2025)
- **Status**: Active development
- **PHP**: 7.2.5+ (8.x recommended)
- **Repository**: https://github.com/aws/aws-sdk-php

#### composer.json Example
```json
{
  "require": {
    "aws/aws-sdk-php": "^3.320"
  }
}
```

### AWS SDK for PHP v2 (EOL)
- **Status**: End of Life (use v3)
- **PHP**: 5.3.3+

## Rust

### AWS SDK for Rust (Developer Preview → GA)
- **Package**: `aws-sdk-*`
- **Latest**: 1.x (services) / 0.x (core)
- **Status**: Generally Available (as of 2024)
- **Rust**: 1.70+
- **Repository**: https://github.com/awslabs/aws-sdk-rust

#### Cargo.toml Example
```toml
[dependencies]
aws-config = "1.5"
aws-sdk-s3 = "1.50"
aws-sdk-dynamodb = "1.40"
```

### Rusoto (Community - Maintenance)
- **Package**: `rusoto_*`
- **Status**: Community maintained
- **Note**: AWS official SDK now available

## C++

### AWS SDK for C++
- **Package**: `aws-sdk-cpp`
- **Latest**: 1.11.x
- **Status**: Active development
- **C++**: C++11 or higher
- **Repository**: https://github.com/aws/aws-sdk-cpp

## Swift

### AWS SDK for Swift
- **Package**: `aws-sdk-swift`
- **Latest**: 1.x (Developer Preview)
- **Status**: Developer Preview
- **Swift**: 5.7+
- **Repository**: https://github.com/awslabs/aws-sdk-swift

## Kotlin

### AWS SDK for Kotlin
- **Package**: `aws.sdk.kotlin:*`
- **Latest**: 1.x
- **Status**: Active development
- **Kotlin**: 1.9+
- **Repository**: https://github.com/awslabs/aws-sdk-kotlin

## CLI and Tools

### AWS CLI v2 (Current)
- **Latest**: 2.x (2.19.x as of 2025)
- **Status**: Active development
- **Python**: Bundled (no Python required)
- **Repository**: https://github.com/aws/aws-cli

### AWS CLI v1 (Maintenance)
- **Latest**: 1.x
- **Status**: Maintenance mode
- **Python**: 3.8+

### AWS SAM CLI
- **Latest**: 1.x (1.127.x as of 2025)
- **Status**: Active development
- **Python**: 3.8+
- **Repository**: https://github.com/aws/aws-sam-cli

### AWS CDK
- **Latest**: v2.x (2.160.x as of 2025)
- **Status**: Active development
- **Node.js**: 14.x+
- **Repository**: https://github.com/aws/aws-cdk

## Version Detection Patterns

### Package Manager Files

#### package.json
```json
"aws-sdk": "^2.1689.0"
"@aws-sdk/client-s3": "^3.650.0"
```

#### requirements.txt
```
boto3==1.35.0
boto3>=1.35.0
botocore==1.35.0
```

#### Gemfile.lock
```
aws-sdk-s3 (1.160.0)
```

#### composer.lock
```json
"aws/aws-sdk-php": "3.320.0"
```

#### go.mod
```
github.com/aws/aws-sdk-go-v2 v1.30.0
```

## Migration Paths

### Node.js: v2 → v3
- Breaking changes: Modular packages, async/await first
- Benefits: Tree-shaking, smaller bundle size, first-class TypeScript

### Python: boto → boto3
- Complete rewrite with resource-oriented interface
- Not backwards compatible

### Java: v1 → v2
- Breaking changes: Package names, async-first API
- Benefits: Non-blocking I/O, better performance

### Go: v1 → v2
- Breaking changes: Context-based API, modular design
- Benefits: Better error handling, middleware support

## Detection Confidence

- **HIGH**: Exact version in lock files
- **HIGH**: Version in package manager manifest
- **MEDIUM**: Version range/constraint
- **LOW**: No version specified (latest assumed)
