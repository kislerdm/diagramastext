terraform {
  backend "s3" {
    bucket               = "dev.diagramastext.terraform"
    key                  = "core"
    region               = "us-east-2"
    workspace_key_prefix = "environment"
  }
}

