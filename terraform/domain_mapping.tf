# Cloud Run Domain Mapping (Free)
resource "google_cloud_run_domain_mapping" "yodahunters_mapping" {
  location = var.region
  name     = var.domain_name

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = "yodahunters" # Assumes the Cloud Run service is named "yodahunters"
  }
}

output "dns_records" {
  description = "The DNS records provided by Google to verify and point your domain. Add these as A/AAAA/CNAME records in your DNS provider's console."
  value       = google_cloud_run_domain_mapping.yodahunters_mapping.status[0].resource_records
}
