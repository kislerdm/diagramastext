module github.com/kislerdm/diagramastext/lambda

go 1.19

require (
	github.com/aws/aws-lambda-go v1.37.0
	github.com/kislerdm/diagramastext/core v0.0.2
	github.com/kislerdm/diagramastext/core/storage v0.0.1
)

require github.com/lib/pq v1.10.7 // indirect

replace (
	github.com/kislerdm/diagramastext/core v0.0.2 => ../../
	github.com/kislerdm/diagramastext/core/storage v0.0.1 => ../../storage
)
