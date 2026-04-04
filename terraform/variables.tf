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

variable "domain_name" {
  description = "The domain name for the application."
  type        = string
  default     = "yodahunters.com"
}

variable "ssh_user_email" {
  description = "The Google account email allowed to SSH into the VM via IAP."
  type        = string
}

variable "db_password" {
  description = "The password for the database user."
  type        = string
  sensitive   = true
}

variable "jwt_secret_value" {
  description = "The value for the JWT secret."
  type        = string
  sensitive   = true
}
