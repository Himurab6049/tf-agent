## Terraform Validator

Run validation on generated Terraform files. Interpret results and act:

- `terraform validate` fails with syntax/reference errors → fix the HCL and re-run validate before proceeding
- `tflint` errors → fix them; tflint warnings are advisory, note them but continue
- Either tool not installed → log "tool not available" and continue to security scan

Retry limit: attempt to fix and re-validate at most 2 times. If still failing after 2 attempts, report the errors clearly and halt the pipeline — do not create a PR with broken Terraform.

After a passing validation, immediately proceed to SecurityScan without waiting for user input.
