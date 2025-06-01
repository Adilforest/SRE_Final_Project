variable "replica_count" {
  description = "Number of API Gateway replicas"
  type        = number
  default     = 3
}

variable "docker_image" {
  description = "Docker image for API Gateway"
  type        = string
  default     = "api-gateway:latest"
}