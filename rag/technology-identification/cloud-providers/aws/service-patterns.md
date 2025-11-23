# AWS Service Patterns

## S3 (Simple Storage Service)

### Service Identifiers
- Service Name: `s3`
- Client Class: `S3`, `S3Client`, `AmazonS3`
- Resource: `s3.amazonaws.com`

### Common Operations
```
putObject, getObject, deleteObject, copyObject
upload, download
listBuckets, listObjects, listObjectsV2
createBucket, deleteBucket
getBucketLocation, headBucket
putBucketPolicy, getBucketPolicy
```

### URL Patterns
- `https://*.s3.amazonaws.com/*`
- `https://s3.*.amazonaws.com/*`
- `https://*.s3-*.amazonaws.com/*`
- `s3://bucket-name/key`

### ARN Pattern
- `arn:aws:s3:::bucket-name`
- `arn:aws:s3:::bucket-name/*`

## DynamoDB

### Service Identifiers
- Service Name: `dynamodb`
- Client Class: `DynamoDB`, `DynamoDBClient`, `AmazonDynamoDB`
- Resource: `dynamodb.*.amazonaws.com`

### Common Operations
```
putItem, getItem, deleteItem, updateItem
query, scan
batchGetItem, batchWriteItem
createTable, deleteTable, updateTable, describeTable
transactWriteItems, transactGetItems
```

### URL Patterns
- `https://dynamodb.*.amazonaws.com`

### ARN Pattern
- `arn:aws:dynamodb:*:*:table/*`

## Lambda

### Service Identifiers
- Service Name: `lambda`
- Client Class: `Lambda`, `LambdaClient`, `AWSLambda`
- Resource: `lambda.*.amazonaws.com`

### Common Operations
```
invoke, invokeAsync
createFunction, updateFunctionCode, updateFunctionConfiguration
deleteFunction, getFunction
listFunctions, listVersionsByFunction
publishVersion, createAlias
addPermission, removePermission
```

### Handler Patterns
```javascript
exports.handler = async (event, context) => {}
module.exports.handler = async (event) => {}
```

```python
def lambda_handler(event, context):
    pass
```

### Runtime Identifiers
- `nodejs18.x`, `nodejs20.x`
- `python3.9`, `python3.10`, `python3.11`, `python3.12`
- `java11`, `java17`, `java21`
- `dotnet6`, `dotnet8`
- `go1.x`
- `ruby3.2`, `ruby3.3`
- `provided.al2`, `provided.al2023`

### ARN Pattern
- `arn:aws:lambda:*:*:function:*`

## SQS (Simple Queue Service)

### Service Identifiers
- Service Name: `sqs`
- Client Class: `SQS`, `SQSClient`, `AmazonSQS`
- Resource: `sqs.*.amazonaws.com`

### Common Operations
```
sendMessage, sendMessageBatch
receiveMessage, deleteMessage, deleteMessageBatch
createQueue, deleteQueue, listQueues
getQueueUrl, getQueueAttributes, setQueueAttributes
purgeQueue, changeMessageVisibility
```

### URL Patterns
- `https://sqs.*.amazonaws.com/*`
- `https://*.queue.amazonaws.com/*`

### ARN Pattern
- `arn:aws:sqs:*:*:*`

## SNS (Simple Notification Service)

### Service Identifiers
- Service Name: `sns`
- Client Class: `SNS`, `SNSClient`, `AmazonSNS`
- Resource: `sns.*.amazonaws.com`

### Common Operations
```
publish
subscribe, unsubscribe, confirmSubscription
createTopic, deleteTopic, listTopics
getTopicAttributes, setTopicAttributes
addPermission, removePermission
```

### ARN Pattern
- `arn:aws:sns:*:*:*`

## EC2 (Elastic Compute Cloud)

### Service Identifiers
- Service Name: `ec2`
- Client Class: `EC2`, `EC2Client`, `AmazonEC2`
- Resource: `ec2.*.amazonaws.com`

### Common Operations
```
runInstances, terminateInstances, stopInstances, startInstances
describeInstances, describeImages, describeSecurityGroups
createSecurityGroup, authorizeSecurityGroupIngress
createKeyPair, deleteKeyPair
allocateAddress, associateAddress
```

### Instance Patterns
- Instance IDs: `i-*`
- AMI IDs: `ami-*`
- Security Group IDs: `sg-*`

### ARN Pattern
- `arn:aws:ec2:*:*:instance/*`
- `arn:aws:ec2:*:*:security-group/*`

## RDS (Relational Database Service)

### Service Identifiers
- Service Name: `rds`
- Client Class: `RDS`, `RDSClient`, `AmazonRDS`
- Resource: `rds.amazonaws.com`

### Common Operations
```
createDBInstance, deleteDBInstance, modifyDBInstance
describeDBInstances, describeDBClusters
createDBSnapshot, deleteDBSnapshot
restoreDBInstanceFromSnapshot
```

### Endpoint Patterns
- `*.*.rds.amazonaws.com`
- `*-cluster.*.rds.amazonaws.com`

### ARN Pattern
- `arn:aws:rds:*:*:db:*`
- `arn:aws:rds:*:*:cluster:*`

## ECS (Elastic Container Service)

### Service Identifiers
- Service Name: `ecs`
- Client Class: `ECS`, `ECSClient`, `AmazonECS`
- Resource: `ecs.*.amazonaws.com`

### Common Operations
```
createCluster, deleteCluster, describeClusters
createService, updateService, deleteService
runTask, stopTask, describeTasks
registerTaskDefinition, deregisterTaskDefinition
```

### ARN Pattern
- `arn:aws:ecs:*:*:cluster/*`
- `arn:aws:ecs:*:*:service/*`
- `arn:aws:ecs:*:*:task/*`

## CloudWatch

### Service Identifiers
- Service Name: `cloudwatch`, `logs`
- Client Class: `CloudWatch`, `CloudWatchLogs`
- Resource: `monitoring.*.amazonaws.com`, `logs.*.amazonaws.com`

### Common Operations
```
putMetricData, getMetricStatistics
putMetricAlarm, describeAlarms
createLogGroup, createLogStream
putLogEvents, getLogEvents
```

### Log Group Patterns
- `/aws/lambda/*`
- `/aws/ecs/*`
- `/aws/rds/*`

### ARN Pattern
- `arn:aws:logs:*:*:log-group:*`

## API Gateway

### Service Identifiers
- Service Name: `apigateway`, `apigatewayv2`
- Client Class: `APIGateway`, `ApiGatewayV2`
- Resource: `*.execute-api.*.amazonaws.com`

### Common Operations
```
createRestApi, deleteRestApi
createResource, createMethod
createDeployment, createStage
putIntegration, putIntegrationResponse
```

### URL Patterns
- `https://*.execute-api.*.amazonaws.com/*`

### ARN Pattern
- `arn:aws:apigateway:*::/restapis/*`

## CloudFormation

### Service Identifiers
- Service Name: `cloudformation`
- Client Class: `CloudFormation`, `CloudFormationClient`

### Common Operations
```
createStack, updateStack, deleteStack
describeStacks, describeStackResources
validateTemplate, estimateTemplateCost
```

### Template Patterns
```yaml
AWSTemplateFormatVersion: '2010-09-09'
Resources:
Outputs:
Parameters:
```

### ARN Pattern
- `arn:aws:cloudformation:*:*:stack/*/*`

## IAM (Identity and Access Management)

### Service Identifiers
- Service Name: `iam`
- Client Class: `IAM`, `IAMClient`, `AmazonIdentityManagement`

### Common Operations
```
createUser, deleteUser, listUsers
createRole, deleteRole, attachRolePolicy
createPolicy, deletePolicy
getUser, getRole, getPolicy
```

### ARN Pattern
- `arn:aws:iam::*:user/*`
- `arn:aws:iam::*:role/*`
- `arn:aws:iam::*:policy/*`

## Secrets Manager

### Service Identifiers
- Service Name: `secretsmanager`
- Client Class: `SecretsManager`, `SecretsManagerClient`

### Common Operations
```
createSecret, deleteSecret
getSecretValue, putSecretValue
updateSecret, rotateSecret
listSecrets, describeSecret
```

### ARN Pattern
- `arn:aws:secretsmanager:*:*:secret:*`

## KMS (Key Management Service)

### Service Identifiers
- Service Name: `kms`
- Client Class: `KMS`, `KMSClient`, `AWSKMS`

### Common Operations
```
createKey, deleteKey
encrypt, decrypt
generateDataKey, generateDataKeyWithoutPlaintext
describeKey, listKeys
```

### Key Patterns
- Key IDs: UUID format
- Aliases: `alias/*`

### ARN Pattern
- `arn:aws:kms:*:*:key/*`
- `arn:aws:kms:*:*:alias/*`

## Detection Confidence

- **HIGH**: Service-specific SDK client initialization
- **HIGH**: Service-specific API method calls
- **HIGH**: Service ARN patterns in IAM policies
- **MEDIUM**: Service endpoint URLs
- **MEDIUM**: CloudFormation resource types
- **LOW**: Generic AWS patterns without service specifics
