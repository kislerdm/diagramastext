resource "aws_api_gateway_request_validator" "this" {
  name                        = "main-validator${local.suffix}"
  rest_api_id                 = aws_api_gateway_rest_api.this.id
  validate_request_body       = true
  validate_request_parameters = true
}

resource "aws_api_gateway_rest_api" "this" {
  name           = "main${local.suffix}"
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

locals {
  allowed_headers_response = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
  }

  cors_headers_gw = { for k, v in local.cors_headers : "method.response.header.${k}" => "'${v}'" }

  endpoints = {
    "c4" : ["POST"],
  }

  request_parameters = merge({
    # "method.request.header.UserID" = true
    # "method.request.header.Authorization" = true
    }
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

  deployment_trigger_obj = merge(
    local.endpoints,
    local.allowed_headers_response,
    local.cors_headers_gw,
    local.request_parameters,
    local.lambda_trigger,
    {
      "${module.core_rendering_c4.function_name}" = module.core_rendering_c4.codebase_md5
    },
  )
  deployment_trigger = jsonencode(local.deployment_trigger_obj)
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
  function_name = module.core_rendering_c4.function_name
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
  response_parameters = local.cors_headers_gw
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
  uri                     = module.core_rendering_c4.invoke_arn

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
  response_parameters = local.cors_headers_gw
  response_templates = {
    "application/json" = aws_api_gateway_model.schema_response.name
  }
}

# stage and deployment

resource "aws_cloudwatch_log_group" "gw" {
  name              = "API-Gateway-Execution-Logs_${aws_api_gateway_rest_api.this.id}"
  retention_in_days = 7
}

resource "aws_api_gateway_deployment" "this" {
  rest_api_id = aws_api_gateway_rest_api.this.id

  triggers = {
    redeployment = sha1(jsonencode(
      concat([
        local.deployment_trigger,
        aws_api_gateway_request_validator.this.id,
        aws_api_gateway_rest_api.this.id,
        aws_api_gateway_model.schema_request.schema,
        aws_api_gateway_model.schema_response.schema,
        ],
        [for i in aws_api_gateway_resource.route_top : i.id],
      )
    ))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "this" {
  cache_cluster_size    = "0.5"
  cache_cluster_enabled = false
  deployment_id         = aws_api_gateway_deployment.this.id
  rest_api_id           = aws_api_gateway_rest_api.this.id
  stage_name            = "base"
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.gw.arn
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

# plan

resource "aws_api_gateway_usage_plan" "test" {
  name        = "test${local.suffix}"
  description = "Test usage plan"

  api_stages {
    api_id = aws_api_gateway_rest_api.this.id
    stage  = aws_api_gateway_stage.this.stage_name
    throttle {
      path        = "/c4/POST"
      burst_limit = 10
      rate_limit  = 2
    }
  }

  throttle_settings {
    burst_limit = 100
    rate_limit  = 10
  }
}
# authN
resource "aws_api_gateway_api_key" "main" {
  name        = "main${local.suffix}"
  description = "Main API key to authN/Z webclient"
  enabled     = true
}

resource "aws_api_gateway_usage_plan_key" "main" {
  key_id        = aws_api_gateway_api_key.main.id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.test.id
}

# custom diagram
resource "aws_api_gateway_domain_name" "this" {
  domain_name     = "api.${local.subdomain_prefix}diagramastext.dev"
  certificate_arn = "arn:aws:acm:us-east-1:027889758114:certificate/74feb1e2-797b-4ebb-8399-e1eee4ace87d"
}

resource "aws_api_gateway_base_path_mapping" "this" {
  api_id      = aws_api_gateway_rest_api.this.id
  stage_name  = aws_api_gateway_stage.this.stage_name
  domain_name = aws_api_gateway_domain_name.this.domain_name
}

output "gw_domain_name" {
  value       = aws_api_gateway_domain_name.this.cloudfront_domain_name
  description = "API GW domain name required to configure custom DNS, e.g. Cloudflaire"
}
