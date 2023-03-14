resource "google_project_iam_custom_role" "developer" {
  project     = google_project.this.project_id
  role_id     = "developer"
  title       = "Developer"
  description = "Developer's access role"
  permissions = [
    "firebaseauth.users.get",
    "firebaseauth.users.create",
    "firebaseauth.users.createSession",
    "firebaseauth.configs.get",
    "firebaseauth.configs.getHashConfig",
    "firebaseauth.users.sendEmail",
    "identitytoolkit.tenants.get",
    "identitytoolkit.tenants.getIamPolicy",
    "identitytoolkit.tenants.list",
  ]
}

# developers

locals {
  developers = [
    "medvedevtimofei@gmail.com",
    "deim.medvedev@gmail.com",
    "pyvinci@gmail.com",
    "diego.hordi@gmail.com",
  ]
}

resource "google_project_iam_member" "developer" {
  project  = google_project.this.project_id
  role     = google_project_iam_custom_role.developer.id
  for_each = toset(local.developers)
  member   = "user:${each.key}"
}
