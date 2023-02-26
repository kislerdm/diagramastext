.DEFAULT_GOAL := help

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

CODE_PATH :=
ZIPNAME := lambda.zip
BIN_PATH := $(PWD)/bin

compile:
	@ test -d $(BIN_PATH) || mkdir -p $(BIN_PATH) && \
		test $(CODE_PATH) || echo "CODE_PATH not set" && exit 137  &&\
		cd $(CODE_PATH) && \
 		go mod tidy && \
  		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN_PATH)/lambda -ldflags="-s -w" .

build: compile ## Builds the lambda binary and archives it.
	@ cd $(BIN_PATH) && zip -9 $(ZIPNAME) lambda && rm lambda
