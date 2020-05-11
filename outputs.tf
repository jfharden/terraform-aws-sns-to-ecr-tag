output "sns_topic_arn" {
  description = "ARN of the SNS topic to trigger the tags"
  value = ""
}

output "lambda_function_arn" {
  description = "ARN of the lambda function which performs the tagging"
  value = ""
}

output "dead_letter_queue_arn" {
  description = "ARN of the dead letter queue for the SNS topic"
  value = ""
}
