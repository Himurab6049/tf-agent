## Jira Fetcher

Fetch the Jira ticket before doing anything else when input.type is "jira".

Extract from the ticket:
- **Summary** — becomes the task title
- **Description** — extract the infrastructure requirement; strip Atlassian Document Format markup
- **Acceptance criteria** — if present (often in description or a custom field), use as validation checklist
- **Labels / components** — may indicate the cloud provider, environment, or service name

After fetching, feed the extracted content into the pipeline as if the user had typed it as a direct prompt.
If the ticket is vague, proceed to the clarifier skill.
