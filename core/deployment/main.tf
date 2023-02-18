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
  suffix = var.environment == "production" ? "" : "-stg"
}
