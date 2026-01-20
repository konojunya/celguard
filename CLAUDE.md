# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

celguard is a GitHub Action that validates Pull Requests using Common Expression Language (CEL). It reads validation rules from `.github/celguard.yaml` and ensures PR metadata (title, body, branch, labels, etc.) follows team conventions.

## Development Commands

### Testing
```bash
go test ./...
```

### Building
```bash
go build -o celguard .
```

### Running locally
```bash
# Requires GITHUB_EVENT_PATH and GITHUB_TOKEN environment variables
./celguard [path/to/config.yaml]
# Default config path: .github/celguard.yaml
```

## Architecture

### Core Components

The application is structured as a single-package Go application with the following key files:

- **main.go**: Entry point and core CEL evaluation logic
  - `run()`: Main validation engine that compiles and evaluates CEL expressions against PR data
  - Provides two variables to CEL expressions: `value` (current field being validated) and `pr` (entire PR context map)
  - Returns formatted error messages with `[field] error message` format

- **config.go**: Configuration file parsing
  - `ReadConfig()`: Reads YAML config from `.github/celguard.yaml` (or custom path)
  - Uses `GITHUB_WORKSPACE` env var for path resolution in GitHub Actions context
  - Accepts config path as first command-line argument

- **github.go**: GitHub API integration
  - `upsertFailedComment()`: Creates or updates PR comment with validation errors (uses special marker for idempotency)
  - `deleteFailedComment()`: Removes validation error comments when PR passes
  - Comments are marked with `<!-- celguard:konojunya/celguard -->` for identification

- **model.go**: Type definitions
  - `Config`: Map of field names to `Rule` structs
  - `Rule`: Contains CEL expression and error message
  - `Event`: GitHub webhook event structure (mirrors GitHub's pull_request event)

### Validation Flow

1. Load config from `.github/celguard.yaml`
2. Parse GitHub event from `GITHUB_EVENT_PATH` environment variable
3. For each configured rule:
   - Create CEL environment with string extensions
   - Compile and evaluate CEL expression with `value` and `pr` variables
   - Collect validation failures
4. If running in GitHub Actions (`GITHUB_EVENT_NAME=pull_request`):
   - On failure: upsert comment with errors
   - On success: delete any existing error comments
5. Exit with status 1 on validation failure, 0 on success

### Supported PR Fields

The following fields can be validated (accessible via `value` in CEL):
- `title`: PR title (string)
- `body`: PR body/description (string)
- `author`: PR author's GitHub username (string)
- `base_ref`: Target branch name (string)
- `head_ref`: Source branch name (string)
- `labels`: Array of label names ([]string)

All fields are also available in the `pr` map for cross-field validation.

### CEL Expression Context

CEL expressions have access to:
- `value`: The current field being validated (type depends on field)
- `pr`: Map containing all PR fields (useful for cross-field validation)
- CEL string extensions (via `ext.Strings()`) providing `matches()`, `size()`, etc.

## Configuration Examples

The `.github/celguard.yaml` file in this repo demonstrates conventional commit validation:

```yaml
title:
  cel: "value.matches('^(feat|fix|docs|style|refactor|test|chore): .+')"
  error: PR title must follow conventional commits format
```

See README.md for more configuration examples and supported validation patterns.

## GitHub Actions Integration

This repository is both:
1. A Go application that validates PRs
2. A composite GitHub Action defined in `action.yaml`

The action downloads the pre-built binary from GitHub releases and executes it with the provided config path.

## Testing Strategy

Tests are in `main_test.go` using table-driven approach. Each test case:
- Creates an `Event` with specific PR data
- Defines a `Config` with validation rules
- Asserts expected error message or success

When adding new validation features, add test cases covering both pass and fail scenarios for each supported field type.
