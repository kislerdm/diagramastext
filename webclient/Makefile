.DEFAULT_GOAL := help

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

tests: ## Tests the webclient logic.
	@ docker run --rm \
		-v ${PWD}:/app \
		-w /app \
		-t node:alpine3.16 \
		/bin/sh -c "npm install > '/dev/null' && npm test"
