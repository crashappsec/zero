# AWS SDK Patterns

## Package Names

### JavaScript/Node.js
- `aws-sdk` (AWS SDK v2 - legacy)
- `@aws-sdk/client-*` (AWS SDK v3 modular packages)
- `@aws-sdk/lib-dynamodb`
- `@aws-sdk/smithy-client`
- `@aws-sdk/credential-providers`

### Python (boto3/botocore)
- `boto3` (official AWS SDK)
- `botocore` (low-level interface)
- `aioboto3` (async boto3)
- `aiobotocore` (async botocore)

### Java
- `com.amazonaws:aws-java-sdk`
- `com.amazonaws:aws-java-sdk-*` (modular)
- `software.amazon.awssdk:*` (AWS SDK v2)

### .NET
- `AWSSDK.Core`
- `AWSSDK.*` (service-specific packages)
- `Amazon.Lambda.*`

### Go
- `github.com/aws/aws-sdk-go`
- `github.com/aws/aws-sdk-go-v2`
- `github.com/aws/aws-sdk-go-v2/service/*`

### Ruby
- `aws-sdk` (full SDK)
- `aws-sdk-*` (service-specific gems)
- `aws-sdk-core`

### PHP
- `aws/aws-sdk-php`

### Rust
- `aws-sdk-*`
- `rusoto_*`

### C++
- `aws-sdk-cpp`

## Import Patterns

### JavaScript/Node.js (SDK v2)
```javascript
const AWS = require('aws-sdk');
import AWS from 'aws-sdk';
import * as AWS from 'aws-sdk';
```

### JavaScript/Node.js (SDK v3)
```javascript
import { S3Client } from "@aws-sdk/client-s3";
import { DynamoDBClient } from "@aws-sdk/client-dynamodb";
import { LambdaClient } from "@aws-sdk/client-lambda";
import { SQSClient } from "@aws-sdk/client-sqs";
import { SNSClient } from "@aws-sdk/client-sns";
```

### Python
```python
import boto3
import botocore
from boto3 import client, resource, session
from botocore.exceptions import ClientError
import aioboto3
```

### Java (SDK v1)
```java
import com.amazonaws.services.s3.AmazonS3;
import com.amazonaws.services.dynamodbv2.AmazonDynamoDB;
import com.amazonaws.services.lambda.AWSLambda;
import com.amazonaws.auth.AWSCredentials;
```

### Java (SDK v2)
```java
import software.amazon.awssdk.services.s3.S3Client;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.lambda.LambdaClient;
import software.amazon.awssdk.auth.credentials.*;
```

### .NET/C#
```csharp
using Amazon;
using Amazon.S3;
using Amazon.DynamoDBv2;
using Amazon.Lambda;
using Amazon.Runtime;
```

### Go (SDK v1)
```go
import "github.com/aws/aws-sdk-go/aws"
import "github.com/aws/aws-sdk-go/service/s3"
import "github.com/aws/aws-sdk-go/service/dynamodb"
```

### Go (SDK v2)
```go
import "github.com/aws/aws-sdk-go-v2/aws"
import "github.com/aws/aws-sdk-go-v2/service/s3"
import "github.com/aws/aws-sdk-go-v2/service/dynamodb"
```

### Ruby
```ruby
require 'aws-sdk'
require 'aws-sdk-s3'
require 'aws-sdk-dynamodb'
require 'aws-sdk-lambda'
```

### PHP
```php
require 'vendor/autoload.php';
use Aws\S3\S3Client;
use Aws\DynamoDb\DynamoDbClient;
use Aws\Lambda\LambdaClient;
```

## Client Initialization Patterns

### JavaScript (SDK v2)
```javascript
const s3 = new AWS.S3();
const dynamodb = new AWS.DynamoDB();
const lambda = new AWS.Lambda();
AWS.config.update({ region: 'us-east-1' });
```

### JavaScript (SDK v3)
```javascript
const s3Client = new S3Client({ region: "us-east-1" });
const dynamoClient = new DynamoDBClient({ region: "us-east-1" });
const lambdaClient = new LambdaClient({ region: "us-east-1" });
```

### Python
```python
s3 = boto3.client('s3')
s3 = boto3.resource('s3')
dynamodb = boto3.client('dynamodb')
dynamodb = boto3.resource('dynamodb')
lambda_client = boto3.client('lambda')
```

### Java
```java
AmazonS3 s3Client = AmazonS3ClientBuilder.standard().build();
AmazonDynamoDB dynamoClient = AmazonDynamoDBClientBuilder.standard().build();
AWSLambda lambdaClient = AWSLambdaClientBuilder.standard().build();
```

### .NET
```csharp
var s3Client = new AmazonS3Client();
var dynamoClient = new AmazonDynamoDBClient();
var lambdaClient = new AmazonLambdaClient();
```

### Go
```go
sess := session.Must(session.NewSession())
s3Svc := s3.New(sess)
dynamoSvc := dynamodb.New(sess)
```

### Ruby
```ruby
s3 = Aws::S3::Client.new
dynamodb = Aws::DynamoDB::Client.new
lambda = Aws::Lambda::Client.new
```

## Configuration Patterns

### Region Configuration
```
us-east-1, us-west-2, eu-west-1, ap-southeast-1
AWS_REGION, AWS_DEFAULT_REGION
region: 'us-east-1'
```

### Credential Configuration
```javascript
credentials: new AWS.Credentials(accessKeyId, secretAccessKey)
credentials: new AWS.SharedIniFileCredentials({profile: 'default'})
credentials: new AWS.EnvironmentCredentials('AWS')
```

### Endpoint Configuration
```
endpoint: 'https://s3.amazonaws.com'
endpoint: 'http://localhost:4566' (LocalStack)
```

## Common Method Patterns

### S3
```
putObject, getObject, deleteObject
listBuckets, listObjects, listObjectsV2
createBucket, deleteBucket
upload, download
```

### DynamoDB
```
putItem, getItem, deleteItem, updateItem
query, scan, batchGetItem, batchWriteItem
createTable, deleteTable, describeTable
```

### Lambda
```
invoke, invokeAsync
createFunction, updateFunction, deleteFunction
listFunctions
```

### SQS
```
sendMessage, receiveMessage, deleteMessage
createQueue, deleteQueue, listQueues
getQueueUrl, getQueueAttributes
```

### SNS
```
publish, subscribe, unsubscribe
createTopic, deleteTopic, listTopics
```

## AWS CLI Patterns

### Commands
```bash
aws s3 ls
aws s3 cp
aws dynamodb scan
aws lambda invoke
aws sqs send-message
aws sns publish
aws ec2 describe-instances
aws cloudformation deploy
```

### Configuration Files
```
~/.aws/config
~/.aws/credentials
```

## Infrastructure as Code

### CloudFormation
```yaml
Resources:
  MyBucket:
    Type: AWS::S3::Bucket
  MyTable:
    Type: AWS::DynamoDB::Table
  MyFunction:
    Type: AWS::Lambda::Function
```

### SAM Template
```yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31
Resources:
  MyFunction:
    Type: AWS::Serverless::Function
```

### CDK Patterns
```javascript
import * as cdk from 'aws-cdk-lib';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as lambda from 'aws-cdk-lib/aws-lambda';
```

## Detection Confidence

- **HIGH**: Direct AWS SDK imports and usage
- **HIGH**: boto3/botocore for Python
- **HIGH**: AWS CLI commands in scripts
- **MEDIUM**: CloudFormation/SAM templates
- **MEDIUM**: CDK infrastructure code
- **LOW**: Generic cloud patterns without AWS markers
