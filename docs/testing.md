# Kojirou Test Documentation

This document provides information about the testing strategy, test coverage, and methodology used for Kojirou.

## Test Strategy

Kojirou employs a multi-layered testing approach:

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test interactions between components
3. **End-to-End Tests**: Test complete workflows from input to output
4. **Manual Tests**: Verify output on actual devices

### Unit Testing

Unit tests focus on testing individual functions and components in isolation, mocking dependencies when necessary. Key areas covered by unit tests include:

- Format parsing and validation
- Image processing functions
- Metadata extraction
- Template rendering
- Error handling
- Configuration processing

### Integration Testing

Integration tests verify that components work correctly together. These tests focus on:

- Complete workflow from manga data to output files
- Simultaneous generation of multiple formats
- Processing of various manga types (standard pages, wide pages, etc.)
- Error handling across component boundaries

### End-to-End Testing

End-to-end tests verify the complete workflow using realistic inputs:

- Processing manga data into different output formats
- Verifying correct file structure in output files
- Checking content and metadata in generated files
- Testing with large volumes and diverse content

## Test Coverage

Current test coverage for the codebase:

| Package | Coverage |
|---------|----------|
| formats/epub | ~85% |
| formats/kindle | ~90% |
| mangadex | ~80% |
| cmd | ~75% |
| Overall | ~82% |

### Coverage Details

- **High coverage areas**: Core functionality including manga data processing, EPUB generation, and KEPUB conversion
- **Medium coverage areas**: Command-line interface, progress reporting
- **Lower coverage areas**: Some error handling paths, extreme edge cases

## Testing Methodologies

### Test Helpers

The project uses a shared `testhelpers` package to provide common test data and utilities:

- `CreateTestManga()`: Creates standard test manga data
- `CreateTestImage()`: Generates test images with different dimensions
- `CreateWidePageTestManga()`: Creates test manga with wide pages
- Various other specialized test manga creators

### Test Data Management

- Test data is generated programmatically to avoid large test files
- Test files are cleaned up automatically after tests
- Temporary directories are used for output files

### Continuous Integration

All tests are run automatically on:
- Pull request creation
- Merges to main branch
- Release preparation

## Running Tests

### Running All Tests

```bash
go test ./...
```

### Running Tests with Coverage

```bash
go test -cover ./...
```

### Generating Coverage Report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Environment Setup

For complete testing:

1. Install Go 1.17 or later
2. Clone the repository
3. Run `go mod download` to fetch dependencies
4. Run tests with `go test ./...`