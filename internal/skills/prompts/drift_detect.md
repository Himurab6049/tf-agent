## Drift Detector

Run `terraform plan` against live infrastructure state to detect drift between actual state and IaC.

- Exit code 0 from terraform plan → no drift, report clean
- Exit code 2 from terraform plan → drift detected, parse and summarise which resources changed
- terraform or state not initialised → report "Unable to detect drift: run terraform init first"

When drift is found:
1. List each drifted resource with its type, name, and what changed (add/change/destroy)
2. Suggest the most likely cause (manual console change, out-of-band script, etc.)
3. Propose remediation: either update IaC to match reality, or re-apply to restore desired state

Keep the report concise — one line per resource, followed by a short recommendation block.
