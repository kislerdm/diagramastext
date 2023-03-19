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

module "core" {
  source  = "./../stack"
  project = "diagramastext-stage"
}
