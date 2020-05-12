module "example" {
  source = "../../"

  name        = var.name
  environment = "test"
  tags        = var.tags
}

resource "aws_ecr_repository" "example" {
  name = var.name
}
