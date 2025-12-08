# Terraform Module Template

## Estructura
```
module-name/
├── main.tf         # Recursos principales
├── variables.tf    # Variables de entrada
├── outputs.tf      # Outputs
├── versions.tf     # Provider versions
├── locals.tf       # Local values
└── README.md       # Documentación
```

## main.tf
```hcl
# Main resources go here
resource "aws_example" "main" {
  name = var.name
  
  tags = merge(local.common_tags, {
    Name = var.name
  })
}
```

## variables.tf
```hcl
variable "name" {
  description = "Name of the resource"
  type        = string
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "tags" {
  description = "Additional tags"
  type        = map(string)
  default     = {}
}
```

## outputs.tf
```hcl
output "id" {
  description = "ID of the created resource"
  value       = aws_example.main.id
}

output "arn" {
  description = "ARN of the created resource"
  value       = aws_example.main.arn
}
```

## versions.tf
```hcl
terraform {
  required_version = ">= 1.5.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}
```

## locals.tf
```hcl
locals {
  common_tags = merge(var.tags, {
    Environment = var.environment
    ManagedBy   = "Terraform"
    Module      = "module-name"
  })
}
```

## README.md
```markdown
# Module Name

Description of what this module does.

## Usage

\`\`\`hcl
module "example" {
  source = "./modules/module-name"
  
  name        = "my-resource"
  environment = "prod"
  
  tags = {
    Project = "my-project"
  }
}
\`\`\`

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| name | Name of the resource | string | - | yes |
| environment | Environment | string | - | yes |
| tags | Additional tags | map(string) | {} | no |

## Outputs

| Name | Description |
|------|-------------|
| id | ID of the resource |
| arn | ARN of the resource |
```
