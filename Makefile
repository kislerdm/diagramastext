.DEFAULT_GOAL := help

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

PORT_SERVER := 9000
PORT_CLIENT := 9001
GIT_SHA := `git log --pretty=format:"%H" -1`

localenv: ## Provisions local development environment.
	@ if [ "${OPENAI_API_KEY}" == "" ]; then echo "set OPENAI_API_KEY environment variable"; exit 137; fi
	@ echo "access webclient on http://localhost:${PORT_CLIENT}"
	@ VERSION=${GIT_SHA} PORT_SERVER=${PORT_SERVER} PORT_CLIENT=${PORT_CLIENT} docker compose up --force-recreate
