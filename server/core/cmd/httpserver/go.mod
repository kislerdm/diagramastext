module httpserver

go 1.19

require (
	github.com/kislerdm/diagramastext/server/core v0.0.2
	github.com/kislerdm/diagramastext/server/core/pkg/gcpsecretsmanager v0.0.1
	github.com/kislerdm/diagramastext/server/core/pkg/httpclient v0.0.1
	github.com/kislerdm/diagramastext/server/core/pkg/openai v0.0.1
	github.com/kislerdm/diagramastext/server/core/pkg/postgres v0.0.1
)

require (
	cloud.google.com/go/compute v1.12.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.1 // indirect
	cloud.google.com/go/iam v0.8.0 // indirect
	cloud.google.com/go/secretmanager v1.10.0 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.0 // indirect
	github.com/googleapis/gax-go/v2 v2.7.0 // indirect
	github.com/lib/pq v1.10.7 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/net v0.0.0-20221014081412-f15817d10f9b // indirect
	golang.org/x/oauth2 v0.0.0-20221014153046-6fdb5e3db783 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/api v0.103.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20221201164419-0e50fba7f41c // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace (
	github.com/kislerdm/diagramastext/server/core v0.0.2 => ../../
	github.com/kislerdm/diagramastext/server/core/pkg/gcpsecretsmanager v0.0.1 => ../../pkg/gcpsecretsmanager
	github.com/kislerdm/diagramastext/server/core/pkg/httpclient v0.0.1 => ../../pkg/httpclient
	github.com/kislerdm/diagramastext/server/core/pkg/openai v0.0.1 => ../../pkg/openai
	github.com/kislerdm/diagramastext/server/core/pkg/postgres v0.0.1 => ../../pkg/postgres
)
