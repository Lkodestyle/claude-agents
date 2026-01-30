---
allowed-tools: Read, Glob, Grep, Bash(git:*)
description: Perform a code review on staged changes or specified files
---

# Code Review Command

Perform a thorough code review on changes.

## What to Review

If no files specified, review staged changes:
```bash
git diff --cached --name-only
```

## Review Checklist

### 1. Code Quality
- [ ] Code is readable and self-documenting
- [ ] Functions/methods are small and focused
- [ ] No code duplication (DRY)
- [ ] Proper error handling
- [ ] No hardcoded values (use constants/config)

### 2. Logic & Correctness
- [ ] Logic is correct and handles edge cases
- [ ] No off-by-one errors
- [ ] Null/undefined checks where needed
- [ ] Proper async/await usage
- [ ] No race conditions

### 3. Security
- [ ] No sensitive data exposed (passwords, keys, tokens)
- [ ] Input validation present
- [ ] No SQL injection vulnerabilities
- [ ] No XSS vulnerabilities
- [ ] Proper authentication/authorization checks

### 4. Performance
- [ ] No unnecessary loops or iterations
- [ ] Efficient data structures used
- [ ] No N+1 query problems
- [ ] Proper caching where applicable
- [ ] No memory leaks

### 5. Testing
- [ ] Tests cover new functionality
- [ ] Edge cases tested
- [ ] Tests are readable and maintainable

### 6. Style & Conventions
- [ ] Follows project coding standards
- [ ] Consistent naming conventions
- [ ] Proper indentation and formatting
- [ ] No commented-out code
- [ ] Meaningful commit messages

## Output Format

```markdown
## Code Review Summary

### Overview
[Brief summary of changes]

### Findings

#### Critical (Must Fix)
- [Issue description with file:line reference]

#### Warnings (Should Fix)
- [Issue description with file:line reference]

#### Suggestions (Nice to Have)
- [Improvement suggestion]

#### Positive Notes
- [What was done well]

### Verdict
[ ] Approved
[ ] Approved with suggestions
[ ] Changes requested
```

Be constructive and specific. Reference exact lines when possible.
