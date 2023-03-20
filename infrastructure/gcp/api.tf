resource "google_project_service" "containerregistry" {
  project = "diagramastext-${local.env_suffix}"
  service = "containerregistry.googleapis.com"

  timeouts {
    create = "30m"
    update = "40m"
  }

  disable_dependent_services = true
  disable_on_destroy         = true
}

resource "google_project_service" "artifactregistry" {
  project = "diagramastext-${local.env_suffix}"
  service = "artifactregistry.googleapis.com"

  timeouts {
    create = "30m"
    update = "40m"
  }

  disable_dependent_services = true
  disable_on_destroy         = true
}

resource "google_project_service" "secretmanager" {
  project = "diagramastext-${local.env_suffix}"
  service = "secretmanager.googleapis.com"

  timeouts {
    create = "30m"
    update = "40m"
  }

  disable_dependent_services = true
  disable_on_destroy         = true
}

resource "google_project_service" "iam" {
  project = "diagramastext-${local.env_suffix}"
  service = "iam.googleapis.com"

  timeouts {
    create = "30m"
    update = "40m"
  }

  disable_dependent_services = true
  disable_on_destroy         = true
}

resource "google_project_service" "run" {
  project = "diagramastext-${local.env_suffix}"
  service = "run.googleapis.com"

  timeouts {
    create = "30m"
    update = "40m"
  }

  disable_dependent_services = true
  disable_on_destroy         = true
}
