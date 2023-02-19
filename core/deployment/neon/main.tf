terraform {
  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "~> 0.1.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }

  backend "s3" {
    bucket = "dev.diagramastext.terraform"
    key    = "neon"
    region = "us-east-2"
  }
}

provider "neon" {}

provider "aws" {
  region = "us-east-2"
}

locals {
  neon_branch_id  = "br-steep-silence-472824"
}

resource "neon_project" "this" {
  name = "diagramastext"
}

#resource "neon_database" "this" {
#  name       = "core"
#  project_id = neon_project.this.id
#  branch_id  = local.neon_branch_id
#  owner_name = local.neon_owner_name
#}

resource "neon_branch" "stage" {
  name       = "staging"
  project_id = neon_project.this.id
  parent_id  = local.neon_branch_id
}

resource "neon_role" "lambda" {
  name       = "lambda"
  project_id = neon_project.this.id
  branch_id  = local.neon_branch_id
}

resource "neon_role" "lambda-stg" {
  name       = "lambda-stg"
  project_id = neon_project.this.id
  branch_id  = neon_branch.stage.id
}

resource "aws_secretsmanager_secret" "neon_lambda" {
  name                    = "neon/main/core/lambda"
  recovery_window_in_days = 0
}

#resource "aws_secretsmanager_secret_version" "neon_lambda" {
#  secret_id = aws_secretsmanager_secret.neon_lambda.id
#  secret_string = jsonencode({
#    branch_id  = local.neon_branch_id
#    host       = local.neon_endpoint
#    project_id = neon_project.this.id
#    dbname     = neon_database.this.id
#    user       = neon_role.lambda.name
#    password   = neon_role.lambda.password
#  })
#}
