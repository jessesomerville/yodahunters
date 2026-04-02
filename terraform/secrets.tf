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
}
