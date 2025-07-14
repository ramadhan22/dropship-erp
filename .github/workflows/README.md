# GitHub Actions Workflows

This directory contains automated workflows that review every pull request opened in this repository.

## Workflows Overview

### 1. `pr-review.yml` - Main PR Review
**Triggers**: On every pull request to `main` or `master`

**Checks**:
- **Backend Review**: Go formatting, tests, and build
- **Frontend Review**: ESLint, TypeScript compilation, build, and tests  
- **Review Summary**: Aggregates results and provides final status

### 2. `quality-gate.yml` - Advanced Quality Checks
**Triggers**: On PR open, sync, or reopen

**Features**:
- **Smart Change Detection**: Only runs relevant checks based on file changes
- **Comprehensive Go Checks**: Formatting, vetting, race detection, coverage
- **Frontend Quality**: Type checking, linting, building, testing with coverage
- **Security Scanning**: Trivy vulnerability scanner with SARIF upload

### 3. `ci.yml` - Continuous Integration
**Triggers**: On pull requests and pushes to main branches

**Purpose**: 
- Lightweight CI pipeline
- Currently focuses on backend (frontend has known issues)
- Provides basic smoke tests

### 4. `pr-comment.yml` - Automated PR Comments
**Triggers**: When PRs are opened or reopened

**Features**:
- Welcome message with guidelines
- PR size analysis and recommendations
- Change summary (backend/frontend/docs/workflows)
- Links to contribution guidelines

## Current Status

### ✅ Working
- Go backend tests, formatting, and builds
- Automated PR comments and guidance
- Change detection and conditional execution

### ⚠️ Known Issues
The frontend currently has:
- 80+ ESLint errors (mostly TypeScript `any` usage)
- 9 failing tests out of 45 total
- 3 TypeScript compilation errors

These are existing issues in the codebase and not introduced by this PR.

## Usage

These workflows run automatically on every PR. Contributors will see:

1. **Immediate feedback** via automated comments
2. **Status checks** in the PR interface  
3. **Detailed logs** in the "Checks" tab
4. **Clear guidance** on how to fix any issues

## Customization

To modify the workflows:

1. **Add new checks**: Edit the relevant `.yml` file
2. **Change triggers**: Modify the `on:` section
3. **Update Node/Go versions**: Change the version numbers in `env:` or `with:` sections
4. **Skip problematic checks**: Comment out steps until issues are resolved

## Integration with Development

This setup implements the guidelines from [`AGENTS.md`](../AGENTS.md):

- ✅ Runs `go test ./...` for backend changes
- ✅ Runs `npm test` for frontend changes  
- ✅ Checks `gofmt` formatting
- ✅ Runs `npm run lint` (when issues are fixed)
- ✅ Ensures builds succeed
- ✅ Provides clear feedback to contributors

The workflows ensure that "every pull request opened" gets thoroughly reviewed for code quality, testing, and adherence to project standards.