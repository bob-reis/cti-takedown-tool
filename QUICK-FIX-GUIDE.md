# 🚨 Quick Fix Guide - GitHub Actions & SonarCloud

Summary of common CI/CD issues and their solutions.

## Problems
1. GitHub Actions failing due to wrong Go action or missing coverage file
2. SonarCloud project/organization mismatch

## Solutions
- Replace workflow with a tested Go build and SonarCloud scan
- Update `sonar-project.properties` with correct keys and coverage path
- Configure `SONAR_TOKEN` and optional Docker credentials in repository secrets

## After Fixing
Commit changes and push to trigger the workflows. Verify results in the Actions tab and on SonarCloud.

For detailed YAML and troubleshooting examples, see the [Portuguese version](QUICK-FIX-GUIDE.pt-BR.md).
