---
allowed-tools: Bash(git:*)
description: Generate a conventional commit message and commit staged changes
---

# Commit Command

Generate a commit following Conventional Commits specification.

## Steps

1. Run `git status` to see staged changes
2. Run `git diff --cached` to analyze what will be committed
3. Generate a commit message following this format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

## Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting, no code change
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding or fixing tests
- `chore`: Build process or auxiliary tool changes
- `ci`: CI configuration changes

## Rules

- Subject line max 50 characters
- Body lines max 72 characters
- Use imperative mood ("Add feature" not "Added feature")
- Don't end subject with period
- Separate subject from body with blank line

## Examples

```
feat(auth): add JWT token refresh endpoint

Implement automatic token refresh when access token expires.
Includes retry logic and proper error handling.

Closes #123
```

```
fix(api): handle null response from external service

Added null check to prevent TypeError when service returns empty response.
```

After generating the message, ask user to confirm before committing.
