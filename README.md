
# Module `./config/terraform`

Provider Requirements:
* **docker (`kreuzwerker/docker`):** `2.11.0`
* **google:** (any version)
* **google-beta:** (any version)

## Input Variables
* `env` (default `"develop"`): The environment
* `project_name` (default `"aio-shopify-services"`): The project name
* `region` (default `"asia-southeast1"`): The region
* `service_base_domain` (default `"aiocean.services"`): Base do main for the service. It should be already exists
* `service_name` (default `"polvo"`): The id of the service. Should be a single word

## Output Values
* `docker_image_url`
* `service_address`

## Managed Resources
* `google_cloud_run_domain_mapping.default` from `google-beta`
* `google_cloud_run_service.default` from `google-beta`
* `google_cloud_run_service_iam_policy.noauth` from `google`
* `google_dns_record_set.resource_recordset` from `google-beta`
* `google_service_account.firebase_account` from `google`

## Data Resources
* `data.docker_registry_image.service_image` from `docker`
* `data.google_client_config.default` from `google`
* `data.google_iam_policy.noauth` from `google`

## Test

```
# Set env
export ADDRESS=127.0.0.1:8080
export K_REVISION=local
export FIRESTORE_EMULATOR_HOST=localhost:8181

# Start firebase emulator
gcloud beta emulators firestore start --host-port=$FIRESTORE_EMULATOR_HOST

# Start server
go test ./internal/server -count=1
```
