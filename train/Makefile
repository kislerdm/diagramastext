.DEFAULT_GOAL := help

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

init: ## Initialises the tool.
	@ docker build -t openapi .

data: ## Generate data sample to train the model.
	@ docker run -v ${PWD}/data:/data --entrypoint python3 -t openapi:latest /datagen.py -o /data/sample.jsonl

MODEL := curie

runcli: ## Runs CLI commands.
	@ docker run -v ${PWD}/data:/data -e OPENAI_API_KEY=${OPENAI_API_KEY} -t openapi:latest api

train: ## Triggers training job.
	@ docker run -v ${PWD}/data:/data -e OPENAI_API_KEY=${OPENAI_API_KEY} -t openapi:latest \
      api fine_tunes.create -t /data/sample.jsonl -m $(MODEL)
