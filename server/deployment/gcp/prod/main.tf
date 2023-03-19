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

module "core" {
  source  = "../stack"
  project = "diagramastext-prod"
}
