output "vm_ip" {
  description = "The external IP address of the VM."
  value       = google_compute_address.yodahunters_ip.address
}

output "ssh_command" {
  description = "SSH command to connect to the VM via IAP tunnel."
  value       = "gcloud compute ssh yodahunters --zone ${var.zone} --tunnel-through-iap --project ${var.project_id}"
}

output "backup_bucket" {
  description = "The GCS bucket for database backups."
  value       = google_storage_bucket.backups.name
}

output "github_wif_provider" {
  description = "Full resource name of the GitHub WIF provider. Set as the WIF_PROVIDER GitHub Actions secret."
  value       = google_iam_workload_identity_pool_provider.github.name
}

output "github_ci_service_account" {
  description = "Email of the CI service account GitHub Actions impersonates. Set as the WIF_SERVICE_ACCOUNT GitHub Actions secret."
  value       = google_service_account.github_ci.email
}
