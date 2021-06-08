output "docker_image_url" {
  value = local.docker_image_url
}
output "service_address" {
  value = google_dns_record_set.resource_recordset.name
}
