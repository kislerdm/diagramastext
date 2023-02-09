resource "aws_api_gateway_account" "this" {
  cloudwatch_role_arn = aws_iam_role.cloudwatch.arn
}

resource "aws_iam_role" "cloudwatch" {
  name = "api_gateway_cloudwatch_global"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "cloudwatch" {
  name = "GWCloudwatch"
  role = aws_iam_role.cloudwatch.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:DescribeLogGroups",
                "logs:DescribeLogStreams",
                "logs:PutLogEvents",
                "logs:GetLogEvents",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
EOF
}

resource "aws_api_gateway_request_validator" "this" {
  name                        = "main-validator"
  rest_api_id                 = aws_api_gateway_rest_api.this.id
  validate_request_body       = true
  validate_request_parameters = true
}

resource "aws_api_gateway_rest_api" "this" {
  name           = "main"
  api_key_source = "HEADER"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

resource "aws_api_gateway_model" "schema_request" {
  rest_api_id  = aws_api_gateway_rest_api.this.id
  name         = "UserInputPrompt"
  content_type = "application/json"
  schema       = <<EOF
{
  "type": "object",
  "required": ["prompt"],
  "additionalProperties": false,
  "properties": {
      "prompt": {
          "type": "string",
          "minLength": 3,
          "maxLength": 768
      }
  }
}
EOF
}

resource "aws_api_gateway_model" "schema_response" {
  rest_api_id  = aws_api_gateway_rest_api.this.id
  name         = "SVGResp"
  content_type = "application/json"
  schema       = <<EOF
{
  "type": "object",
  "required": ["svg"],
  "additionalProperties": false,
  "properties": {
      "svg": {
          "type": "string"
      }
  }
}
EOF
}

resource "aws_api_gateway_gateway_response" "response-401" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  status_code   = "401"
  response_type = "UNAUTHORIZED"

  response_templates = {
    "application/json" = "{\"error\":$context.error.messageString}"
  }

  response_parameters = {
    "gatewayresponse.header.Access-Control-Allow-Origin" = "'*'"
  }
}

locals {
  allowed_headers_response = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
  }

  cors_headers = {
    "method.response.header.Access-Control-Allow-Origin"  = "'*'"
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token'"
    "method.response.header.Access-Control-Allow-Methods" = "'POST,OPTIONS'"
  }

  endpoints = {
    "c4" : ["POST"],
  }

  request_parameters = merge({
    # "method.request.header.UserID" = true
    # "method.request.header.Authorization" = true
    }
  )

  deployment_trigger = merge(
    local.endpoints,
    local.allowed_headers_response,
    local.cors_headers,
    local.request_parameters,
  )

  lambda_trigger = {
    for i in flatten([
      for k, v in local.endpoints : [
        for i in v : {
          id     = "${k}-${i}"
          path   = k
          method = i
        }
      ]
    ]) :
    i.id => {
      method = i.method
      path   = i.path
    }
  }
}

resource "aws_api_gateway_resource" "route_top" {
  for_each    = local.endpoints
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_rest_api.this.root_resource_id
  path_part   = each.key
}

resource "aws_lambda_permission" "gw" {
  for_each      = local.lambda_trigger
  statement_id  = "InvokeGWMain-${each.key}"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.core_c4.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.this.execution_arn}/*/${each.value.method}/${each.value.path}"
}

resource "aws_api_gateway_method" "options" {
  for_each           = local.endpoints
  rest_api_id        = aws_api_gateway_rest_api.this.id
  resource_id        = aws_api_gateway_resource.route_top[each.key].id
  http_method        = "OPTIONS"
  authorization      = "NONE"
  request_parameters = local.request_parameters
}

resource "aws_api_gateway_integration" "options" {
  for_each    = local.endpoints
  rest_api_id = aws_api_gateway_method.options[each.key].rest_api_id
  resource_id = aws_api_gateway_method.options[each.key].resource_id
  http_method = "OPTIONS"
  type        = "MOCK"
  request_templates = {
    "application/json" = "{ \"statusCode\": 200 }"
  }
}

resource "aws_api_gateway_method_response" "options" {
  for_each    = local.endpoints
  rest_api_id = aws_api_gateway_method.options[each.key].rest_api_id
  resource_id = aws_api_gateway_method.options[each.key].resource_id
  http_method = "OPTIONS"
  status_code = "200"
  response_models = {
    "application/json" = "Empty"
  }
  response_parameters = local.allowed_headers_response
}

resource "aws_api_gateway_integration_response" "options" {
  for_each    = local.endpoints
  rest_api_id = aws_api_gateway_method.options[each.key].rest_api_id
  resource_id = aws_api_gateway_method.options[each.key].resource_id
  http_method = "OPTIONS"
  status_code = "200"

  response_templates = {
    "application/json" = "{ \"statusCode\": 200 }"
  }
  response_parameters = local.cors_headers
}

resource "aws_api_gateway_method" "this" {
  for_each      = local.lambda_trigger
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.route_top[each.value.path].id
  http_method   = each.value.method
  authorization = "NONE"

  api_key_required = true

  request_models = {
    "application/json" = aws_api_gateway_model.schema_request.name
  }
  request_validator_id = aws_api_gateway_request_validator.this.id

  request_parameters = local.request_parameters
}

resource "aws_api_gateway_integration" "this" {
  for_each                = local.lambda_trigger
  rest_api_id             = aws_api_gateway_method.this[each.key].rest_api_id
  resource_id             = aws_api_gateway_method.this[each.key].resource_id
  http_method             = aws_api_gateway_method.this[each.key].http_method
  integration_http_method = aws_api_gateway_method.this[each.key].http_method
  type                    = "AWS_PROXY"
  content_handling        = "CONVERT_TO_TEXT"
  uri                     = aws_lambda_function.core_c4.invoke_arn

  request_templates = {
    "application/json" = aws_api_gateway_model.schema_request.name
  }
}

resource "aws_api_gateway_method_response" "this" {
  for_each    = local.lambda_trigger
  rest_api_id = aws_api_gateway_method.this[each.key].rest_api_id
  resource_id = aws_api_gateway_method.this[each.key].resource_id
  http_method = aws_api_gateway_method.this[each.key].http_method
  status_code = "200"
  response_models = {
    "application/json" = aws_api_gateway_model.schema_response.name
  }
  response_parameters = local.allowed_headers_response
}

resource "aws_api_gateway_integration_response" "this" {
  for_each            = local.lambda_trigger
  rest_api_id         = aws_api_gateway_method.this[each.key].rest_api_id
  resource_id         = aws_api_gateway_method.this[each.key].resource_id
  http_method         = aws_api_gateway_method.this[each.key].http_method
  status_code         = "200"
  response_parameters = local.cors_headers
  response_templates = {
    "application/json" = aws_api_gateway_model.schema_response.name
  }
}

# stage and deployment

locals {
  stages = ["production"]
}

resource "aws_cloudwatch_log_group" "gw" {
  for_each          = toset(local.stages)
  name              = "API-Gateway-Execution-Logs_${aws_api_gateway_rest_api.this.id}/${each.value}"
  retention_in_days = 7
}

resource "aws_api_gateway_deployment" "this" {
  rest_api_id = aws_api_gateway_rest_api.this.id

  triggers = {
    redeployment = sha1(jsonencode(local.deployment_trigger))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "this" {
  for_each      = toset(local.stages)
  deployment_id = aws_api_gateway_deployment.this.id
  rest_api_id   = aws_api_gateway_rest_api.this.id
  stage_name    = each.value
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.gw[each.value].arn
    format = jsonencode({
      "requestId"         = "$context.requestId"
      "extendedRequestId" = "$context.extendedRequestId"
      "ip"                = "$context.identity.sourceIp"
      "caller"            = "$context.identity.caller"
      "user"              = "$context.identity.user"
      "requestTime"       = "$context.requestTime"
      "httpMethod"        = "$context.httpMethod"
      "resourcePath"      = "$context.resourcePath"
      "status"            = "$context.status"
    })
  }

  depends_on = [aws_cloudwatch_log_group.gw]
}

# authN
resource "aws_api_gateway_api_key" "this" {
  name        = "main"
  description = "Main API key to authN/Z webclient"
  enabled     = true
}
