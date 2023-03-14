resource "aws_secretsmanager_secret" "core" {
  for_each                = toset(["production", "staging"])
  name                    = "${each.key}/core"
  recovery_window_in_days = 0
}
