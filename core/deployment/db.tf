resource "neon_project" "this" {
  name = "diagramastext"
}

resource "neon_branch" "this" {
  project_id = neon_project.this.id
  name       = "master"
}

resource "neon_role" "admin" {
  project_id = neon_project.this.id
  branch_id  = neon_branch.this.id
  name       = "admin"
}

resource "neon_database" "this" {
  project_id = neon_project.this.id
  branch_id  = neon_branch.this.id
  name       = "core"
  owner_name = neon_role.admin.name
}

resource "neon_role" "lambda" {
  project_id = neon_project.this.id
  branch_id  = neon_branch.this.id
  name       = "lambda"
}

resource "aws_secretsmanager_secret" "neon_lambda" {
  name                    = "neon/master/core/lambda"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "neon_lambda" {
  secret_id = aws_secretsmanager_secret.neon_lambda.id
  secret_string = jsonencode({
    project_id = neon_project.this.id
    branch_id  = neon_branch.this.id
    dbname     = neon_database.this.name
    host       = neon_branch.this.host
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
  name   = "${neon_branch.this.name}-${neon_database.this.name}-${neon_role.lambda.name}"
  path   = "/neon/read-only/"
  policy = data.aws_iam_policy_document.neon_lambda.json
}

resource "aws_iam_role_policy_attachment" "neon_lambda" {
  policy_arn = aws_iam_policy.neon_lambda.arn
  role       = aws_iam_role.lambda_core.name
}
