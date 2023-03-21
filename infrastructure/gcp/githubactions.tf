resource "google_project_service" "this" {
  project = google_project.this.project_id
  service = "iamcredentials.googleapis.com"

  timeouts {
    create = "30m"
    update = "40m"
  }

  disable_dependent_services = true
  disable_on_destroy         = true

  depends_on = [google_project.this]
}

resource "google_iam_workload_identity_pool" "this" {
  project                   = google_project.this.project_id
  workload_identity_pool_id = "github-access-pool"
  display_name              = "Github Access Pool"
  description               = "Pool to authorize terraform SA"
  disabled                  = false

  depends_on = [google_project_service.this]
}

resource "google_iam_workload_identity_pool_provider" "this" {
  project = google_project.this.project_id

  workload_identity_pool_provider_id = "github-oidc"
  workload_identity_pool_id          = google_iam_workload_identity_pool.this.workload_identity_pool_id

  display_name = "Github OIDC"
  description  = "OIDC to AuthN/Z project terraform SA to provision resourced during Github Action job"

  attribute_mapping = {
    "google.subject"       = "assertion.sub",
    "attribute.actor"      = "assertion.actor"
    "attribute.repository" = "assertion.repository"
  }

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }

  disabled = false

  depends_on = [google_project_service.this]
}

resource "google_service_account_iam_binding" "this" {
  service_account_id = google_service_account.project_admin.name
  role               = "roles/iam.workloadIdentityUser"
  members = [
    "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.this.name}/attribute.repository/kislerdm/diagramastext"
  ]
}
