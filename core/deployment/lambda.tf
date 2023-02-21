locals {
  lambda = "core${local.suffix}"
  lambda_settings_prod = {
    true = {
      secret_arn = "arn:aws:secretsmanager:us-east-2:027889758114:secret:production/core-MDqc82"
    }
    false = {
      secret_arn = "arn:aws:secretsmanager:us-east-2:027889758114:secret:staging/core-MDqc82"
    }
  }
}

resource "aws_iam_role" "lambda_core" {
  name = "Lambda${local.suffix}"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      },
    ]
  })
}

data "aws_iam_policy_document" "lambda_core" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }

  statement {
    effect    = "Allow"
    actions   = ["secretsmanager:ListSecrets"]
    resources = ["*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "secretsmanager:GetResourcePolicy",
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
      "secretsmanager:ListSecretVersionIds",
    ]
    resources = [
      local.lambda_settings_prod[local.is_prod]["secret_arn"]
    ]
  }
}

resource "aws_iam_policy" "lambda_core" {
  name   = "LambdaCore${local.suffix}"
  policy = data.aws_iam_policy_document.lambda_core.json
}

resource "aws_iam_role_policy_attachment" "lambda_core" {
  policy_arn = aws_iam_policy.lambda_core.arn
  role       = aws_iam_role.lambda_core.name
}

resource "aws_cloudwatch_log_group" "lambda_core" {
  name              = "/aws/lambda/${local.lambda}"
  retention_in_days = 7
}

locals {
  codebase_md5 = base64sha256(join(",", [
    for file in concat(
      [for f in fileset("${path.module}/../", "{*.go,go.mod,go.sum}") : "${path.module}/../${f}"],
      [for f in fileset("${path.module}/../compression", "*.go") : "${path.module}/../compression/${f}"],
      [for f in fileset("${path.module}/../handler", "*.go") : "${path.module}/../handler/${f}"],
      [for f in fileset("${path.module}/../storage", "{*.go,go.mod,go.sum}") : "${path.module}/../storage/${f}"],
      [for f in fileset("${path.module}/../secretsmanager", "{*.go,go.mod,go.sum}") : "${path.module}/../secretsmanager/${f}"],
      [for f in fileset("${path.module}/../cmd/lambda", "{*.go,go.mod,go.sum}") : "${path.module}/../cmd/lambda/${f}"],
    ) : filemd5(file)
  ]))
  archive_name = "${local.lambda}-${local.codebase_md5}"
}

resource "null_resource" "lambda_core" {
  triggers = {
    md5 = local.codebase_md5
  }

  provisioner "local-exec" {
    command = "cd ${path.module}/.. && make build ZIPNAME=${local.archive_name}"
  }
}

resource "aws_lambda_function" "core" {
  function_name = local.lambda
  role          = aws_iam_role.lambda_core.arn

  filename = "${path.module}/../bin/lambda.zip"
  #  source_code_hash = null_resource.lambda_core.triggers.md5
  runtime     = "go1.x"
  handler     = "lambda"
  memory_size = 256
  timeout     = 120

  environment {
    variables = {
      ACCESS_CREDENTIALS_ARN = local.lambda_settings_prod[local.is_prod]["secret_arn"]
      OPENAI_MODEL           = var.openai_model
      OPENAI_MAX_TOKENS      = var.openai_max_tokens
      OPENAI_TEMPERATURE     = var.openai_temperature
      CORS_HEADERS           = jsonencode(local.cors_headers)
    }
  }

  depends_on = [null_resource.lambda_core]
}
