---
allowed-tools: Read, Write, Edit, Glob
description: Generate documentation for code (README, JSDoc, docstrings)
---

# Documentation Command

Generate appropriate documentation for code.

## Types of Documentation

### 1. README.md

For projects or modules:

```markdown
# Project Name

Brief description of what this project does.

## Features

- Feature 1
- Feature 2

## Installation

```bash
npm install project-name
# or
pip install project-name
```

## Quick Start

```code
// Basic usage example
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| option1 | string | "default" | What it does |

## API Reference

### `functionName(param1, param2)`

Description of what it does.

**Parameters:**
- `param1` (string): Description
- `param2` (number): Description

**Returns:** Description of return value

**Example:**
```code
const result = functionName("hello", 42);
```

## Contributing

[How to contribute]

## License

[License type]
```

### 2. Function/Method Documentation

**JavaScript/TypeScript (JSDoc):**
```javascript
/**
 * Brief description of function.
 *
 * @param {string} param1 - Description of param1
 * @param {number} param2 - Description of param2
 * @returns {Promise<Result>} Description of return value
 * @throws {Error} When something goes wrong
 * @example
 * const result = await myFunction("hello", 42);
 */
```

**Python (Docstring):**
```python
def my_function(param1: str, param2: int) -> Result:
    """Brief description of function.

    Longer description if needed.

    Args:
        param1: Description of param1
        param2: Description of param2

    Returns:
        Description of return value

    Raises:
        ValueError: When param1 is empty

    Example:
        >>> result = my_function("hello", 42)
    """
```

**Go:**
```go
// MyFunction does something useful.
//
// It takes param1 and param2 and returns a result.
// Returns an error if something goes wrong.
func MyFunction(param1 string, param2 int) (Result, error) {
```

### 3. Inline Comments

For complex logic only:

```code
// Calculate compound interest using the formula A = P(1 + r/n)^(nt)
// where P = principal, r = rate, n = compounds per year, t = years
```

## Rules

- Don't document obvious code
- Focus on WHY not WHAT
- Keep examples simple and runnable
- Update docs when code changes
- Use consistent format throughout project

## Steps

1. Analyze the code structure
2. Identify what needs documentation
3. Generate appropriate format
4. Ask user where to place it

## Output

Show generated documentation and ask user to confirm before writing.
