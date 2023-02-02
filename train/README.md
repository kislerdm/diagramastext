# OpenAPI Model Fine-tune

The codebase to fine-tune the model following the [instructions](https://platform.openai.com/docs/guides/fine-tuning).

## How to run

Run to list all commands:

```commandline
make help
```

Run to build the CLI:

```commandline
make init
```

Run to generate training data:

```commandline
make data
```

Run to create a fine-tuned model (
see [details](https://platform.openai.com/docs/guides/fine-tuning/create-a-fine-tuned-model)):

```commandline
make train
```

The [`curie` model](https://platform.openai.com/docs/models/curie) is used as the basis.

Run to specify the base model:

```commandline
make train MODEL=##model name##
```

**Note** that the OpenAPI access key must be set as the environment variable `OPENAI_API_KEY`.  
