resource "google_container_registry" "core" {
  project = var.project
}

resource "google_secret_manager_secret" "this" {
  secret_id = "core"

  replication {
    automatic = true
  }
}

locals {
  image_core = "gcr.io/${var.project}/core:${var.imagetag}"
}

data "google_project" "this" {}

resource "google_service_account" "this" {
  account_id   = "core-runner"
  display_name = "Core runner role"
  project      = var.project
}

resource "google_secret_manager_secret_iam_member" "this" {
  secret_id  = google_secret_manager_secret.this.id
  role       = "roles/secretmanager.secretAccessor"
  member     = "serviceAccount:${google_service_account.this.email}"
  depends_on = [google_secret_manager_secret.this]
}

resource "google_project_iam_member" "this" {
  project  = var.project
  for_each = toset(["roles/run.invoker", "roles/monitoring.viewer", "roles/logging.viewer"])
  role     = each.key
  member   = "serviceAccount:${google_service_account.this.email}"
}

resource "google_cloud_run_v2_service" "this" {
  name     = "core"
  location = "us-central1"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.this.email

    scaling {
      max_instance_count = 5
    }

    containers {
      name  = "core"
      image = local.image_core

      env {
        name  = "ACCESS_CREDENTIALS_URI"
        value = "projects/${data.google_project.this.number}/secrets/${google_secret_manager_secret.this.secret_id}"
      }
      env {
        name  = "MODEL_MAX_TOKENS"
        value = var.model_max_tokens
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

resource "google_cloud_run_v2_service_iam_member" "this" {
  project  = google_cloud_run_v2_service.this.project
  location = google_cloud_run_v2_service.this.location
  name     = google_cloud_run_v2_service.this.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_cloud_run_domain_mapping" "this" {
  count = var.api_domain != "" ? 1 : 0

  location = var.location
  name     = var.api_domain

  metadata {
    namespace = var.project
  }

  spec {
    route_name = google_cloud_run_v2_service.this.name
  }
}
