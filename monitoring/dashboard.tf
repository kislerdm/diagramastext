locals {
  gw = {
    name = "main"
    stage = "production"
  }
  lambda_name = "core-c4"
}

resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "core"
  dashboard_body = <<EOF
{
    "widgets": [
        {
            "height": 7,
            "width": 9,
            "y": 1,
            "x": 0,
            "type": "metric",
            "properties": {
                "metrics": [
                    [ { "expression": "errors_5xx/total", "label": "5xx Over Total Requests", "id": "ratio_5xx", "period": 900, "region": "us-east-2" } ],
                    [ { "expression": "errirs_4xx/total", "label": "4xx Over Total Requests", "id": "ratio_4xx", "region": "us-east-2" } ],
                    [ "AWS/ApiGateway", "Count", "ApiName", "${local.gw.name}", { "id": "total", "yAxis": "right", "label": "Total requests" } ],
                    [ ".", "5XXError", ".", ".", { "id": "errors_5xx", "visible": false } ],
                    [ ".", "4XXError", ".", ".", "Stage", "${local.gw.stage}", { "id": "errirs_4xx", "visible": false } ]
                ],
                "view": "timeSeries",
                "stacked": false,
                "region": "us-east-2",
                "stat": "Sum",
                "period": 900,
                "title": "5xx_rate"
            }
        },
        {
            "height": 7,
            "width": 5,
            "y": 1,
            "x": 9,
            "type": "metric",
            "properties": {
                "metrics": [
                    [ "AWS/ApiGateway", "Count", "ApiName", "${local.gw.name}", { "label": "Total Input Prompts", "id": "total" } ],
                    [ { "expression": "total - error_5xx - error_4xx", "label": "Total Successfully Generated Diagrams", "id": "total_success", "region": "us-east-2", "period": 300 } ],
                    [ "AWS/ApiGateway", "5XXError", "ApiName", "${local.gw.name}", { "id": "error_5xx", "visible": false } ],
                    [ ".", "4XXError", ".", ".", { "id": "error_4xx", "visible": false } ]
                ],
                "sparkline": false,
                "view": "singleValue",
                "region": "us-east-2",
                "stat": "Sum",
                "period": 300,
                "title": "Cumulative Stats",
                "stacked": true,
                "liveData": false,
                "setPeriodToTimeRange": true,
                "trend": false
            }
        },
        {
            "height": 10,
            "width": 9,
            "y": 9,
            "x": 0,
            "type": "metric",
            "properties": {
                "metrics": [
                    [ "AWS/Lambda", "Duration", "FunctionName", "${local.lambda_name}", { "label": "Duration" } ]
                ],
                "view": "timeSeries",
                "stacked": false,
                "region": "us-east-2",
                "stat": "p99",
                "period": 300,
                "title": "Core execution duration"
            }
        },
        {
            "height": 1,
            "width": 14,
            "y": 8,
            "x": 0,
            "type": "text",
            "properties": {
                "markdown": "## Lambda"
            }
        },
        {
            "height": 1,
            "width": 9,
            "y": 0,
            "x": 0,
            "type": "text",
            "properties": {
                "markdown": "## Gateway"
            }
        },
        {
            "height": 1,
            "width": 5,
            "y": 0,
            "x": 9,
            "type": "text",
            "properties": {
                "markdown": "## Usage Stats"
            }
        },
        {
            "height": 5,
            "width": 5,
            "y": 9,
            "x": 9,
            "type": "metric",
            "properties": {
                "metrics": [
                    [ "AWS/Lambda", "Duration", "FunctionName", "${local.lambda_name}", { "label": "Request-Response Latency", "color": "#2ca02c" } ]
                ],
                "view": "gauge",
                "region": "us-east-2",
                "stat": "p99",
                "period": 300,
                "yAxis": {
                    "left": {
                        "min": 100,
                        "max": 15000
                    }
                },
                "setPeriodToTimeRange": false,
                "sparkline": true,
                "trend": true,
                "liveData": false,
                "legend": {
                    "position": "hidden"
                },
                "title": "Latency:p99"
            }
        },
        {
            "height": 5,
            "width": 5,
            "y": 14,
            "x": 9,
            "type": "metric",
            "properties": {
                "metrics": [
                    [ "AWS/Lambda", "Throttles", "FunctionName", "${local.lambda_name}" ]
                ],
                "view": "gauge",
                "region": "us-east-2",
                "stat": "Maximum",
                "period": 60,
                "yAxis": {
                    "left": {
                        "min": 0,
                        "max": 20
                    }
                },
                "setPeriodToTimeRange": false,
                "sparkline": true,
                "trend": true,
                "liveData": false,
                "legend": {
                    "position": "hidden"
                },
                "title": "Throttling"
            }
        }
    ]
}
EOF
}