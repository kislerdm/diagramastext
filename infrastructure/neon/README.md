# Neon database provisioning

Required:
- GCP SA with the `role/owner` permissions for the project `diagramastext-root`
- JSON authN key
- `GOOGLE_APPLICATION_CREDENTIALS` with the path to the JSON key
- Neon API key
- tf workspace `neon`:
```commandline
terraform workspace new neon
terraform workspace select neon
```
