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
