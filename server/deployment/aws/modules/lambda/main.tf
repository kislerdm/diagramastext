resource "aws_iam_role" "this" {
  name = "LambdaRole${var.name}"
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
  tags = var.tags
}

data "aws_iam_policy_document" "logs" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["arn:aws:logs:region:*:log-group:/aws/lambda/${var.name}:*"]
  }
}

resource "aws_iam_policy" "logs" {
  name   = "LambdaPolicyLogs-${var.name}"
  policy = data.aws_iam_policy_document.logs.json
  tags   = var.tags
}

resource "aws_iam_role_policy_attachment" "logs" {
  policy_arn = aws_iam_policy.logs.arn
  role       = aws_iam_role.this.name
}

resource "aws_iam_role_policy_attachment" "custom" {
  for_each   = toset(var.policy_arn_list)
  policy_arn = each.value
  role       = aws_iam_role.this.name
}

resource "aws_cloudwatch_log_group" "logs" {
  name              = "/aws/lambda/${var.name}"
  retention_in_days = 7
  tags              = var.tags
}

locals {
  files_list_lambda_module = [
    for f in fileset(var.path_lambda_module, "{*.go,go.mod,go.sum}") : "${var.path_lambda_module}/${f}"
  ]

  files_list_dependencies = var.codebase_rebuild_trigger == null ? [] : flatten(
    [
      for pattern in var.codebase_rebuild_trigger.modules_dir_patterns :
      [
        for f in fileset(var.codebase_rebuild_trigger.base, "${pattern}{*.go,go.mod,go.sum}") :
        "${var.codebase_rebuild_trigger.base}/${f}"
      ]
    ]
  )

  codebase_md5 = md5(
    join(",", [for file in concat(local.files_list_lambda_module, local.files_list_dependencies) : filemd5(file)])
  )

  archive_name = "${var.name}-${local.codebase_md5}.zip"
  dir_module   = abspath(path.module)
}

resource "null_resource" "this" {
  triggers = {
    md5  = local.codebase_md5
    name = local.archive_name
  }

  provisioner "local-exec" {
    command = "cd ${local.dir_module} && make build ZIPNAME=${local.archive_name} CODE_PATH=${var.path_lambda_module}"
  }
}

resource "aws_lambda_function" "this" {
  function_name = var.name
  role          = aws_iam_role.this.arn

  filename = "${local.dir_module}/bin/${local.archive_name}"
  runtime  = "go1.x"
  handler  = "lambda"

  memory_size = var.lambda_memory
  timeout     = var.exec_timeout_sec

  dynamic "environment" {
    for_each = length(var.env_vars) == 0 ? {} : { foo = null }
    content {
      variables = var.env_vars
    }
  }

  tags = var.tags

  depends_on = [null_resource.this]
}
