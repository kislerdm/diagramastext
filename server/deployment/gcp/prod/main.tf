terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.56.0"
    }
  }

  backend "gcs" {
    bucket = "tf-diagramastext-prod"
  }
}

provider "google" {
  project = "diagramastext-prod"
  region  = "us-central1"
}

data "google_project" "project" {}

variable "imagetag" {
  type        = string
  description = "Docker image tag."
  default     = ""
}

module "core" {
  source   = "../stack"
  project  = "diagramastext-prod"
  imagetag = var.imagetag
}

output "secrets_uri" {
  value = "projects/${data.google_project.project.number}/secrets/${module.core.secret_id}"
}
