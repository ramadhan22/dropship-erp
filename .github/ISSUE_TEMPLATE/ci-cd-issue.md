---
name: CI/CD Issue
about: Report problems with automated PR reviews or CI/CD workflows
title: '[CI/CD] '
labels: ['ci/cd', 'automation']
assignees: []
---

## Issue Description
Describe the problem with the automated PR review system.

## Workflow Affected
- [ ] PR Review (`pr-review.yml`)
- [ ] Quality Gate (`quality-gate.yml`) 
- [ ] CI Pipeline (`ci.yml`)
- [ ] PR Comments (`pr-comment.yml`)

## Expected Behavior
What should have happened?

## Actual Behavior
What actually happened?

## Steps to Reproduce
1. 
2. 
3. 

## Pull Request Link
If applicable, link to the PR where the issue occurred.

## Error Messages
```
Paste any error messages from the workflow logs here
```

## Additional Context
Add any other context about the problem here.

---

### For Contributors
If you're having trouble with failing checks:

1. **Check the workflow logs** in the "Checks" tab of your PR
2. **Review our guidelines** in [AGENTS.md](../AGENTS.md)
3. **Common fixes**:
   - Run `gofmt -w .` in the `backend` directory
   - Run `go test ./...` to check for test failures
   - Run `npm run lint` and fix any issues
   - Ensure `npm run build` succeeds