# AWS Endpoint Patterns

## Service Endpoint Formats

### Standard Regional Endpoints
```
https://{service}.{region}.amazonaws.com
https://{service}.{region}.amazonaws.com.cn (China)
https://{service}.{region}.c2s.ic.gov (C2S)
https://{service}.{region}.sc2s.sgov.gov (SC2S)
```

### S3 Endpoint Variations
```
https://s3.{region}.amazonaws.com/{bucket}/{key}
https://{bucket}.s3.{region}.amazonaws.com/{key}
https://s3.amazonaws.com/{bucket}/{key} (legacy)
https://{bucket}.s3.amazonaws.com/{key} (legacy)
https://s3-{region}.amazonaws.com/{bucket}/{key} (legacy)
s3://{bucket}/{key} (CLI/SDK format)
```

### DynamoDB Endpoints
```
https://dynamodb.{region}.amazonaws.com
https://dynamodb.{region}.amazonaws.com.cn (China)
http://localhost:8000 (DynamoDB Local)
```

### Lambda Endpoints
```
https://lambda.{region}.amazonaws.com
```

### API Gateway Endpoints
```
https://{api-id}.execute-api.{region}.amazonaws.com/{stage}
https://{custom-domain}/
```

### CloudFront Distribution URLs
```
https://{distribution-id}.cloudfront.net
https://{custom-domain} (with custom domain)
```

### ECS Task URLs
```
https://{task-id}.{cluster}.{region}.ecs.amazonaws.com
```

### ELB/ALB Endpoints
```
https://{load-balancer}.{region}.elb.amazonaws.com
https://{load-balancer}.elb.{region}.amazonaws.com
```

### RDS Endpoints
```
{instance-name}.{random-id}.{region}.rds.amazonaws.com
{cluster-name}-cluster.{random-id}.{region}.rds.amazonaws.com
```

### ElastiCache Endpoints
```
{cache-id}.{random-id}.{az-id}.cache.amazonaws.com
{cache-id}-cluster.{random-id}.{region}.cache.amazonaws.com
```

### SQS Queue URLs
```
https://sqs.{region}.amazonaws.com/{account-id}/{queue-name}
https://{region}.queue.amazonaws.com/{account-id}/{queue-name}
```

### SNS Endpoints
```
https://sns.{region}.amazonaws.com
```

### EC2 Endpoints
```
https://ec2.{region}.amazonaws.com
```

### CloudWatch Endpoints
```
https://monitoring.{region}.amazonaws.com
https://logs.{region}.amazonaws.com
https://events.{region}.amazonaws.com
```

### Secrets Manager Endpoints
```
https://secretsmanager.{region}.amazonaws.com
```

### KMS Endpoints
```
https://kms.{region}.amazonaws.com
```

### STS (Security Token Service) Endpoints
```
https://sts.{region}.amazonaws.com
https://sts.amazonaws.com (global)
```

### IAM Endpoints
```
https://iam.amazonaws.com (global service)
```

### Route53 Endpoints
```
https://route53.amazonaws.com (global service)
```

### CloudFront Endpoints
```
https://cloudfront.amazonaws.com (global service)
```

### S3 Transfer Acceleration
```
https://{bucket}.s3-accelerate.amazonaws.com
https://{bucket}.s3-accelerate.dualstack.amazonaws.com
```

### VPC Endpoints (PrivateLink)
```
https://{vpce-id}.{service}.{region}.vpce.amazonaws.com
```

### AWS Regions

#### US East
- `us-east-1` (N. Virginia)
- `us-east-2` (Ohio)

#### US West
- `us-west-1` (N. California)
- `us-west-2` (Oregon)

#### Africa
- `af-south-1` (Cape Town)

#### Asia Pacific
- `ap-east-1` (Hong Kong)
- `ap-south-1` (Mumbai)
- `ap-south-2` (Hyderabad)
- `ap-northeast-1` (Tokyo)
- `ap-northeast-2` (Seoul)
- `ap-northeast-3` (Osaka)
- `ap-southeast-1` (Singapore)
- `ap-southeast-2` (Sydney)
- `ap-southeast-3` (Jakarta)
- `ap-southeast-4` (Melbourne)

#### Canada
- `ca-central-1` (Central)
- `ca-west-1` (Calgary)

#### Europe
- `eu-central-1` (Frankfurt)
- `eu-central-2` (Zurich)
- `eu-west-1` (Ireland)
- `eu-west-2` (London)
- `eu-west-3` (Paris)
- `eu-south-1` (Milan)
- `eu-south-2` (Spain)
- `eu-north-1` (Stockholm)

#### Middle East
- `me-south-1` (Bahrain)
- `me-central-1` (UAE)

#### South America
- `sa-east-1` (Sao Paulo)

#### AWS GovCloud (US)
- `us-gov-east-1`
- `us-gov-west-1`

#### China
- `cn-north-1` (Beijing)
- `cn-northwest-1` (Ningxia)

## Local Development Endpoints

### LocalStack
```
http://localhost:4566 (default)
http://localstack:4566 (Docker)
```

### DynamoDB Local
```
http://localhost:8000
```

### S3 Compatible (MinIO)
```
http://localhost:9000
http://minio:9000 (Docker)
```

### SAM Local
```
http://localhost:3000 (API)
http://localhost:3001 (Lambda)
```

### ElasticMQ (SQS)
```
http://localhost:9324
```

## VPC Endpoint Patterns

### Interface Endpoints
```
{vpce-id}.{service}.{region}.vpce.amazonaws.com
*.{vpce-id}.{service}.{region}.vpce.amazonaws.com
```

### Gateway Endpoints
- S3: Routes through VPC routing table
- DynamoDB: Routes through VPC routing table

## Custom Domain Patterns

### API Gateway Custom Domains
```
api.example.com
*.example.com
```

### CloudFront Custom Domains
```
cdn.example.com
www.example.com
```

## SDK Endpoint Configuration

### JavaScript
```javascript
const s3 = new AWS.S3({
  endpoint: 'https://s3.us-west-2.amazonaws.com',
  region: 'us-west-2'
});
```

### Python
```python
s3 = boto3.client('s3',
  endpoint_url='https://s3.us-west-2.amazonaws.com',
  region_name='us-west-2'
)
```

### Java
```java
AmazonS3 s3 = AmazonS3ClientBuilder.standard()
  .withEndpointConfiguration(
    new EndpointConfiguration("https://s3.us-west-2.amazonaws.com", "us-west-2")
  )
  .build();
```

### Go
```go
sess := session.Must(session.NewSession(&aws.Config{
  Endpoint: aws.String("https://s3.us-west-2.amazonaws.com"),
  Region:   aws.String("us-west-2"),
}))
```

## Environment Variables

```bash
AWS_ENDPOINT_URL
AWS_ENDPOINT_URL_S3
AWS_ENDPOINT_URL_DYNAMODB
AWS_ENDPOINT_URL_LAMBDA
```

## Endpoint Detection in Configuration

### aws-cli config
```ini
[profile dev]
endpoint_url = https://localhost:4566
```

### Terraform
```hcl
provider "aws" {
  endpoints {
    s3       = "http://localhost:4566"
    dynamodb = "http://localhost:4566"
  }
}
```

### Docker Compose
```yaml
environment:
  - AWS_ENDPOINT_URL=http://localstack:4566
  - AWS_ENDPOINT_URL_S3=http://localstack:4566
```

## Detection Confidence

- **HIGH**: Standard AWS service endpoints with region
- **HIGH**: S3 bucket URLs with amazonaws.com domain
- **HIGH**: API Gateway execute-api endpoints
- **MEDIUM**: Custom domains (requires additional verification)
- **MEDIUM**: Local development endpoints (localhost, localstack)
- **LOW**: Generic HTTPS endpoints without AWS markers
