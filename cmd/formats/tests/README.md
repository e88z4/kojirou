# Test Organization

This directory is for high-level, integration, and business logic tests that span multiple formats or require the main application context.

## Guidelines
- Place business logic and end-to-end tests here (e.g., business_format_test.go).
- Place format-specific unit tests in their respective format subfolders (e.g., kindle/, epub/, progress/).
- Place test helpers in `testhelpers/`.

## Structure
- `business_format_test.go`: End-to-end and business logic tests for multi-format generation.
- `../epub/`: EPUB and KEPUB format-specific tests.
- `../kindle/`: Kindle format-specific tests.
- `../progress/`: Progress bar and reporting tests.
- `../testhelpers/`: Shared test helpers for all formats.

This structure keeps tests organized, maintainable, and easy to run with `go test ./...`.
