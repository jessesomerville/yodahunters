variable "project_id" {
  description = "The GCP Project ID."
  type        = string
}

variable "region" {
  description = "The region to deploy resources to."
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "The zone to deploy the VM to."
  type        = string
  default     = "us-central1-a"
}

variable "ssh_user_emails" {
  description = "Google account emails allowed to SSH into the VM via IAP."
  type        = set(string)
}

variable "db_password" {
  description = "Initial value to seed into the yodahunters-db-password Secret Manager secret. Leave unset to manage secret values out-of-band (e.g. via `gcloud secrets versions add`)."
  type        = string
  sensitive   = true
  default     = null
}

variable "jwt_secret_value" {
  description = "Initial value to seed into the yodahunters-jwt-secret Secret Manager secret. Leave unset to manage secret values out-of-band (e.g. via `gcloud secrets versions add`)."
  type        = string
  sensitive   = true
  default     = null
}