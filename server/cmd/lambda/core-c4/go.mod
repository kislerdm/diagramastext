module lambda/core-c4

go 1.19

require (
	github.com/aws/aws-lambda-go v1.37.0
	github.com/aws/aws-sdk-go-v2/config v1.18.14
	github.com/kislerdm/diagramastext/server v0.0.2
	github.com/kislerdm/diagramastext/server/pkg/core v0.0.1
	github.com/kislerdm/diagramastext/server/pkg/rendering/plantuml v0.0.1
	github.com/kislerdm/diagramastext/server/pkg/secretsmanager v0.0.1
	github.com/kislerdm/diagramastext/server/pkg/storage v0.0.1
)

require (
	github.com/aws/aws-sdk-go-v2 v1.17.5 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.14 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.30 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.18.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.4 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/lib/pq v1.10.7 // indirect
)

replace (
	github.com/kislerdm/diagramastext/server v0.0.2 => ./../../../
	github.com/kislerdm/diagramastext/server/pkg/core v0.0.1 => ./../../../pkg/core
	github.com/kislerdm/diagramastext/server/pkg/rendering/plantuml v0.0.1 => ./../../../pkg/rendering/plantuml
	github.com/kislerdm/diagramastext/server/pkg/secretsmanager v0.0.1 => ./../../../pkg/secretsmanager
	github.com/kislerdm/diagramastext/server/pkg/storage v0.0.1 => ./../../../pkg/storage
)
