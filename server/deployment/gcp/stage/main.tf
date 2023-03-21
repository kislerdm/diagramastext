terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.56.0"
    }
  }

  backend "gcs" {
    bucket = "tf-diagramastext-stage"
  }
}

provider "google" {
  project = "diagramastext-stage"
  region  = "us-central1"
}

data "google_project" "project" {}

variable "imagetag" {
  type        = string
  description = "Docker image tag."
}

module "core" {
  source     = "../stack"
  project    = "diagramastext-stage"
  imagetag   = var.imagetag
  location   = "us-central1"
  api_domain = "api-stage.diagramastext.dev"
  cors_headers = {
    "Access-Control-Allow-Origin"  = "https://stage.diagramastext.dev"
    "Access-Control-Allow-Headers" = "Content-Type,X-Amz-Date,x-api-key,Authorization,X-Api-Key,X-Amz-Security-Token"
    "Access-Control-Allow-Methods" = "POST,OPTIONS"
  }
}
