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
}

module "core" {
  source           = "../stack"
  project          = "diagramastext-prod"
  imagetag         = var.imagetag
  location         = "us-central1"
  api_domain       = "api.diagramastext.dev"
  model_max_tokens = 600
  cors_headers = {
    "Access-Control-Allow-Origin"  = "https://diagramastext.dev"
    "Access-Control-Allow-Headers" = "Content-Type,Authorization"
    "Access-Control-Allow-Methods" = "POST,OPTIONS,GET"
  }
}
