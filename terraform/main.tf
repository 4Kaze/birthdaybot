provider "google" {
  project = var.project_id
}

locals {
  manager_service_name  = "birthday-bot-manager"
  notifier_service_name = "birthday-bot-notifier"
}

resource "google_cloud_run_v2_service" "manager_service" {
  name     = local.manager_service_name
  location = var.service_location

  template {
    containers {
      image = var.manager_container_image
      env {
        name  = "K8S_SERVICE_NAME"
        value = "namespaces/${var.project_id}/services/${local.manager_service_name}"
      }
      env {
        name  = "LOCATION"
        value = var.service_location
      }
      env {
        name  = "DATABASE_URL"
        value = var.database_connection_uri
      }
      env {
        name  = "BOT_TOKEN"
        value = var.telegram_bot_token
      }
    }
    timeout                          = "2s"
    max_instance_request_concurrency = 10
  }
}

resource "google_cloud_run_service_iam_binding" "manager_allow_unauthenticated_invocations" {
  location = google_cloud_run_v2_service.manager_service.location
  service  = google_cloud_run_v2_service.manager_service.name
  members  = ["allUsers"]
  role     = "roles/run.invoker"
}

resource "random_id" "random_id" {
  byte_length = 8
}

resource "google_cloud_tasks_queue" "birthdays_queue" {
  name     = "birthday-notifications-${random_id.random_id.dec}" # workaround for a 7-day wait time to create a queue with the same name
  location = var.service_location
  rate_limits {
    max_concurrent_dispatches = 1
  }
  retry_config {
    max_attempts = 3
    min_backoff  = "60s"
  }
}

resource "google_cloud_run_v2_service" "notifier_service" {
  name     = local.notifier_service_name
  location = var.service_location

  template {
    max_instance_request_concurrency = 1
    timeout                          = "120s"
    scaling {
      max_instance_count = 1
      min_instance_count = 0
    }
    containers {
      image = var.notifier_container_image
      resources {
        limits = {
          memory = "1024Mi"
        }
      }
      env {
        name  = "K8S_SERVICE_NAME"
        value = "namespaces/${var.project_id}/services/${local.notifier_service_name}"
      }
      env {
        name  = "LOCATION"
        value = var.service_location
      }
      env {
        name  = "BOT_TOKEN"
        value = var.telegram_bot_token
      }
      env {
        name  = "MANAGER_URL"
        value = google_cloud_run_v2_service.manager_service.uri
      }
      env {
        name  = "QUEUE_ID"
        value = google_cloud_tasks_queue.birthdays_queue.id
      }
      env {
        name  = "TASK_DEADLINE_S"
        value = 3 * 60
      }
      env {
        name  = "TASK_DELAY_S"
        value = 30
      }
    }
  }
}

resource "google_cloud_run_service_iam_binding" "notifier_allow_unauthenticated_invocations" {
  location = google_cloud_run_v2_service.notifier_service.location
  service  = google_cloud_run_v2_service.notifier_service.name
  members  = ["allUsers"]
  role     = "roles/run.invoker"
}

resource "google_cloud_scheduler_job" "job" {
  name             = "notifier-job"
  schedule         = "0 9 * * *"
  time_zone        = "CET"
  attempt_deadline = "60s"
  region           = var.service_location

  http_target {
    http_method = "POST"
    uri         = "${google_cloud_run_v2_service.notifier_service.uri}/schedule"
  }
}
