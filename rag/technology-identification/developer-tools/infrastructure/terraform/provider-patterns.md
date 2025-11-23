# Terraform Provider Patterns

## Provider Registry Format

### Official HashiCorp Providers
```
hashicorp/aws
hashicorp/azurerm
hashicorp/google
hashicorp/kubernetes
hashicorp/vault
hashicorp/consul
```

### Partner/Verified Providers
```
datadog/datadog
cloudflare/cloudflare
mongodb/mongodbatlas
pagerduty/pagerduty
```

### Community Providers
```
{organization}/{provider}
```

## Cloud Providers

### AWS (Amazon Web Services)
```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"

  # Authentication methods:
  # 1. AWS credentials file (~/.aws/credentials)
  # 2. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
  # 3. IAM role (EC2, ECS, Lambda)
  # 4. SSO

  profile = "default"

  assume_role {
    role_arn = "arn:aws:iam::123456789012:role/TerraformRole"
  }
}
```

### Azure (Microsoft Azure)
```hcl
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}

  # Authentication methods:
  # 1. Azure CLI
  # 2. Service Principal
  # 3. Managed Identity

  subscription_id = var.subscription_id
  tenant_id       = var.tenant_id
  client_id       = var.client_id
  client_secret   = var.client_secret
}
```

### Google Cloud Platform
```hcl
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = "my-project-id"
  region  = "us-central1"
  zone    = "us-central1-a"

  # Authentication methods:
  # 1. Service account key file
  # 2. Application default credentials
  # 3. gcloud auth

  credentials = file("service-account-key.json")
}
```

### DigitalOcean
```hcl
terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}
```

### Linode
```hcl
terraform {
  required_providers {
    linode = {
      source  = "linode/linode"
      version = "~> 2.0"
    }
  }
}

provider "linode" {
  token = var.linode_token
}
```

### Oracle Cloud Infrastructure
```hcl
terraform {
  required_providers {
    oci = {
      source  = "oracle/oci"
      version = "~> 5.0"
    }
  }
}

provider "oci" {
  region = "us-ashburn-1"
  tenancy_ocid = var.tenancy_ocid
  user_ocid = var.user_ocid
  fingerprint = var.fingerprint
  private_key_path = var.private_key_path
}
```

### Alibaba Cloud
```hcl
terraform {
  required_providers {
    alicloud = {
      source  = "aliyun/alicloud"
      version = "~> 1.0"
    }
  }
}

provider "alicloud" {
  region = "cn-hangzhou"
  access_key = var.access_key
  secret_key = var.secret_key
}
```

## Kubernetes and Container Orchestration

### Kubernetes
```hcl
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
  config_context = "my-context"

  # Or use direct configuration
  host = var.k8s_host
  client_certificate = var.client_cert
  client_key = var.client_key
  cluster_ca_certificate = var.cluster_ca
}
```

### Helm
```hcl
terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
  }
}
```

### Docker
```hcl
terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0"
    }
  }
}

provider "docker" {
  host = "unix:///var/run/docker.sock"
}
```

## Monitoring and Observability

### Datadog
```hcl
terraform {
  required_providers {
    datadog = {
      source  = "DataDog/datadog"
      version = "~> 3.0"
    }
  }
}

provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
  api_url = "https://api.datadoghq.com"
}
```

### New Relic
```hcl
terraform {
  required_providers {
    newrelic = {
      source  = "newrelic/newrelic"
      version = "~> 3.0"
    }
  }
}

provider "newrelic" {
  account_id = var.newrelic_account_id
  api_key = var.newrelic_api_key
  region = "US"
}
```

### PagerDuty
```hcl
terraform {
  required_providers {
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "~> 3.0"
    }
  }
}

provider "pagerduty" {
  token = var.pagerduty_token
}
```

### Grafana
```hcl
terraform {
  required_providers {
    grafana = {
      source  = "grafana/grafana"
      version = "~> 2.0"
    }
  }
}

provider "grafana" {
  url  = "http://grafana.example.com"
  auth = var.grafana_auth
}
```

## DNS and CDN

### Cloudflare
```hcl
terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
  # Or use email + api_key
  email = var.cloudflare_email
  api_key = var.cloudflare_api_key
}
```

### Route53 (via AWS provider)
```hcl
resource "aws_route53_zone" "primary" {
  name = "example.com"
}
```

### NS1
```hcl
terraform {
  required_providers {
    ns1 = {
      source  = "ns1-terraform/ns1"
      version = "~> 2.0"
    }
  }
}

provider "ns1" {
  apikey = var.ns1_api_key
}
```

## Databases

### MongoDB Atlas
```hcl
terraform {
  required_providers {
    mongodbatlas = {
      source  = "mongodb/mongodbatlas"
      version = "~> 1.0"
    }
  }
}

provider "mongodbatlas" {
  public_key = var.mongodb_atlas_public_key
  private_key = var.mongodb_atlas_private_key
}
```

### PostgreSQL
```hcl
terraform {
  required_providers {
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "~> 1.0"
    }
  }
}

provider "postgresql" {
  host = var.db_host
  port = 5432
  database = "postgres"
  username = var.db_username
  password = var.db_password
  sslmode = "require"
}
```

### MySQL
```hcl
terraform {
  required_providers {
    mysql = {
      source  = "petoju/mysql"
      version = "~> 3.0"
    }
  }
}

provider "mysql" {
  endpoint = "${var.db_host}:3306"
  username = var.db_username
  password = var.db_password
}
```

## Version Control and CI/CD

### GitHub
```hcl
terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = "~> 6.0"
    }
  }
}

provider "github" {
  token = var.github_token
  owner = var.github_org
}
```

### GitLab
```hcl
terraform {
  required_providers {
    gitlab = {
      source  = "gitlabhq/gitlab"
      version = "~> 16.0"
    }
  }
}

provider "gitlab" {
  token = var.gitlab_token
  base_url = "https://gitlab.com/api/v4/"
}
```

## Security and Secrets Management

### Vault (HashiCorp)
```hcl
terraform {
  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = "~> 4.0"
    }
  }
}

provider "vault" {
  address = "https://vault.example.com"
  token = var.vault_token

  # Or use other auth methods
  auth_login {
    path = "auth/userpass/login/${var.username}"
    parameters = {
      password = var.password
    }
  }
}
```

### AWS Secrets Manager (via AWS provider)
```hcl
resource "aws_secretsmanager_secret" "example" {
  name = "example-secret"
}
```

## Identity and Access Management

### Okta
```hcl
terraform {
  required_providers {
    okta = {
      source  = "okta/okta"
      version = "~> 4.0"
    }
  }
}

provider "okta" {
  org_name = var.okta_org
  base_url = "okta.com"
  api_token = var.okta_api_token
}
```

### Auth0
```hcl
terraform {
  required_providers {
    auth0 = {
      source  = "auth0/auth0"
      version = "~> 1.0"
    }
  }
}

provider "auth0" {
  domain = var.auth0_domain
  client_id = var.auth0_client_id
  client_secret = var.auth0_client_secret
}
```

## Communication and Collaboration

### Slack
```hcl
terraform {
  required_providers {
    slack = {
      source  = "pablovarela/slack"
      version = "~> 1.0"
    }
  }
}

provider "slack" {
  token = var.slack_token
}
```

## Utility Providers

### Random
```hcl
terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "random" {}

resource "random_password" "db_password" {
  length = 16
  special = true
}
```

### Time
```hcl
terraform {
  required_providers {
    time = {
      source  = "hashicorp/time"
      version = "~> 0.11"
    }
  }
}

provider "time" {}

resource "time_sleep" "wait_30_seconds" {
  create_duration = "30s"
}
```

### Null
```hcl
terraform {
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
  }
}

provider "null" {}

resource "null_resource" "example" {
  provisioner "local-exec" {
    command = "echo 'Hello World'"
  }
}
```

### External
```hcl
terraform {
  required_providers {
    external = {
      source  = "hashicorp/external"
      version = "~> 2.0"
    }
  }
}

data "external" "example" {
  program = ["python", "script.py"]
}
```

### TLS
```hcl
terraform {
  required_providers {
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "tls" {}

resource "tls_private_key" "example" {
  algorithm = "RSA"
  rsa_bits  = 4096
}
```

## Detection Confidence

- **HIGH**: terraform block with required_providers
- **HIGH**: provider block with known provider name
- **HIGH**: Provider-specific resource types (aws_*, azurerm_*, google_*)
- **MEDIUM**: Version constraints for providers
- **MEDIUM**: Provider aliases
- **LOW**: Generic configuration without provider specifics
