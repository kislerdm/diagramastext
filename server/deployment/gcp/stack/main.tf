variable "project" {
  type        = string
  description = "Project ID."
}

resource "google_container_registry" "core" {
  project = var.project
}

resource "google_secret_manager_secret" "this" {
  secret_id = "core"

  replication {
    automatic = true
  }
}

output "secret_id" {
  value       = google_secret_manager_secret.this.secret_id
  description = "Secret ID."
}

variable "imagetag" {
  type        = string
  description = "Docker image tag."
}

locals {
  image_core = "gcr.io/diagramastext-stage/core:${var.imagetag}"
}

data "google_project" "project" {}

resource "google_service_account" "core_runner" {
  account_id   = "core-runner"
  display_name = "Core runner role"
  project      = var.project
}

resource "google_secret_manager_secret_iam_member" "core_runner_secret" {
  secret_id  = google_secret_manager_secret.this.id
  role       = "roles/secretmanager.secretAccessor"
  member     = "serviceAccount:${google_service_account.core_runner.email}"
  depends_on = [google_secret_manager_secret.this]
}

resource "google_project_iam_member" "core_runner" {
  project  = var.project
  for_each = toset(["roles/run.invoker", "roles/monitoring.viewer", "roles/logging.viewer"])
  role     = each.key
  member   = "serviceAccount:${google_service_account.core_runner.email}"
}

variable "cors_headers" {
  type        = map(string)
  description = "CORS headers map."
  default     = null
}

resource "google_cloud_run_v2_service" "core" {
  name     = "core"
  location = "us-central1"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.core_runner.email

    scaling {
      max_instance_count = 5
    }

    containers {
      name  = "core"
      image = local.image_core

      env {
        name  = "ACCESS_CREDENTIALS_URI"
        value = "projects/${data.google_project.project.number}/secrets/${google_secret_manager_secret.this.secret_id}"
      }
      env {
        name  = "MODEL_MAX_TOKENS"
        value = "500"
      }

      dynamic "env" {
        for_each = var.cors_headers != null ? { foo = 0 } : {}
        content {
          name  = "CORS_HEADERS"
          value = jsonencode(var.cors_headers)
        }
      }

      ports {
        container_port = 9000
      }

      liveness_probe {
        http_get {
          path = "/status"
        }
        failure_threshold = 2
      }
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}

resource "google_cloud_run_v2_service_iam_member" "core_runner" {
  project  = google_cloud_run_v2_service.core.project
  location = google_cloud_run_v2_service.core.location
  name     = google_cloud_run_v2_service.core.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

output "core_url" {
  value       = google_cloud_run_v2_service.core.uri
  description = "Core logic URL."
}
