output "secret_id" {
  value       = google_secret_manager_secret.this.secret_id
  description = "Secret ID."
}

output "core_url" {
  value       = google_cloud_run_v2_service.this.uri
  description = "Core logic URL."
}
