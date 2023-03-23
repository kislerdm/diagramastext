# Core logic stack

<!-- BEGIN_TF_DOCS -->
## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [google_cloud_run_domain_mapping.this](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_domain_mapping) | resource |
| [google_cloud_run_v2_service.this](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_v2_service) | resource |
| [google_cloud_run_v2_service_iam_member.this](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_v2_service_iam_member) | resource |
| [google_container_registry.core](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/container_registry) | resource |
| [google_project_iam_member.this](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/project_iam_member) | resource |
| [google_secret_manager_secret.this](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/secret_manager_secret) | resource |
| [google_secret_manager_secret_iam_member.this](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/secret_manager_secret_iam_member) | resource |
| [google_service_account.this](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/service_account) | resource |
| [google_project.this](https://registry.terraform.io/providers/hashicorp/google/latest/docs/data-sources/project) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_api_domain"></a> [api\_domain](#input\_api\_domain) | API domain to map the Cloud Run to. | `string` | `""` | no |
| <a name="input_cors_headers"></a> [cors\_headers](#input\_cors\_headers) | CORS headers map. | `map(string)` | `null` | no |
| <a name="input_imagetag"></a> [imagetag](#input\_imagetag) | Docker image tag. | `string` | n/a | yes |
| <a name="input_location"></a> [location](#input\_location) | Geo-location. | `string` | `"us-central1"` | no |
| <a name="input_model_max_tokens"></a> [model\_max\_tokens](#input\_model\_max\_tokens) | https://platform.openai.com/docs/api-reference/chat/create#chat/create-max_tokens | `number` | `500` | no |
| <a name="input_project"></a> [project](#input\_project) | Project ID. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_core_url"></a> [core\_url](#output\_core\_url) | Core logic URL. |
| <a name="output_secret_id"></a> [secret\_id](#output\_secret\_id) | Secret ID. |
<!-- END_TF_DOCS -->