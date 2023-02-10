module github.com/kislerdm/diagramastext/lambda

go 1.19

require (
	github.com/aws/aws-lambda-go v1.37.0
	github.com/kislerdm/diagramastext/core v0.0.1
)

replace github.com/kislerdm/diagramastext/core v0.0.1 => ../../
