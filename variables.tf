variable "repos_to_grant_permission" {
  type = list
  description = "A list of ECR repo names, if set permission will be granted for the lambda created to apply tags to this repo, and read the image manifests for the listed repos. If unset permission will be granted to all repos in the account"
  default = []
}

variable "name" {
  type = string
  description = "Name to give to all created resources"
}

variable "tags" {
  type = map
  description = "Additional tags to add to all taggable resources created"
  default = {}
}

variable "environment" {
  type = string
  description = "Deployment environment (e.g. prod, test, dev)"
}
