terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    neon = {
      source  = "kislerdm/neon"
      version = "~> 0.1.0"
    }
  }
}

provider "aws" {
  region = "us-east-2"
}

provider "neon" {}

locals {
  is_prod = terraform.workspace == "production" || terraform.workspace == "default"
  suffix  = local.is_prod ? "" : "-stg"
  subdomain_prefix = local.is_prod ? "" : "stage."

  cors_headers = {
    "Access-Control-Allow-Origin"  = "https://${local.subdomain_prefix}diagramastext.dev"
    "Access-Control-Allow-Headers" = "Content-Type,X-Amz-Date,x-api-key,Authorization,X-Api-Key,X-Amz-Security-Token"
    "Access-Control-Allow-Methods" = "POST,OPTIONS"
  }
}
