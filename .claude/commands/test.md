---
allowed-tools: Read, Write, Edit, Bash
description: Generate unit tests for specified code
---

# Test Generation Command

Generate comprehensive unit tests for specified code.

## Steps

1. Read the target file/function
2. Identify the testing framework (detect from package.json, requirements.txt, etc.)
3. Analyze function signatures, inputs, outputs
4. Generate tests covering:
   - Happy path (normal operation)
   - Edge cases
   - Error cases
   - Boundary conditions

## Framework Detection

| Language | Look For | Framework |
|----------|----------|-----------|
| JavaScript/TS | package.json | jest, vitest, mocha |
| Python | requirements.txt, pyproject.toml | pytest, unittest |
| Go | go.mod | testing (stdlib) |
| Rust | Cargo.toml | cargo test |

## Test Structure

### JavaScript/TypeScript (Jest/Vitest)

```typescript
describe('FunctionName', () => {
  describe('when [scenario]', () => {
    it('should [expected behavior]', () => {
      // Arrange
      const input = ...;

      // Act
      const result = functionName(input);

      // Assert
      expect(result).toBe(expected);
    });
  });
});
```

### Python (pytest)

```python
class TestClassName:
    def test_function_when_condition_should_behavior(self):
        # Arrange
        input_data = ...

        # Act
        result = function_name(input_data)

        # Assert
        assert result == expected

    def test_function_raises_error_when_invalid_input(self):
        with pytest.raises(ValueError):
            function_name(invalid_input)
```

## Test Cases to Generate

1. **Happy Path**: Normal expected usage
2. **Empty Input**: Empty strings, arrays, objects
3. **Null/None**: Null or undefined values
4. **Boundary**: Min/max values, limits
5. **Type Errors**: Wrong types if applicable
6. **Error Conditions**: Expected exceptions

## Rules

- Use descriptive test names
- One assertion per test (when practical)
- Use AAA pattern (Arrange, Act, Assert)
- Mock external dependencies
- Don't test implementation details, test behavior
- Keep tests independent (no shared state)

## Output

Create test file in appropriate location:
- JS/TS: `__tests__/` or `.test.ts` suffix
- Python: `tests/` or `test_` prefix
- Go: `_test.go` suffix

Ask user where to save the tests before writing.
