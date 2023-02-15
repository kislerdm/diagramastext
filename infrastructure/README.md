# Infrastructure provisioning

## Prerequisites

- An account with the IAM Policy `arn:aws:iam::aws:policy/AdministratorAccess` attached
- Programmatic access, i.e. key and secret pair
- Terraform ~> 1.3
- S3 bucket to use as tf backend

## GitHub Actions

Follow
the [instructions](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services#adding-the-identity-provider-to-aws)
to setup OIDC authentication.

See the official AWS github action's [documentation](https://github.com/aws-actions/configure-aws-credentials) as for
reference.
