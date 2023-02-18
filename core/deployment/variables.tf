variable "openai_api_key" {
  description = "OpenAI API Key"
  type        = string
}

variable "openai_max_tokens" {
  description = <<EOT
Max tokens.
See details: https://platform.openai.com/docs/api-reference/completions/create#completions/create-max_tokens
EOT

  type    = number
  default = 768
}

variable "openai_temperature" {
  description = <<EOT
What sampling temperature to use.
See details: https://platform.openai.com/docs/api-reference/completions/create#completions/create-temperature
EOT

  type    = number
  default = 0
}

variable "environment" {
  type        = string
  description = "Deployment environment."
  validation {
    condition     = var.environment == "production" || var.environment == "staging"
    error_message = "'production' and 'staging' are the only allowed values"
  }
}

variable "neon_password" {
  type        = string
  description = "(Temp!) Database access password."
}
