# Contributing to Go Wallet SDK

Thank you for your interest in contributing to the Go Wallet SDK! This document provides guidelines and information for contributors.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)

## Getting Started

### Prerequisites

- **Go 1.24.0 or higher**: Check your version with `go version`
- **Git**: For version control
- **Make**: For build automation (optional but recommended)

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-wallet-sdk.git
   cd go-wallet-sdk
   ```

3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/status-im/go-wallet-sdk.git
   ```

4. **Install dependencies**:
   ```bash
   go mod download
   ```

5. **Verify your setup**:
   ```bash
   go test ./...
   ```

## Development Workflow

### Branching Strategy

- `master`: Main development branch
- `feature/*`: New features
- `fix/*`: Bug fixes
- `docs/*`: Documentation updates

### Making Changes

1. **Create a new branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the [Code Style](#code-style) guidelines

3. **Add tests** for your changes (see [Testing](#testing))

4. **Run tests locally**:
   ```bash
   go test ./...
   ```

5. **Run linter** (if available):
   ```bash
   golangci-lint run
   ```

6. **Commit your changes** with clear, descriptive messages:
   ```bash
   git commit -m "pkg/balance: add support for ERC1155 balance fetching"
   ```

### Commit Message Guidelines

Use conventional commit format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

**Example:**
```
feat(gas): add support for Linea gas estimation

- Implement LineaStack chain class
- Add Linea-specific gas price fetching
- Update DefaultConfig for Linea

Closes #123
```

## Code Style

### General Guidelines

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Keep functions small and focused (preferably under 50 lines)
- Use descriptive variable and function names
- Add comments for exported functions, types, and packages

### Package Organization

```
pkg/
└── yourpackage/
    ├── doc.go           # Package documentation
    ├── yourpackage.go   # Main implementation
    ├── types.go         # Type definitions
    ├── errors.go        # Error definitions
    ├── yourpackage_test.go  # Tests
    ├── mock/            # Mock implementations
    └── README.md        # Package-specific documentation
```

### Documentation Comments

All exported identifiers must have documentation comments:

```go
// FetchBalances fetches native token balances for multiple addresses using
// Multicall3 batching. It automatically falls back to standard RPC calls if
// Multicall3 is not available on the chain.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - addresses: List of addresses to fetch balances for
//   - atBlock: Block number (nil for latest)
//   - rpcClient: RPC client implementing the RPCClient interface
//   - batchSize: Number of addresses per batch
//
// Returns a map of addresses to balances in Wei.
func FetchBalances(ctx context.Context, addresses []common.Address, ...) (map[common.Address]*big.Int, error) {
    // Implementation
}
```

### Error Handling

- Use sentinel errors for known error conditions:
  ```go
  var (
      ErrInvalidAddress = errors.New("invalid ethereum address")
      ErrNoBalances     = errors.New("no balances found")
  )
  ```

- Wrap errors with context:
  ```go
  if err != nil {
      return fmt.Errorf("failed to fetch balance for %s: %w", addr, err)
  }
  ```

### Naming Conventions

- **Packages**: Short, lowercase, single-word names (e.g., `fetcher`, `gas`, `multicall`)
- **Interfaces**: Noun or verb ending in -er (e.g., `Fetcher`, `Parser`, `Builder`)
- **Constructors**: Use `New` or `NewWithConfig` (e.g., `func New() *Fetcher`)

## Testing

### Writing Tests

- Write tests for all new functionality
- Aim for at least 80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

**Example table-driven test:**

```go
func TestFetchBalances(t *testing.T) {
    tests := []struct {
        name      string
        addresses []common.Address
        want      map[common.Address]*big.Int
        wantErr   bool
    }{
        {
            name:      "single address",
            addresses: []common.Address{addr1},
            want:      map[common.Address]*big.Int{addr1: big.NewInt(1000)},
            wantErr:   false,
        },
        {
            name:      "multiple addresses",
            addresses: []common.Address{addr1, addr2},
            want: map[common.Address]*big.Int{
                addr1: big.NewInt(1000),
                addr2: big.NewInt(2000),
            },
            wantErr: false,
        },
        {
            name:      "empty addresses",
            addresses: []common.Address{},
            want:      nil,
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FetchBalances(context.Background(), tt.addresses, ...)
            if (err != nil) != tt.wantErr {
                t.Errorf("FetchBalances() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("FetchBalances() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run tests for a specific package
go test ./pkg/balance/fetcher/...

# Run a specific test
go test -run TestFetchBalances ./pkg/balance/fetcher/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Benchmark Tests

Add benchmarks for performance-critical code:

```go
func BenchmarkFetchBalances(b *testing.B) {
    // Setup
    addresses := generateAddresses(100)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := FetchBalances(context.Background(), addresses, ...)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

Run benchmarks:
```bash
go test -bench=. ./pkg/balance/fetcher/...
```

## Documentation

### Package Documentation

Every package should have:

1. **doc.go**: Package-level documentation
2. **README.md**: Detailed usage guide with examples
3. **Inline comments**: For all exported identifiers

### README Structure

Each package README should follow this structure:

```markdown
# Package Name

Brief description of what the package does.

## Use it when

- Bullet points explaining when to use this package

## Key entrypoints

- List of main functions/types to start with

## Overview

Detailed explanation of the package functionality.

## Key Features

- Feature 1
- Feature 2

## Quick Start

\```go
// Simple usage example
\```

## API Reference

### Types

### Functions

### Examples

## Integration

How to use with other packages

## Testing

How to test

## Performance Considerations

## Common Pitfalls

## See Also

Links to related packages
```

### Example Code

- All examples must compile and run
- Use real-world scenarios
- Include error handling
- Add comments explaining key steps

## Pull Request Process

### Before Submitting

1. ✅ All tests pass locally
2. ✅ Code follows style guidelines
3. ✅ Documentation is updated
4. ✅ Commit messages are clear
5. ✅ Branch is up to date with `master`

### Submitting Your PR

1. **Push your branch** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Open a Pull Request** on GitHub

3. **Fill out the PR template**:
   - Description of changes
   - Related issues
   - Testing performed
   - Screenshots (if UI changes)

4. **Respond to review feedback** promptly

### PR Title Format

```
<type>(<scope>): <description>
```

Example: `feat(gas): add Linea gas estimation support`

### Review Process

- At least two maintainer approvals required
- All CI checks must pass
- Address all review comments
- Keep PRs focused and reasonably sized (< 500 lines when possible)

## Issue Reporting

### Before Creating an Issue

1. **Search existing issues** to avoid duplicates
2. **Check the documentation** for your question
3. **Try the latest version** to see if it's already fixed

### Creating a Good Issue

Use the appropriate issue template:

**Bug Report:**
- Clear title
- Steps to reproduce
- Expected behavior
- Actual behavior
- Go version and OS
- Code sample (minimal reproducible example)

**Feature Request:**
- Clear description of the feature
- Use cases and motivation
- Proposed API (if applicable)
- Alternatives considered

**Question:**
- What you're trying to accomplish
- What you've tried
- Relevant code snippets

## Development Tips

### Useful Commands

```bash
# Format code
go fmt ./...

# Update dependencies
go get -u ./...
go mod tidy

# Generate mocks (if using gomock)
go generate ./...

# Build C library
make shared-library
make static-library

# Clean build artifacts
make clean
```

### Working with Examples

When adding or modifying examples:

1. Place in `examples/<example-name>/`
2. Include a `README.md` with:
   - Description
   - How to run
   - Expected output
   - Prerequisites (API keys, etc.)
3. Add a `go.mod` file
4. Keep examples simple and focused

### Debugging Tips

- Use `-v` flag for verbose test output
- Use `t.Logf()` for debug output in tests
- Run single tests during development: `go test -run TestSpecificTest`
- Use delve debugger: `dlv test ./pkg/yourpackage`

## Getting Help

- **GitHub Issues**: For bugs and feature requests
- **Discussions**: For questions and general discussion
- **Documentation**: Check package READMEs and GoDoc

## License

By contributing, you agree that your contributions will be licensed under the Mozilla Public License Version 2.0.

---

Thank you for contributing to Go Wallet SDK! 🚀
