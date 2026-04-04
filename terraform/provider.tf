provider "google" {
  project = var.project_id
  region  = var.region
}

terraform {
  required_version = "~> 1.9"

  # Partial backend config — supply the bucket at init time:
  #   terraform init -backend-config="bucket=<YOUR_STATE_BUCKET>"
  backend "gcs" {
    prefix = "yodahunters/state"
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}
