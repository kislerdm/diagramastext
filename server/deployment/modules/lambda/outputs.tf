output "function_name" {
  value       = aws_lambda_function.this.function_name
  description = "Unique name for your Lambda Function."
}

output "arn" {
  value       = aws_lambda_function.this.arn
  description = "Amazon Resource Name (ARN) identifying your Lambda Function."
}

output "invoke_arn" {
  value       = aws_lambda_function.this.invoke_arn
  description = "ARN to be used for invoking Lambda Function from API Gateway."
}
