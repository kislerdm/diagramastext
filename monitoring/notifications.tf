resource "aws_sns_topic" "this" {
  name   = "core-monitoring"
  policy = <<EOF
{
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "cloudwatch.amazonaws.com"
        },
        "Action" : "SNS:Publish",
        "Resource" : "arn:aws:sns:${local.region}:${local.account_id}:core-monitoring",
        "Condition" : {
          "ArnLike" : {
            "aws:SourceArn" : "arn:aws:cloudwatch:${local.region}:${local.account_id}:alarm:*"
          },
          "StringEquals" : {
            "aws:SourceAccount" : "${local.account_id}"
          }
        }
      }
    ]
  }
EOF
}
