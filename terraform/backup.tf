# Service account for the VM
resource "google_service_account" "yodahunters_sa" {
  account_id   = "yodahunters-vm"
  display_name = "Yodahunters VM Service Account"
}

# Allow SA to write to the backup bucket
resource "google_storage_bucket_iam_member" "backup_writer" {
  bucket = google_storage_bucket.backups.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.yodahunters_sa.email}"
}

# GCS bucket for database backups
resource "google_storage_bucket" "backups" {
  name     = "${var.project_id}-db-backups"
  location = var.region

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type = "Delete"
    }
  }

  depends_on = [google_project_service.project_services]
}
