#  Terraform module to provision a Go lambda

<!-- BEGIN_TF_DOCS -->
## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | n/a |
| <a name="provider_null"></a> [null](#provider\_null) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_log_group.logs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_group) | resource |
| [aws_iam_policy.logs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_role.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy_attachment.custom](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.logs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_lambda_function.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_function) | resource |
| [null_resource.this](https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource) | resource |
| [aws_iam_policy_document.logs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_codebase_rebuild_trigger"></a> [codebase\_rebuild\_trigger](#input\_codebase\_rebuild\_trigger) | Patterns of the codebase dirs which change will trigger lambda rebuild and redeployment. | <pre>object({<br>    base                 = string<br>    modules_dir_patterns = list(string)<br>  })</pre> | `null` | no |
| <a name="input_env_vars"></a> [env\_vars](#input\_env\_vars) | Key-value pairs to set environemnt variables. | `map(string)` | `{}` | no |
| <a name="input_exec_timeout_sec"></a> [exec\_timeout\_sec](#input\_exec\_timeout\_sec) | Lambda invocation timeout in sec. | `number` | `30` | no |
| <a name="input_lambda_memory"></a> [lambda\_memory](#input\_lambda\_memory) | Lambda RAM in Mb. | `number` | `256` | no |
| <a name="input_name"></a> [name](#input\_name) | Lambda function name. | `string` | n/a | yes |
| <a name="input_path_lambda_module"></a> [path\_lambda\_module](#input\_path\_lambda\_module) | Path to the lambda module's codebase, i.e. directory with main.go and go.mod files. | `string` | n/a | yes |
| <a name="input_policy_arn_list"></a> [policy\_arn\_list](#input\_policy\_arn\_list) | Custom policies to attache to the Lambda execution role. | `list(string)` | `[]` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Resource tags. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_arn"></a> [arn](#output\_arn) | Amazon Resource Name (ARN) identifying your Lambda Function. |
| <a name="output_function_name"></a> [function\_name](#output\_function\_name) | Unique name for your Lambda Function. |
| <a name="output_invoke_arn"></a> [invoke\_arn](#output\_invoke\_arn) | ARN to be used for invoking Lambda Function from API Gateway. |
<!-- END_TF_DOCS -->
