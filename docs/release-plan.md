# Mockoho Release Plan

This document outlines the tasks needed to prepare Mockoho for release as a standalone application that users can run with a single command.

## 1. Remove Generated Files

The repository contains files that shouldn't be included in the published source:

- Remove compiled binaries (`main` and `mockoho`)
- Remove log files (`debug.log`)

These are already in the `.gitignore` file, but should be manually removed before creating a release.

## 2. Complete Pending Refactoring Tasks

Some refactoring tasks may still be pending:

- **Logger Enhancement**: Ensure the prepend-only logger adds a blank line between sessions for better readability
- **UI Model Optimization**: Simplify list initialization and update functions
- **Server Request Handler**: Break down complex functions into smaller, more focused functions
- **Response Body Processing**: Simplify template processing

Completing these refactorings will make the codebase more maintainable and efficient.

## 3. Add Minimum Tests for Core Functionality

Add essential tests for the core components that currently lack coverage:

- **Config Package**: Test loading, saving, and manipulating configurations
- **Mock Package**: Test endpoint matching and response generation
- **Proxy Package**: Test request forwarding and path rewriting
- **UI Components**: Test basic rendering and state management

Focus on testing critical paths and edge cases rather than aiming for complete coverage.

## 4. Version Management

- Update the version number in `cmd/mockoho/main.go` to follow semantic versioning
- Ensure the version is properly injected during the build process

## 5. Documentation Consolidation

The current documentation still includes 6 separate files in the docs directory. Consolidate these into at most two documents:

### User Guide (README.md or docs/user-guide.md)

Combine these files into a single, comprehensive user manual:

- `getting-started.md`
- `keyboard-shortcuts.md`
- `configuration-reference.md`
- `mock-examples.md`
- `tui-guide.md`

### Developer Guide (docs/developer-guide.md)

Move the technical architecture information here:

- `tui-architecture.md`

This consolidation will make the documentation more maintainable and easier for users to navigate.

## 6. Code Quality Checks

- Run a linter (e.g., golangci-lint) to identify and fix any code quality issues
- Check for any hardcoded values that should be configurable
- Ensure proper error handling throughout the codebase
- Verify that all log messages are appropriate for production use

## 7. Security Review

- Check for any potential security issues (e.g., path traversal vulnerabilities)
- Ensure proper input validation for user-provided data
- Review proxy implementation for security best practices

## 8. Performance Optimization

- Profile the application to identify any performance bottlenecks
- Optimize resource usage, especially for long-running operations

## 9. Cross-Platform Testing

- Test the application on all target platforms (Windows, macOS, Linux)
- Ensure that file paths work correctly on all platforms
- Verify that editor integration works across platforms

## 10. Distribution Setup

### GoReleaser Configuration

Create a `.goreleaser.yml` file with the following configuration:

```yaml
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.Version={{.Version}}
    main: ./cmd/mockoho/main.go

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- .Os }}_
      {{- .Arch }}
    format_overrides:
      - goos: windows
        format: zip

release:
  github:
    owner: mockoho
    name: mockoho
```

### GitHub Actions Workflow

Create a `.github/workflows/release.yml` file:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.23"

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v4
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Installation Instructions

Update the README.md with clear installation instructions:

````markdown
## Installation

### Option 1: Download Binary (Recommended)

1. Download the latest release for your platform from [GitHub Releases](https://github.com/mockoho/mockoho/releases)
2. Extract the archive
3. Move the binary to a location in your PATH (optional)

### Option 2: Using Go Install (Requires Go)

```bash
go install github.com/mockoho/mockoho@latest
```
````

## Usage

Run Mockoho with default configuration:

```bash
mockoho
```

Specify a custom configuration directory:

```bash
mockoho --config /path/to/your/mocks
```

Run in server-only mode (without TUI):

```bash
mockoho server --config /path/to/your/mocks
```

```

By addressing these tasks before publishing, you'll ensure that Mockoho is robust, maintainable, and ready for users to install and run with a single command.
```
