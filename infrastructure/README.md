# Provision app deployment environment

## Structure

The core project `root` owns the environment projects to deploy the core logic to:
```commandline
root
|- staging
`- production
```

## Prerequisites

1. Service Account with the JSON key
2. Export path to the key:
```commandline
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/json/key
```

## Commands

1. Navigate to the directory with the terraform instructions

```commandline
cd gcp
```

2. Switch environments: `staging`/`production`:

```commandline
terraform workspace select staging
```

```commandline
terraform workspace select production
```

3. Execute the tf commands:

```commandline
terraform init
terraform plan
terraform apply -auto-approve
```
