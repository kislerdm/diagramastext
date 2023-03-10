variable "openai_model" {
  description = "OpenAI Model ID/Name."
  type        = string
  default     = "code-davinci-002"
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
