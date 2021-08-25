terraform {
  backend "gcs" {
    bucket = "-backend-config='bucket=$PROJECT_ID-tfstate'"
    prefix = "-backend-config='prefix=polvoservice/$_ENV'"
  }
}
