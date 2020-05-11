/**
* # terraform-aws-sns-to-ecr-tag
* Terraform module which allows ECR images to have tags added by sending a message to an SNS topic
*
* The following directories are in the github repo:
*
* * /: The terraform module itself in the root of the project
* * examples: Fully functional terraform examples using this module
* * src: Python code for the lambda function
* * tests: Tests for the module written in go using terratest
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
*
* ## How this works
*
* The SNS topic will trigger a lambda function, that lambda function adds a tag to the image (without having to pull or
* push the whole image (see the guide https://docs.aws.amazon.com/AmazonECR/latest/userguide/image-retag.html).
*
* This does require the docker images to made with Docker image Manifest V2 Schema 2.
*
* If this ever fails SNS will keep retrying for 24 hours, after which time it will deliver the message body to an SQS
* dead letter queue. If you wish to monitor this you should setup some alarms on the dead letter queue, and maybe also
* on the number of failed delivery attempts on the SNS topic
* (see https://docs.aws.amazon.com/sns/latest/dg/sns-monitoring-using-cloudwatch.html).
*
*/
