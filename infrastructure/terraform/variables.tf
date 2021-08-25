variable "service_id" {
  type        = string
  default     = "polvo"
  description = "The id of the service. Should be a single word"
}

variable "service_base_domain" {
  type        = string
  default     = "truepro.fit"
  description = "Base do main for the service. It should be already exists"
}

variable "project_id" {
  type        = string
  default     = "trueprofit-frontend"
  description = "The project name"
}

variable "env" {
  type        = string
  default     = "develop"
  description = "The environment"
}

variable "region" {
  type        = string
  default     = "asia-southeast1"
  description = "The region"
}

variable "docker_image_url" {
  type        = string
  description = "The image digest"
}


variable "managed_dns_zone" {
  type        = string
  default     = "trueprofit"
  description = "The managed zone to create domain"
}

locals {
  service_domain    = "${var.service_id}.${var.service_base_domain}"
  service_full_name = "${var.service_id}-service"
}
