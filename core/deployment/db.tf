resource "neon_project" "this" {
  name = "diagramastext"
}

locals {
  neon_branch_id  = "br-steep-silence-472824"
  neon_owner_name = "diagramastext"
  neon_endpoint   = "ep-wild-wind-389577.us-east-2.aws.neon.tech"
}

resource "neon_database" "this" {
  name       = "core"
  project_id = neon_project.this.id
  branch_id  = local.neon_branch_id
  owner_name = local.neon_owner_name
}

resource "neon_role" "lambda" {
  name       = "lambda"
  project_id = neon_project.this.id
  branch_id  = local.neon_branch_id
}

resource "aws_secretsmanager_secret" "neon_lambda" {
  name                    = "neon/main/core/lambda"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "neon_lambda" {
  secret_id = aws_secretsmanager_secret.neon_lambda.id
  secret_string = jsonencode({
    branch_id  = local.neon_branch_id
    host       = local.neon_endpoint
    project_id = neon_project.this.id
    dbname     = neon_database.this.name
    user       = neon_role.lambda.name
    password   = neon_role.lambda.password
  })
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
      aws_secretsmanager_secret_version.neon_lambda.arn,
    ]
  }
}

resource "aws_iam_policy" "neon_lambda" {
  name   = "main-core-lambda"
  path   = "/neon/read-only/"
  policy = data.aws_iam_policy_document.neon_lambda.json
}

resource "aws_iam_role_policy_attachment" "neon_lambda" {
  policy_arn = aws_iam_policy.neon_lambda.arn
  role       = aws_iam_role.lambda_core.name
}
