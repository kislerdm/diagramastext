locals {
  lambda_c4 = "core-c4-${local.suffix}"
}

resource "aws_iam_policy" "lambda_core" {
  name = "LambdaCore"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = concat([
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ]
        Resource = ["arn:aws:logs:*:*:*"]
      },
      ]
    )
  })
}

resource "aws_iam_role" "lambda_core" {
  name = "Lambda"
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

resource "aws_iam_role_policy_attachment" "core_c4" {
  policy_arn = aws_iam_policy.lambda_core.arn
  role       = aws_iam_role.lambda_core.name
}

resource "aws_cloudwatch_log_group" "core_c4" {
  name              = "/aws/lambda/${local.lambda_c4}"
  retention_in_days = 7
}

resource "null_resource" "core_c4" {
  triggers = {
    md5 = base64sha256(join(",", [
      for file in concat(
        [for f in fileset("${path.module}/../", "{*.go,go.mod,go.sum}") : "${path.module}/../${f}"],
        [for f in fileset("${path.module}/../compression", "*.go") : "${path.module}/../compression/${f}"],
        [for f in fileset("${path.module}/../handler", "*.go") : "${path.module}/../handler/${f}"],
        [for f in fileset("${path.module}/../storage", "{*.go,go.mod,go.sum}") : "${path.module}/../storage/${f}"],
        [for f in fileset("${path.module}/../cmd/lambda", "{*.go,go.mod,go.sum}") : "${path.module}/../cmd/lambda/${f}"],
      ) : filemd5(file)
    ]))
  }

  provisioner "local-exec" {
    command = "cd ${path.module}/.. && make build"
  }
}

data "local_file" "core_c4" {
  filename   = "${path.module}/../bin/lambda.zip"
  depends_on = [null_resource.core_c4]
}

resource "aws_lambda_function" "core_c4" {
  function_name = local.lambda_c4
  role          = aws_iam_role.lambda_core.arn

  filename         = data.local_file.core_c4.filename
  source_code_hash = null_resource.core_c4.triggers.md5
  runtime          = "go1.x"
  handler          = "lambda"
  memory_size      = 256
  timeout          = 120

  environment {
    variables = {
      OPENAI_API_KEY     = var.openai_api_key
      OPENAI_MAX_TOKENS  = var.openai_max_tokens
      OPENAI_TEMPERATURE = var.openai_temperature
      CORS_HEADERS       = jsonencode(local.cors_headers)
      NEON_DBNAME        = "core"
      NEON_USER          = "lambda"
      NEON_HOST          = local.neon_project[var.environment]["endpoint"]
      NEON_PASSWORD      = var.neon_password
    }
  }

  depends_on = [null_resource.core_c4]
}

locals {
  neon_db = {
    production = {
      endpoint = "ep-wild-wind-389577.us-east-2.aws.neon.tech"
    }
  }
  lambda_secret = {
    production = "arn:aws:secretsmanager:us-east-2:027889758114:secret:neon/main/core/lambda-C335bP"
  }
}

data "aws_iam_policy_document" "neon_lambda" {
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
      ""
    ]
  }
}

resource "aws_iam_policy" "neon_lambda" {
  name   = "main-core-lambda-${local.suffix}"
  path   = "/neon/read-only/"
  policy = data.aws_iam_policy_document.neon_lambda.json
}

resource "aws_iam_role_policy_attachment" "neon_lambda" {
  policy_arn = aws_iam_policy.neon_lambda.arn
  role       = aws_iam_role.lambda_core.name
}
