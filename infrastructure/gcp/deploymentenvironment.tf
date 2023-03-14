resource "google_project" "this" {
  name            = local.deploymentenv_project_name
  project_id      = "diagramastext-${local.env_suffix}"
  org_id          = "1067387145153"
  billing_account = "01389A-81FA3D-9CE293"
  labels = {
    firebase = "enabled"
  }
}

resource "google_project_iam_member" "root_admin" {
  project    = google_project.this.project_id
  role       = "roles/owner"
  member     = "serviceAccount:${local.root_sa_email}"
  depends_on = [google_project.this]
}

resource "google_service_account_iam_member" "root_admin" {
  service_account_id = "projects/diagramastext-root/serviceAccounts/${local.root_sa_email}"
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${local.root_sa_email}"
  depends_on         = [google_project.this]
}

resource "google_storage_bucket" "projects_tf" {
  location = "us-central1"
  project  = google_project.this.project_id

  name = "tf-${google_project.this.project_id}"

  public_access_prevention = "enforced"

  force_destroy = true
  versioning {
    enabled = false
  }

  depends_on = [google_project.this]
}

resource "google_service_account" "project_admin" {
  account_id   = "terraform"
  display_name = "Terraform Admin"
  description  = "Role to provision project's resources"
  project      = google_project.this.project_id
  depends_on   = [google_project.this]
}

resource "google_service_account_iam_member" "project_admin" {
  service_account_id = google_service_account.project_admin.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.project_admin.email}"
  depends_on         = [google_project.this]
}

resource "google_project_iam_member" "project_admin" {
  project    = google_service_account.project_admin.project
  role       = "roles/owner"
  member     = "serviceAccount:${google_service_account.project_admin.email}"
  depends_on = [google_project.this]
}
