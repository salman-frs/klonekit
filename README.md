# KloneKit

KloneKit is a CLI tool that helps DevOps engineers provision infrastructure and set up GitLab projects using blueprint configurations.

## Development

This project uses a Makefile to standardize development tasks. Use the following commands for all development activities:

### Building from Source

To build the application:
```bash
make build
```

To run tests:
```bash
make test
```

To lint the codebase:
```bash
make lint
```

To clean build artifacts:
```bash
make clean
```

To see all available commands:
```bash
make help
```

### Prerequisites

- Go 1.22.x or higher
- golangci-lint (will be installed automatically when running `make lint`)

### Project Structure

The project follows Go best practices with a clear separation of concerns:

- `cmd/klonekit/` - CLI entry point and command definitions
- `internal/` - Private application code
  - `app/` - Core orchestrator logic
  - `parser/` - Blueprint parsing and validation
  - `provisioner/` - Terraform-via-Docker logic
  - `scaffolder/` - File generation and manipulation
  - `scm/` - GitLab integration
- `pkg/blueprint/` - Shared data models
- `test/e2e/` - End-to-end tests

## Usage

[Usage documentation would be added in future stories]

## License

[License information would be added in future stories]