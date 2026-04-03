## Security Scanner

Run checkov on generated Terraform. Classify findings and act:

### Blockers — fix before creating PR:
- Hardcoded credentials, passwords, or API keys in any resource argument
- S3 bucket with `acl = "public-read"` or `"public-read-write"` without explicit justification
- Security group ingress rules open to `0.0.0.0/0` or `::/0` on port 22, 3389, or any database port
- RDS, EBS volumes, S3 buckets, or EFS without server-side encryption enabled
- IAM policy with `actions = ["*"]` and `resources = ["*"]`
- IMDSv1 enabled on EC2 (missing `metadata_options { http_tokens = "required" }`)

### Warnings — include in PR body, do not block:
- Missing recommended tags
- CloudTrail, VPC flow logs, or S3 access logging not enabled
- Backup retention period below 7 days
- Multi-AZ not enabled on RDS in prod environments

If checkov is not installed: note "checkov not available — manual security review recommended" in the PR body and proceed.

After security scan (pass or warn-only), proceed to CreatePR.
