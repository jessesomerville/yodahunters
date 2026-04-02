# Cloud SQL (Postgres)
resource "google_sql_database_instance" "yodahunters_instance" {
  name             = var.db_instance_name
  database_version = "POSTGRES_18"
  region           = var.region

  settings {
    tier    = "db-f1-micro"
    edition = "ENTERPRISE"
    ip_configuration {
      ipv4_enabled = true
    }
    backup_configuration {
      enabled = true
    }

    database_flags {
      name  = "cloudsql.iam_authentication"
      value = "on"
    }
  }

  deletion_protection = false

  depends_on = [google_project_service.project_services]
}

resource "google_sql_database" "yodahunters_db" {
  name     = var.db_name
  instance = google_sql_database_instance.yodahunters_instance.name
}

resource "google_sql_user" "iam_user" {
  name     = trimsuffix(google_service_account.yodahunters_sa.email, ".gserviceaccount.com")
  instance = google_sql_database_instance.yodahunters_instance.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}

# Still keeping a built-in user for local dev/migrations if needed
resource "google_sql_user" "yodahunters_user" {
  name     = var.db_user
  instance = google_sql_database_instance.yodahunters_instance.name
  password = var.db_password
}
