---
allowed-tools: Read, Write, Edit, Glob, Grep
description: Suggest and apply refactoring improvements to code
---

# Refactor Command

Analyze code and suggest/apply refactoring improvements.

## Steps

1. Read the target code
2. Identify code smells and improvement opportunities
3. Suggest refactoring with explanation
4. Apply changes if user approves

## Code Smells to Detect

### Complexity
- Long methods (>20 lines)
- Deep nesting (>3 levels)
- Complex conditionals
- God classes/functions

### Duplication
- Repeated code blocks
- Similar logic in multiple places
- Copy-paste patterns

### Naming
- Unclear variable/function names
- Inconsistent naming conventions
- Abbreviations without context

### Structure
- Large files (>300 lines)
- Mixed responsibilities
- Tight coupling
- Missing abstractions

## Refactoring Techniques

| Smell | Technique |
|-------|-----------|
| Long method | Extract Method |
| Duplicate code | Extract Function/Class |
| Complex conditional | Replace with Polymorphism, Guard Clauses |
| Deep nesting | Early Returns, Extract Method |
| Large class | Extract Class, Single Responsibility |
| Long parameter list | Introduce Parameter Object |
| Feature envy | Move Method |
| Primitive obsession | Replace with Value Object |

## Output Format

```markdown
## Refactoring Analysis

### Current Issues

1. **[Issue Type]** in `file:line`
   - Problem: [Description]
   - Impact: [Why it matters]

### Suggested Refactorings

#### 1. [Refactoring Name]

**Before:**
```code
[current code]
```

**After:**
```code
[refactored code]
```

**Why:** [Explanation of improvement]

### Summary

| Metric | Before | After |
|--------|--------|-------|
| Lines of code | X | Y |
| Cyclomatic complexity | X | Y |
| Functions | X | Y |
```

## Rules

- Don't change behavior, only structure
- One refactoring at a time
- Explain the "why" not just the "what"
- Consider test coverage before refactoring
- Preserve existing tests
- Keep changes minimal and focused

Ask user to confirm before applying changes.
