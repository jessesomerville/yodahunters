# Secret Manager for JWT Secret
resource "google_secret_manager_secret" "jwt_secret" {
  secret_id = "yodahunters-jwt-secret"

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }

  depends_on = [google_project_service.project_services]
}

resource "google_secret_manager_secret_version" "jwt_secret_v1" {
  secret      = google_secret_manager_secret.jwt_secret.id
  secret_data = var.jwt_secret_value

  lifecycle {
    ignore_changes = [secret_data]
  }
}

# Secret Manager for Database Password
resource "google_secret_manager_secret" "db_password" {
  secret_id = "yodahunters-db-password"

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }

  depends_on = [google_project_service.project_services]
}

resource "google_secret_manager_secret_version" "db_password_v1" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = var.db_password

  lifecycle {
    ignore_changes = [secret_data]
  }
}

# Allow the VM service account to read both secrets
resource "google_secret_manager_secret_iam_member" "jwt_secret_accessor" {
  secret_id = google_secret_manager_secret.jwt_secret.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.yodahunters_sa.email}"
}

resource "google_secret_manager_secret_iam_member" "db_password_accessor" {
  secret_id = google_secret_manager_secret.db_password.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.yodahunters_sa.email}"
}
