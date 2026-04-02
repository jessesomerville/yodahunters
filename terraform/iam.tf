# Service Account for Cloud Run
resource "google_service_account" "yodahunters_sa" {
  account_id   = "yodahunters-runner"
  display_name = "Cloud Run Service Account for YodaHunters"
}

# Allow SA to connect to Cloud SQL
resource "google_project_iam_member" "sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.yodahunters_sa.email}"
}

# Allow SA to access Secret Manager
resource "google_project_iam_member" "secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.yodahunters_sa.email}"
}

# Allow SA to be a Cloud SQL Instance User (for IAM Auth)
resource "google_project_iam_member" "sql_instance_user" {
  project = var.project_id
  role    = "roles/cloudsql.instanceUser"
  member  = "serviceAccount:${google_service_account.yodahunters_sa.email}"
}

# Allow SA to write logs
resource "google_project_iam_member" "logging_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.yodahunters_sa.email}"
}
