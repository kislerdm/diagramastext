variable "project" {
  type        = string
  description = "Project ID."
}

variable "imagetag" {
  type        = string
  description = "Docker image tag."
}

variable "cors_headers" {
  type        = map(string)
  description = "CORS headers map."
  default     = null
}

variable "api_domain" {
  type        = string
  description = "API domain to map the Cloud Run to."
  default     = ""
}

variable "location" {
  type        = string
  description = "Geo-location."
  default     = "us-central1"
}

variable "model_max_tokens" {
  type        = number
  description = "https://platform.openai.com/docs/api-reference/chat/create#chat/create-max_tokens"
  default     = 500
  validation {
    condition     = var.model_max_tokens > 0 && var.model_max_tokens < 1000
    error_message = "'model_max_tokens' must be between 0 and 1000"
  }
}
