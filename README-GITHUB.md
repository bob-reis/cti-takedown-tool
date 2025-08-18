# 🚀 GitHub Setup & CI/CD Guide

This guide summarizes how to configure the repository with GitHub Actions, SonarCloud and contribution workflows.

## Repository Setup
- `.gitignore` – ignore temporary files
- `.golangci.yml` – Go linter configuration
- `sonar-project.properties` – SonarCloud settings
- GitHub Actions workflows for CI, PR validation and code quality

## Required Secrets
Set the following secrets in **Settings > Secrets and variables > Actions**:
- `SONAR_TOKEN`
- `DOCKER_USERNAME` and `DOCKER_PASSWORD` (optional)
- `SLACK_WEBHOOK` (optional)

## Branch Protection
Require pull requests, passing status checks and code reviews before merging to `main`.

## Development Flow
1. Create a feature branch
2. Run tests and lint locally
3. Commit and push
4. Open a pull request and wait for automated checks

For detailed YAML workflow examples and troubleshooting, see the [Portuguese version](README-GITHUB.pt-BR.md).
