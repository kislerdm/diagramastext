variable "project" {
  type        = string
  description = "Project ID."
}

resource "google_container_registry" "core" {
  project  = var.project
}

output "gcr" {
  value = {
    core = google_container_registry.core.id
  }
  description = "GCR identifiers."
}
