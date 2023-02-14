terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }

  backend "s3" {
    bucket = "dev.diagramastext.terraform"
    key    = "monitoring"
    region = "us-east-2"
  }
}

provider "aws" {
  region = "us-east-2"
}
