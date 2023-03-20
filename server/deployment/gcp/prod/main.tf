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
  cors_headers = {
    "Access-Control-Allow-Origin"  = "https://diagramastext.dev"
    "Access-Control-Allow-Headers" = "Content-Type,X-Amz-Date,x-api-key,Authorization,X-Api-Key,X-Amz-Security-Token"
    "Access-Control-Allow-Methods" = "POST,OPTIONS"
  }
}

output "core_url" {
  value = module.core.core_url
}
