variable "name" {
  type        = string
  description = "Name to give to all created resources"
}

variable "tags" {
  type        = map
  description = "Additional tags to add to all taggable resources created"
  default     = {}
}
