# Enable APIs
locals {
  services = [
    "compute.googleapis.com",
    "storage.googleapis.com",
    "iap.googleapis.com",
    "secretmanager.googleapis.com",
  ]
}

resource "google_project_service" "project_services" {
  for_each           = toset(local.services)
  service            = each.value
  disable_on_destroy = false
}
