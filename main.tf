/**
* # terraform-aws-sns-to-ecr-tag
* Terraform module which allows ECR images to have tags added by sending a message to an SNS topic
*
* The following directories are in the github repo:
*
* * /: The terraform module itself in the root of the project
* * examples: Fully functional terraform examples using this module
* * src: Python code for the lambda function
* * test: Tests for the module written in go using terratest and deploying the examples from the examples directory
*
* ## SNS Topic Body
*
* You can send a JSON message to the created SNS topic with the following fields:
*
* | name              | description                                 |
* |-------------------|---------------------------------------------|
* | ecr_repo_name     | The name of the ECR repo with the image in  |
* | ecr_tag_to_update | The tag for the image to update             |
* | ecr_tag_to_add    | The tag to add to the image                 |
*
* Example SNS message body:
* ```
* {
*   "ecr_repo_name": "my_wonderful_repository",
*   "ecr_tag_to_update": "1.2",
*   "ecr_tag_to_add": "deployed_on_20200511T2321Z"
* }
* ```
*
* ## How this works
*
* The SNS topic will trigger a lambda function, that lambda function adds a tag to the image (without having to pull or
* push the whole image (see the guide https://docs.aws.amazon.com/AmazonECR/latest/userguide/image-retag.html).
*
* This does require the docker images to made with Docker image Manifest V2 Schema 2.
*
* If the lambda fails it will deliver the failures to the dead letter queue. I would also like to send the failed sns
* messages to the DLQ but terraform support is lacking for that at the moment (see 
* https://github.com/terraform-providers/terraform-provider-aws/issues/10931 )
*/

locals {
  tags = merge(
    { "Environment" : var.environment },
    var.tags,
  )
}

resource "aws_sqs_queue" "dead_letter" {
  name = "${var.name}-dead-letter"

  tags = local.tags
}

resource "aws_sns_topic" "this" {
  name = var.name

  tags = local.tags
}

module "lambda" {
  source = "git::https://github.com/claranet/terraform-aws-lambda.git//?ref=v1.2.0"

  function_name = var.name
  description   = "Add tag to an ECR image"
  handler       = "lambdas.tag_ecr_image.tag_ecr_image"
  runtime       = "python3.8"

  source_path = "${path.module}/src"

  policy = {
    json = data.aws_iam_policy_document.lambda.json
  }

  dead_letter_config = {
    target_arn = aws_sqs_queue.dead_letter.arn
  }

  tags = local.tags
}

data "aws_iam_policy_document" "lambda" {
  statement {
    actions = [
      "ecr:BatchGetImage",
      "ecr:PutImage",
    ]

    resources = length(var.repos_to_grant_permission) == 0 ? ["*"] : var.repos_to_grant_permission
  }
}

resource "aws_lambda_permission" "with_sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.this.arn
}

resource "aws_sns_topic_subscription" "lambda" {
  topic_arn = aws_sns_topic.this.arn
  protocol  = "lambda"
  endpoint  = module.lambda.function_arn
}
