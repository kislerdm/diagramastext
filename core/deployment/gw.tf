resource "aws_api_gateway_rest_api" "this" {
  name = "main"
  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

resource "aws_api_gateway_api_key" "this" {
  name        = "main"
  description = "Main API key to authN/Z webclient"
  enabled     = true
}

resource "aws_api_gateway_request_validator" "this" {
  name                        = "main-validator"
  rest_api_id                 = aws_api_gateway_rest_api.this.id
  validate_request_body       = true
  validate_request_parameters = true
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
