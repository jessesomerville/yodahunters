# Service account used by the GitHub Actions deploy workflow.
resource "google_service_account" "github_ci" {
  account_id   = "yodahunters-github-ci"
  display_name = "Yodahunters GitHub Actions CI"
}

# Workload Identity Pool for GitHub OIDC tokens.
resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github-actions-pool"
  display_name              = "GitHub Actions Pool"
  description               = "Identity pool for GitHub Actions OIDC federation"

  depends_on = [google_project_service.project_services]
}

# Workload Identity Provider for GitHub.
resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"
  display_name                       = "GitHub OIDC Provider"

  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.repository" = "assertion.repository"
    "attribute.ref"        = "assertion.ref"
  }

  attribute_condition = "assertion.repository == \"jessesomerville/yodahunters\""

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

# Allow workflows running in the configured GitHub repo to impersonate the CI service account.
resource "google_service_account_iam_member" "github_wif_user" {
  service_account_id = google_service_account.github_ci.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/jessesomerville/yodahunters"
}

# Permit the CI service account to open an IAP tunnel to the VM.
resource "google_iap_tunnel_instance_iam_member" "github_ci_tunnel" {
  project  = var.project_id
  zone     = var.zone
  instance = google_compute_instance.yodahunters.name
  role     = "roles/iap.tunnelResourceAccessor"
  member   = "serviceAccount:${google_service_account.github_ci.email}"
}

# Permit the CI service account to SSH into the VM via OS Login (non-root POSIX user).
resource "google_project_iam_member" "github_ci_oslogin" {
  project = var.project_id
  role    = "roles/compute.osLogin"
  member  = "serviceAccount:${google_service_account.github_ci.email}"
}

# The deploy script needs sudo to restart the systemd unit and write to /opt/yodahunters.
resource "google_project_iam_member" "github_ci_oslogin_admin" {
  project = var.project_id
  role    = "roles/compute.osAdminLogin"
  member  = "serviceAccount:${google_service_account.github_ci.email}"
}

# gcloud compute ssh to a VM with an attached service account requires actAs on that SA.
resource "google_service_account_iam_member" "github_ci_act_as_vm_sa" {
  service_account_id = google_service_account.yodahunters_sa.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.github_ci.email}"
}
