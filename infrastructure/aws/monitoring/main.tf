terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }

  backend "s3" {
    bucket = "dev.diagramastext.terraform"
    key    = "monitoring"
    region = "us-east-2"
  }
}

provider "aws" {
  region = "us-east-2"
  default_tags {
    tags = {
      environment = "production"
      system      = "monitoring"
    }
  }
}

locals {
  account_id = "027889758114"
  region     = "us-east-2"
}

# alarm notifications are published to the slack channel using AWS Chatbot
# Chatbot configured manually:
# - https://docs.aws.amazon.com/chatbot/latest/adminguide/slack-setup.html
# - https://docs.aws.amazon.com/chatbot/latest/adminguide/test-notifications-cw.html
# Guardian policies applied:
# - CloudWatchEventsReadOnlyAccess
# - CloudWatchLogsReadOnlyAccess
# - AmazonLookoutMetricsReadOnlyAccess
# IAM role created: arn:aws:iam::027889758114:role/service-role/monitoring-slack-production
# Role setting: Channel role
