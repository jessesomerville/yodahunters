# Enable APIs
locals {
  services = [
    "run.googleapis.com",
    "sqladmin.googleapis.com",
    "secretmanager.googleapis.com",
    "artifactregistry.googleapis.com",
    "iam.googleapis.com",
    "cloudbuild.googleapis.com"
  ]
}

resource "google_project_service" "project_services" {
  for_each = toset(local.services)
  service  = each.value
  disable_on_destroy = false
}

# Artifact Registry for Docker images
resource "google_artifact_registry_repository" "yodahunters_repo" {
  location      = var.region
  repository_id = "yodahunters"
  description   = "Docker repository for yodahunters"
  format        = "DOCKER"

  depends_on = [google_project_service.project_services]
}
