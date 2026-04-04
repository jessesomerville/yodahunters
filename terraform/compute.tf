# Static external IP
resource "google_compute_address" "yodahunters_ip" {
  name   = "yodahunters-ip"
  region = var.region
}

# Firewall: allow HTTP and HTTPS from anywhere
resource "google_compute_firewall" "yodahunters_web" {
  name    = "yodahunters-allow-web"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["yodahunters"]
}

# Firewall: allow SSH only from Google's IAP range
resource "google_compute_firewall" "yodahunters_iap_ssh" {
  name    = "yodahunters-allow-iap-ssh"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["35.235.240.0/20"]
  target_tags   = ["yodahunters"]
}

# Allow the SSH user to tunnel through IAP
resource "google_iap_tunnel_instance_iam_member" "ssh_access" {
  project  = var.project_id
  zone     = var.zone
  instance = google_compute_instance.yodahunters.name
  role     = "roles/iap.tunnelResourceAccessor"
  member   = "user:${var.ssh_user_email}"
}

# e2-micro VM (free tier eligible in us-central1)
resource "google_compute_instance" "yodahunters" {
  name         = "yodahunters"
  machine_type = "e2-micro"
  zone         = var.zone
  tags         = ["yodahunters"]

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
      size  = 20
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.yodahunters_ip.address
    }
  }

  service_account {
    email  = google_service_account.yodahunters_sa.email
    scopes = ["cloud-platform"]
  }

  depends_on = [google_project_service.project_services]
}
