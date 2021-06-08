provider "google" {
  region  = var.region
  project = var.project_id
}

provider "google-beta" {
  region  = var.region
  project = var.project_id
}

resource "google_cloud_run_service" "default" {
  name                       = local.service_full_name
  location                   = var.region
  provider                   = google-beta
  autogenerate_revision_name = true
  template {
    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" = "100"
        "client.knative.dev/user-image"    = local.docker_image_url
        "run.googleapis.com/client-name"   = "terraform"
      }
    }
    spec {
      container_concurrency = 80
      containers {
        image = local.docker_image_url
        ports {
          container_port = 8080
          name           = "h2c"
        }
        env {
          name  = "ENVIRONMENT"
          value = var.env
        }
        env {
          name  = "SERVICE_BASE_DOMAIN"
          value = var.service_base_domain
        }
        env {
          name  = "SERVICE_NAME"
          value = var.service_id
        }
        env {
          name  = "GOOGLE_SERVICE_ACCOUNT"
          value = google_service_account.firebase_account.id
        }

        resources {
          limits = {
            cpu    = "1000m"
            memory = "256Mi"
          }
        }
      }
    }
  }

  metadata {
    annotations = {
      "client.knative.dev/user-image"     = local.docker_image_url
      "run.googleapis.com/ingress"        = "all"
      "run.googleapis.com/ingress-status" = "all"
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth" {
  location = google_cloud_run_service.default.location
  project  = google_cloud_run_service.default.project
  service  = google_cloud_run_service.default.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_cloud_run_domain_mapping" "default" {
  location = var.region
  name     = local.service_domain
  provider = google-beta

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = google_cloud_run_service.default.name
  }
}

resource "google_dns_record_set" "resource_recordset" {
  provider     = google-beta
  managed_zone = "aiocean-services"
  name         = "${local.service_domain}."
  type         = "CNAME"
  rrdatas      = ["ghs.googlehosted.com."]
  ttl          = 86400
}

data "google_client_config" "default" {

}

resource "google_service_account" "firebase_account" {
  account_id   = "${var.service_id}-firebase"
  display_name = "Service Account"
}
