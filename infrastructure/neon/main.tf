terraform {
  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "~> 0.1.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "4.56.0"
    }
  }

  backend "gcs" {
    bucket = "tf-diagramastext-root"
    prefix = "neon"
  }
}

provider "google" {
  project = "diagramastext-root"
  region  = "us-central1"
}

provider "neon" {}

locals {
  neon_branch_id = "br-steep-silence-472824"
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
