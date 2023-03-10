.DEFAULT_GOAL := help

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

test-postgres:
	@ cd storage && \
 		go mod tidy && \
 		go test -timeout 1m -coverprofile=.coverage.out -coverpkg=. && \
 		go tool cover -func .coverage.out && rm -r .coverage.out

tests: ## Run tests of the core module.
	@ go mod tidy && \
  		go test -timeout 2m --tags=unittest ./... -coverprofile=.coverage.out -coverpkg=./...  && \
		go tool cover -func .coverage.out && rm -r .coverage.out

test-lambda: ## Run tests for the lambda function.
	@ cd cmd/lambda && \
 		go mod tidy && \
 		go test -timeout 3m -coverprofile=.coverage.out . -coverpkg=. && \
		go tool cover -func .coverage.out && rm .coverage.out

compile:
	@ test -d bin || mkdir -p bin && \
 		cd cmd/lambda && \
 		go mod tidy && \
  		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../bin/lambda -ldflags="-s -w" .

ZIPNAME := lambda.zip

build: compile ## Builds the lambda binary and archives it.
	@ cd bin && zip -9 $(ZIPNAME) lambda && rm lambda

deploy-lambda: ## Deploys the core logic as AWS lambda.
	@ cd deployment && terraform plan && terraform apply -auto-approve

compile-httpserver: ## Compiles httpserver
	@ test -d bin || mkdir -p bin && \
		cd cmd/httpserver && \
		go mod tidy && \
			CGO_ENABLED=0 go build -o ../../bin/httpserver -ldflags="-s -w" .

lambda-test:
	@ test -d bin || mkdir -p bin && \
		cd cmd/test && \
		go mod tidy && \
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../bin/test -ldflags="-s -w" .
	@ cd bin && mv test lambda && zip -9 test.zip lambda && rm lambda
