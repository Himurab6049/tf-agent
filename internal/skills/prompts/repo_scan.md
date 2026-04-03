## Repo Scanner

Scan the repository before generating any Terraform. Extract:

1. Existing `.tf` files — provider versions, naming patterns, tag structure, variable naming style, module structure
2. `*.tfvars` files — environment names, region defaults, existing variable values
3. `README.md` or `docs/` — architecture context, team conventions
4. Directory layout — determine if new code should be a root config or a reusable module

Return a structured summary covering:
- Cloud provider and version constraints in use
- Naming convention (e.g. `<env>-<service>-<resource>`)
- Tag schema (e.g. `env`, `team`, `managed-by`)
- Regions and environments already defined
- Any modules already in use

If no Terraform files exist, state "greenfield project" and apply AWS best practices as defaults.
