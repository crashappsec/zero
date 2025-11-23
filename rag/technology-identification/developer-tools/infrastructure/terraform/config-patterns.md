# Terraform Configuration Patterns

## File Names

### Standard Terraform Files
- `*.tf` - Terraform configuration files
- `*.tfvars` - Variable values
- `terraform.tfvars` - Default variable values
- `*.auto.tfvars` - Auto-loaded variable files
- `*.tf.json` - JSON configuration format

### Common File Naming Conventions
- `main.tf` - Main configuration
- `variables.tf` - Variable declarations
- `outputs.tf` - Output declarations
- `providers.tf` - Provider configurations
- `versions.tf` - Version constraints
- `terraform.tf` - Terraform block
- `backend.tf` - Backend configuration
- `data.tf` - Data sources
- `locals.tf` - Local values

### State Files (Should be in .gitignore)
- `terraform.tfstate`
- `terraform.tfstate.backup`
- `.terraform/` - Directory with providers and modules
- `.terraform.lock.hcl` - Dependency lock file

## Terraform Block

### Version Constraints
```hcl
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}
```

### Backend Configuration
```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "prod/terraform.tfstate"
    region = "us-east-1"
  }
}

terraform {
  backend "azurerm" {
    resource_group_name  = "terraform-rg"
    storage_account_name = "terraformstate"
    container_name       = "tfstate"
    key                  = "prod.terraform.tfstate"
  }
}

terraform {
  backend "gcs" {
    bucket = "tf-state-bucket"
    prefix = "terraform/state"
  }
}

terraform {
  backend "remote" {
    organization = "my-org"
    workspaces {
      name = "my-workspace"
    }
  }
}
```

### Cloud Block (Terraform Cloud)
```hcl
terraform {
  cloud {
    organization = "my-org"
    workspaces {
      name = "my-workspace"
    }
  }
}
```

## Provider Configuration

### AWS Provider
```hcl
provider "aws" {
  region = "us-east-1"
  profile = "default"

  default_tags {
    tags = {
      Environment = "production"
      ManagedBy   = "Terraform"
    }
  }
}
```

### Azure Provider
```hcl
provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
  tenant_id       = var.tenant_id
}
```

### Google Cloud Provider
```hcl
provider "google" {
  project = "my-project"
  region  = "us-central1"
  zone    = "us-central1-a"
}
```

### Multiple Provider Instances
```hcl
provider "aws" {
  alias  = "west"
  region = "us-west-2"
}

provider "aws" {
  alias  = "east"
  region = "us-east-1"
}
```

## Resource Blocks

### Basic Resource
```hcl
resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  tags = {
    Name = "web-server"
  }
}
```

### Resource with Dependencies
```hcl
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "subnet" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"

  depends_on = [aws_vpc.main]
}
```

### Resource with Count
```hcl
resource "aws_instance" "server" {
  count         = 3
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  tags = {
    Name = "server-${count.index}"
  }
}
```

### Resource with For_Each
```hcl
resource "aws_instance" "server" {
  for_each = toset(["web", "api", "db"])

  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  tags = {
    Name = each.key
  }
}
```

### Resource Lifecycle
```hcl
resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = true
    ignore_changes        = [tags]
  }
}
```

## Data Sources

### AWS Data Sources
```hcl
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}

data "aws_vpc" "default" {
  default = true
}

data "aws_availability_zones" "available" {
  state = "available"
}
```

### Azure Data Sources
```hcl
data "azurerm_resource_group" "example" {
  name = "existing-rg"
}

data "azurerm_virtual_network" "example" {
  name                = "production"
  resource_group_name = data.azurerm_resource_group.example.name
}
```

## Variables

### Variable Declaration
```hcl
variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t2.micro"
}

variable "instance_count" {
  description = "Number of instances"
  type        = number
  default     = 1
}

variable "enable_monitoring" {
  description = "Enable detailed monitoring"
  type        = bool
  default     = false
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}

variable "instance_config" {
  description = "Instance configuration"
  type = object({
    ami           = string
    instance_type = string
    key_name      = string
  })
}
```

### Variable Validation
```hcl
variable "instance_type" {
  type = string

  validation {
    condition     = contains(["t2.micro", "t2.small", "t2.medium"], var.instance_type)
    error_message = "Instance type must be t2.micro, t2.small, or t2.medium"
  }
}
```

### Sensitive Variables
```hcl
variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}
```

## Outputs

### Basic Outputs
```hcl
output "instance_id" {
  description = "ID of the EC2 instance"
  value       = aws_instance.web.id
}

output "instance_public_ip" {
  description = "Public IP of the instance"
  value       = aws_instance.web.public_ip
  sensitive   = true
}

output "instance_ids" {
  description = "IDs of all instances"
  value       = aws_instance.server[*].id
}
```

## Local Values

```hcl
locals {
  common_tags = {
    Environment = var.environment
    Project     = var.project_name
    ManagedBy   = "Terraform"
  }

  instance_name = "${var.project_name}-${var.environment}-instance"

  availability_zones = slice(data.aws_availability_zones.available.names, 0, 3)
}

resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
  tags          = local.common_tags
}
```

## Modules

### Module Usage
```hcl
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"

  name = "my-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-east-1a", "us-east-1b", "us-east-1c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  enable_vpn_gateway = true

  tags = {
    Terraform   = "true"
    Environment = "dev"
  }
}
```

### Local Module
```hcl
module "servers" {
  source = "./modules/ec2-cluster"

  cluster_name  = "web-cluster"
  instance_type = "t2.micro"
  instance_count = 3
}
```

### Module Output Reference
```hcl
output "vpc_id" {
  value = module.vpc.vpc_id
}
```

## Dynamic Blocks

```hcl
resource "aws_security_group" "example" {
  name = "example"

  dynamic "ingress" {
    for_each = var.ingress_rules
    content {
      from_port   = ingress.value.from_port
      to_port     = ingress.value.to_port
      protocol    = ingress.value.protocol
      cidr_blocks = ingress.value.cidr_blocks
    }
  }
}
```

## Conditional Expressions

```hcl
resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = var.environment == "production" ? "t2.large" : "t2.micro"

  count = var.create_instance ? 1 : 0
}
```

## Functions

### Common Functions
```hcl
# String functions
locals {
  upper_env    = upper(var.environment)
  lower_region = lower(var.region)
  joined       = join("-", ["web", var.environment, "server"])
  formatted    = format("instance-%03d", count.index + 1)
}

# Collection functions
locals {
  merged_tags = merge(var.default_tags, var.custom_tags)
  subnet_ids  = concat(var.public_subnets, var.private_subnets)
  unique_azs  = distinct(var.availability_zones)
  first_az    = element(var.availability_zones, 0)
}

# Type conversions
locals {
  port_number = tonumber(var.port)
  is_enabled  = tobool(var.enabled)
  tags_list   = tolist(var.tag_set)
}

# Encoding functions
locals {
  user_data = base64encode(file("userdata.sh"))
  config    = jsonencode(var.config_map)
}
```

## Provisioners

### Remote-exec Provisioner
```hcl
resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  provisioner "remote-exec" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y nginx"
    ]

    connection {
      type        = "ssh"
      user        = "ubuntu"
      private_key = file("~/.ssh/id_rsa")
      host        = self.public_ip
    }
  }
}
```

### Local-exec Provisioner
```hcl
resource "null_resource" "example" {
  provisioner "local-exec" {
    command = "echo ${aws_instance.web.public_ip} > ip_address.txt"
  }
}
```

## Import Blocks (Terraform 1.5+)

```hcl
import {
  to = aws_instance.example
  id = "i-1234567890abcdef0"
}
```

## Moved Blocks

```hcl
moved {
  from = aws_instance.old_name
  to   = aws_instance.new_name
}
```

## Detection Confidence

- **HIGH**: Files with .tf extension
- **HIGH**: terraform block with required_version
- **HIGH**: provider blocks
- **HIGH**: resource or data blocks
- **MEDIUM**: .tfvars files
- **MEDIUM**: terraform.tfstate files
- **LOW**: HCL syntax without Terraform-specific blocks
