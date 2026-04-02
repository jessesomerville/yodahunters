output "db_instance_connection_name" {
  description = "The connection name of the Cloud SQL instance."
  value       = google_sql_database_instance.yodahunters_instance.connection_name
}

output "service_account_email" {
  description = "The email of the Cloud Run service account."
  value       = google_service_account.yodahunters_sa.email
}

output "artifact_registry_repository" {
  description = "The URL of the Artifact Registry repository."
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.yodahunters_repo.name}"
}

output "jwt_secret_name" {
  description = "The name of the JWT secret in Secret Manager."
  value       = google_secret_manager_secret.jwt_secret.name
}

output "db_password_secret_name" {
  description = "The name of the DB password secret in Secret Manager."
  value       = google_secret_manager_secret.db_password.name
}
