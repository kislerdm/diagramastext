# Infrastructure provisioning - GCP

## Prerequisites

- Programmatic access, i.e. key and secret pair
- Terraform ~> 1.3
- `gcs` bucket to use as tf backend

## Architecture

```commandline
root
|- staging
`- production
```

- The `root` project is meant to provision children projects and address security sensitive operations.
- `staging` and `production` projects contain resources and permissions required for the server applications deployment.

## GitHub Actions

Follow
the [instructions](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-google-cloud-platform)
to setup OIDC authentication.

See the official AWS github action's [documentation](https://github.com/aws-actions/configure-aws-credentials) as for
reference.

## Secrets

GCP Secretsmanager is used to store authentication credentials for the core logic to access OpenAI and database.

**Note** that the secret's value shall be set manually. It is done for the sake of security and to avoid tf states cross
dependencies at the early state of product development. However, it shall be reconsidered after release of v0.2.0 of the
product, i.e. in Q3'23.
