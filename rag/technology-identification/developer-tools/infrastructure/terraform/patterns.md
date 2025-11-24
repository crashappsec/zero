# Terraform

**Category**: developer-tools/infrastructure
**Description**: Infrastructure as Code tool for building, changing, and versioning infrastructure
**Homepage**: https://www.terraform.io

## Configuration Files

- `*.tf`
- `*.tfvars`
- `*.tfvars.json`
- `.terraform.lock.hcl`
- `terraform.tfstate`
- `terraform.tfstate.backup`

## Package Detection

### NPM
- `cdktf` (CDK for Terraform)
- `@cdktf/provider-*`

### PYPI
- `cdktf`
- `python-terraform`

### GO
- `github.com/hashicorp/terraform-exec`

## Environment Variables

- `TF_VAR_*`
- `TF_CLI_ARGS`
- `TF_DATA_DIR`
- `TF_INPUT`
- `TF_LOG`
- `TF_LOG_PATH`
- `TERRAFORM_CLOUD_TOKEN`
- `TF_TOKEN_*`

## Provider Patterns

Common providers to detect:
- `hashicorp/aws`
- `hashicorp/azurerm`
- `hashicorp/google`
- `hashicorp/kubernetes`

## Detection Notes

- Look for .tf files in repository
- Check for terraform.lock.hcl (indicates initialized project)
- Provider blocks indicate cloud infrastructure
- Module blocks indicate reusable infrastructure

## Detection Confidence

- **.tf files Detection**: 95% (HIGH)
- **terraform.lock.hcl Detection**: 95% (HIGH)
- **.tfvars Detection**: 90% (HIGH)
- **CDK packages Detection**: 85% (MEDIUM)
