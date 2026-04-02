variable "project_id" {
  description = "The GCP Project ID."
  type        = string
}

variable "region" {
  description = "The region to deploy resources to."
  type        = string
  default     = "us-central1"
}

variable "db_instance_name" {
  description = "The name of the Cloud SQL instance."
  type        = string
  default     = "yodahunters-db"
}

variable "db_name" {
  description = "The name of the database."
  type        = string
  default     = "yodahunters-db"
}

variable "db_user" {
  description = "The database user."
  type        = string
  default     = "yodahunters-user"
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

variable "domain_name" {
  description = "The domain name for the application."
  type        = string
  default     = "yodahunters.com"
}
