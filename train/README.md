# OpenAPI Model Fine-tune

The codebase to fine-tune the model following the [instructions](https://platform.openai.com/docs/guides/fine-tuning).

## How to run

Build the CLI:

```commandline
docker build -t openapi .
```

Run to generate training data:
```commandline
docker run -v ${PWD}/data:/data --entrypoint python3 -t openapi:latest \
/datagen.py -o /data/sample.jsonl
```

Run to create a fine-tuned model (
see [details](https://platform.openai.com/docs/guides/fine-tuning/create-a-fine-tuned-model)):

```commandline
docker run -v ${PWD}/data:/data -e OPENAI_API_KEY=${OPENAI_API_KEY} -t openapi:latest \
api fine_tunes.create -t /data/sample.jsonl -m curie
```

**Note** that the OpenAPI access key must be set as the environment variable `OPENAI_API_KEY`.  
