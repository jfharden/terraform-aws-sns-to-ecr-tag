# terraform-aws-sns-to-ecr-tag  
Terraform module which allows ECR images to have tags added by sending a message to an SNS topic

The following directories are in the github repo:

* /: The terraform module itself in the root of the project
* examples: Fully functional terraform examples using this module
* src: Python code for the lambda function
* test: Tests for the module written in go using terratest and deploying the examples from the examples directory

## SNS Topic Body

You can send a JSON message to the created SNS topic with the following fields:

| name              | description                                 |
|-------------------|---------------------------------------------|
| ecr\_repo\_name     | The name of the ECR repo with the image in  |
| ecr\_tag\_to\_update | The tag for the image to update             |
| ecr\_tag\_to\_add    | The tag to add to the image                 |

Example SNS message body:
```
{
  "ecr_repo_name": "my_wonderful_repository",
  "ecr_tag_to_update": "1.2",
  "ecr_tag_to_add": "deployed_on_20200511T2321Z"
}
```

## How this works

The SNS topic will trigger a lambda function, that lambda function adds a tag to the image (without having to pull or  
push the whole image (see the guide https://docs.aws.amazon.com/AmazonECR/latest/userguide/image-retag.html).

This does require the docker images to made with Docker image Manifest V2 Schema 2.

If the lambda fails it will deliver the failures to the dead letter queue. I would also like to send the failed sns  
messages to the DLQ but terraform support is lacking for that at the moment (see  
https://github.com/terraform-providers/terraform-provider-aws/issues/10931 )

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| environment | Deployment environment (e.g. prod, test, dev) | `string` | n/a | yes |
| name | Name to give to all created resources | `string` | n/a | yes |
| repos\_to\_grant\_permission | A list of ECR repo arns, if set permission will be granted for the lambda created to apply tags to this repo, and read the image manifests for the listed repos. If unset permission will be granted to all repos in the account | `list` | `[]` | no |
| tags | Additional tags to add to all taggable resources created | `map` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| dead\_letter\_queue\_arn | ARN of the dead letter queue for the SNS topic |
| lambda\_function\_arn | ARN of the lambda function which performs the tagging |
| sns\_topic\_arn | ARN of the SNS topic to trigger the tags |

