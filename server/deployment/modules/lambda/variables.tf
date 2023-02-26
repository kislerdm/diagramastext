variable "name" {
  type        = string
  description = "Lambda function name."
}

variable "path_lambda_module" {
  type        = string
  description = "Path to the lambda module's codebase, i.e. directory with main.go and go.mod files."
}

variable "codebase_rebuild_trigger" {
  type = object({
    base         = string
    modules_dirs = list(string)
  })
  description = "Location of the codebase which change will trigger lambda rebuild and redeployment."
  default     = null
}

variable "policy_arn_list" {
  type        = list(string)
  description = "Custom policies to attache to the Lambda execution role."
  validation {
    condition     = length(var.policy_arn_list) < 19
    error_message = "Number of custom policies cannot exceed 19"
  }
  default = []
}

variable "env_vars" {
  type        = map(string)
  description = "Key-value pairs to set environemnt variables."
  default     = {}
}

variable "lambda_memory" {
  type        = number
  description = "Lambda RAM in Mb."
  default     = 256
}

variable "exec_timeout_sec" {
  type        = number
  description = "Lambda invocation timeout in sec."
  default     = 30
}

variable "tags" {
  type        = map(string)
  description = "Resource tags."
  default     = {}
}
