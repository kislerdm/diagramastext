locals {
  core_rendering_c4_lambda_settings = {
    secret_arn_is_prod = {
      true  = "arn:aws:secretsmanager:us-east-2:027889758114:secret:production/core-MDqc82"
      false = "arn:aws:secretsmanager:us-east-2:027889758114:secret:staging/core-MDqc82"
    }
  }
}

data "aws_iam_policy_document" "core_rendering_c4_secret" {
  statement {
    effect    = "Allow"
    actions   = ["secretsmanager:ListSecrets"]
    resources = ["*"]
  }

  statement {
    effect  = "Allow"
    actions = [
      "secretsmanager:GetResourcePolicy",
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
      "secretsmanager:ListSecretVersionIds",
    ]
    resources = [
      local.core_rendering_c4_lambda_settings.secret_arn_is_prod[local.is_prod]
    ]
  }
}

resource "aws_iam_policy" "core_rendering_c4_secret" {
  name   = "LambdaCoreRenderingC4${local.suffix}"
  policy = data.aws_iam_policy_document.core_rendering_c4_secret.json
}

module "core_rendering_c4" {
  source                   = "./modules/lambda"
  name                     = "core-rendering-c4${local.suffix}"
  path_lambda_module       = "${abspath(path.module)}/../cmd/lambda/core-c4"
  codebase_rebuild_trigger = {
    base                 = abspath("${path.module}/..")
    modules_dir_patterns = [
      "",
      "{errors,openai,secretsmanager,storage,utils}/**",
      "c4container/**",
      "cmd/lambda/core-c4",
    ]
  }
  policy_arn_list = [aws_iam_policy.core_rendering_c4_secret.arn]
  env_vars        = {
    ACCESS_CREDENTIALS_ARN = local.core_rendering_c4_lambda_settings.secret_arn_is_prod[local.is_prod]
    CORS_HEADERS           = jsonencode(local.cors_headers)
  }
  tags = {
    environment    = local.environment
    system         = "core"
    backend        = "c4containers"
  }
  depends_on = [aws_iam_policy.core_rendering_c4_secret]
}
