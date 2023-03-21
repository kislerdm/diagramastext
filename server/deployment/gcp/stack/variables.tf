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
