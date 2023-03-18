.DEFAULT_GOAL := help

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

LS_CORE := `go list ./... | grep -v compression`

.test-core:
	@ echo "test core"
	@ go mod tidy
	@ go test $(LS_CORE) -coverprofile=.coverage.out -timeout=2m
	@ go tool cover -func .coverage.out | tail -1 && rm .coverage.out

LS_PKGS := $(dir $(wildcard $(PWD)/pkg/*/go.mod))

.test-packages:
	@ echo "test packages"
	@ for pkg in $(LS_PKGS); do \
		  cd $$pkg ;\
		  go mod tidy ;\
		  go test . -coverprofile=.coverage.out -timeout=2m ;\
		  rm .coverage.out ;\
	  done

tests: .test-core .test-packages ## Run tests of the core module.

ARCH := `uname -m`
OS := `uname | tr '[:upper:]' '[:lower:]'`

compile: ## Compiles httpserver.
	@ test -d bin || mkdir -p bin && \
		cd cmd/httpserver && \
		go mod tidy && \
			CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o ../../bin/httpserver -ldflags="-s -w" .

