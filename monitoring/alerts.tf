locals {
  aggregation_period = {
    request_error_rate = "600"
  }
}

resource "aws_cloudwatch_metric_alarm" "request_error_rate" {
  alarm_name          = "core-response-error-rate"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = "1"
  threshold           = "10"
  alarm_description   = "Request error rate has exceeded 10%"
  treat_missing_data  = "ignore"

  metric_query {
    id          = "error_rate"
    expression  = "total_errors/total_requests*100"
    label       = "Error Rate"
    return_data = "true"
  }

  metric_query {
    id          = "total_errors"
    expression  = "total_5xx+total_4xx"
    label       = "Total Errors"
    return_data = "false"
  }

  metric_query {
    id = "total_requests"

    metric {
      metric_name = "Count"
      namespace   = "AWS/ApiGateway"
      period      = local.aggregation_period.request_error_rate
      stat        = "Sum"
      unit        = "Count"

      dimensions = {
        ApiName = local.gw.name
      }
    }
  }

  metric_query {
    id = "total_5xx"

    metric {
      metric_name = "5XXError"
      namespace   = "AWS/ApiGateway"
      period      = local.aggregation_period.request_error_rate
      stat        = "Sum"
      unit        = "Count"

      dimensions = {
        ApiName = local.gw.name
      }
    }
  }

  metric_query {
    id = "total_4xx"

    metric {
      metric_name = "4XXError"
      namespace   = "AWS/ApiGateway"
      period      = local.aggregation_period.request_error_rate
      stat        = "Sum"
      unit        = "Count"

      dimensions = {
        ApiName = local.gw.name
      }
    }
  }
}
