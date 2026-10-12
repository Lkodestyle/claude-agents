---
allowed-tools: Read, Glob, Grep, Bash
description: Help debug an error or unexpected behavior
---

# Debug Command

Systematic approach to debugging issues.

## Input

User should provide:
- Error message or unexpected behavior
- File/function where it occurs (if known)
- Steps to reproduce (if known)

## Debug Process

### 1. Gather Information

```bash
# Check error logs
# Check recent changes
git log --oneline -10
git diff HEAD~1
```

### 2. Understand the Error

Parse the error message:
- Error type (TypeError, ValueError, etc.)
- Stack trace
- Line numbers
- Variable values mentioned

### 3. Locate the Problem

- Read the file(s) mentioned in stack trace
- Check related imports/dependencies
- Look for recent changes in that area

### 4. Analyze Root Cause

Common causes:
| Error Type | Common Causes |
|------------|---------------|
| TypeError | Wrong type, null/undefined access |
| ReferenceError | Undefined variable, typo |
| SyntaxError | Missing bracket, quote, semicolon |
| ImportError | Wrong path, missing dependency |
| ConnectionError | Network, credentials, timeout |
| PermissionError | File/resource access rights |

### 5. Propose Solution

## Output Format

```markdown
## Debug Analysis

### Error Summary
[What the error is in plain language]

### Root Cause
[Why this error is happening]

### Location
`file:line` - [description of problematic code]

### Solution

**Option 1 (Recommended):**
```code
[fix code]
```

**Why this works:** [explanation]

**Option 2 (Alternative):**
[alternative approach if applicable]

### Prevention
[How to prevent this in the future]

### Related Files to Check
- [file1] - [why]
- [file2] - [why]
```

## Debug Techniques

### Add Logging
```javascript
console.log('Variable X:', x);
console.log('Type:', typeof x);
```

```python
print(f"Variable X: {x}")
import pdb; pdb.set_trace()  # breakpoint
```

### Check Assumptions
- Is the variable the expected type?
- Is it null/undefined/None?
- Is the array/object empty?
- Are async operations completing?

### Binary Search
- Comment out half the code
- If error persists, problem is in remaining half
- Repeat until isolated

## Rules

- Don't guess - verify with evidence
- Check the obvious first (typos, imports)
- Read error messages carefully
- One change at a time when testing fixes
- Explain why the fix works
