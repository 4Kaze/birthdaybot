variable "project_id" {
  type        = string
  description = "The id of project where resources will be created"
}

variable "service_location" {
  type        = string
  description = "The location of the service"
  default     = "europe-central2"
}

variable "manager_container_image" {
  type = string
  description = "The image of the manager container"
}

variable "notifier_container_image" {
  type = string
  description = "The image of the notifier container"
}

variable "telegram_bot_token" {
  type        = string
  description = "The bot token from the BotFather"
}

variable "database_connection_uri" {
  type        = string
  description = "The connection uri to postgres database"
}