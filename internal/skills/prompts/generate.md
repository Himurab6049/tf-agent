## Terraform Generator

Generate production-quality Terraform HCL. Always split into these files:

- `providers.tf` — terraform block with required_version and required_providers with pinned versions
- `variables.tf` — all input variables with type, description, and default where appropriate
- `main.tf` — resource and data source definitions
- `outputs.tf` — output values with descriptions

Rules:
- Pin provider versions: `version = "~> X.Y"` — never use `>= X.Y` without upper bound
- Minimum Terraform version: `required_version = ">= 1.5.0"`
- Use `locals {}` for repeated values, computed strings, and tag maps
- Every variable must have `type` and `description`; add `default` for optional vars
- No hardcoded secrets, account IDs, or regions — use variables or data sources
- Apply tags on every resource that supports them using a `local.tags` map
- Follow naming conventions extracted from repo scan; default to `<env>-<service>-<resource>` in kebab-case
- Use `lifecycle { prevent_destroy = true }` on stateful resources (RDS, DynamoDB, S3, EFS)
- Prefer data sources over hardcoded IDs for VPCs, subnets, and AMIs
- Group related resources in `main.tf` with comment headers (e.g. `# --- Networking ---`)
