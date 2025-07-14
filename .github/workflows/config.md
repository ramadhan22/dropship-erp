# GitHub Actions Configuration

# Branch protection rules (to be set in repository settings)
PROTECTED_BRANCHES:
  - main
  - master

# Required status checks (to be configured in branch protection)
REQUIRED_CHECKS:
  - "Backend Review (Go)"
  - "Frontend Review (React/TypeScript)" 
  - "All CI Checks"

# Environment variables used across workflows
ENVIRONMENT_VARS:
  GO_VERSION: "1.23"
  NODE_VERSION: "18"
  
# Cache keys for better performance
CACHE_KEYS:
  go_modules: "${{ runner.os }}-go-${{ hashFiles('backend/go.sum') }}"
  npm_deps: "node-modules-${{ hashFiles('frontend/dropship-erp-ui/package-lock.json') }}"

# Workflow permissions (minimal required)
PERMISSIONS:
  contents: read
  pull-requests: write  # For automated comments
  security-events: write  # For security scanning uploads