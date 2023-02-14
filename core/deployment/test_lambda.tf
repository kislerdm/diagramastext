resource "aws_cloudwatch_log_group" "test" {
  name              = "/aws/lambda/test-lambda"
  retention_in_days = 1
}

data "local_file" "test" {
  filename   = "${path.module}/../bin/test.zip"
}

resource "aws_lambda_function" "test" {
  function_name = "test-lambda"
  role          = aws_iam_role.lambda_core.arn

  filename         = data.local_file.test.filename
  source_code_hash = base64sha256(data.local_file.test.content_base64)
  runtime          = "go1.x"
  handler          = "lambda"
  memory_size      = 128
  timeout          = 10
}
