terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.56.0"
    }
  }

  backend "gcs" {
    bucket = "tf-diagramastext-root"
    prefix = "deploymentenvironment"
  }
}

provider "google" {
  project = "diagramastext-root"
  region  = "us-central1"
}

locals {
  is_prod                    = terraform.workspace == "production" || terraform.workspace == "default"
  root_sa_email              = "admin-sa@diagramastext-root.iam.gserviceaccount.com"
  env_suffix                 = local.is_prod ? "prod" : "stage"
  deploymentenv_project_name = local.is_prod ? "production" : "staging"
}
