# Shopmon CLI

A command-line tool for monitoring and managing Shopware applications with deployment telemetry.

## Features

- Execute deployment commands with transparent output
- Capture command execution metrics (output, return code, execution time)
- Read composer.json dependencies and include in telemetry
- Send telemetry data to monitoring service
- Support for custom monitoring endpoint via environment variable
- Authorization token support for secure telemetry submission

## Usage

### Basic Usage

```bash
./shopmon-cli deploy -- <command>
```

### Examples

```bash
# Execute a simple command
./shopmon-cli deploy -- echo "Hello World"

# Run PHP artisan commands
./shopmon-cli deploy -- php artisan migrate

# Execute composer commands
./shopmon-cli deploy -- composer install

# Multiple word commands
./shopmon-cli deploy -- npm run build
```

### Environment Variables

- `SHOPMON_BASE_URL`: Override the default monitoring service URL (default: `https://shopmon.fos.gg`)
- `SHOPMON_DEPLOY_TOKEN`: Authorization token for the monitoring service (sent as Bearer token)

```bash
# With custom URL
SHOPMON_BASE_URL="http://localhost:8080" ./shopmon-cli deploy -- php artisan migrate

# With authorization token
SHOPMON_DEPLOY_TOKEN="your-secret-token" ./shopmon-cli deploy -- composer install

# With both URL and token
SHOPMON_BASE_URL="http://staging.example.com" \
SHOPMON_DEPLOY_TOKEN="staging-token" \
./shopmon-cli deploy -- npm run build
```

## Telemetry Payload

The deploy command sends the following JSON payload to the monitoring service endpoint (`/api/cli/deployment`):

```json
{
  "command": "php foo",
  "output": "the full output of the command",
  "return_code": 0,
  "start_date": "2024-06-01T12:00:00Z",
  "end_date": "2024-06-01T12:00:01Z",
  "execution_time": 1.23,
  "composer": {
    "php": ">=8.1",
    "shopware/core": "6.5.0.0"
  }
}
```

When `SHOPMON_DEPLOY_TOKEN` is set, the request includes an `Authorization: Bearer <token>` header for authentication.

## Development

### Project Structure

```
shopmon-cli/
├── cmd/                    # Command definitions
│   ├── root.go
│   └── deploy.go
├── internal/
│   └── deployment/         # Core deployment logic
│       ├── types.go        # Data structures
│       ├── executor.go     # Command execution
│       ├── composer.go     # Composer.json parsing
│       ├── telemetry.go    # Telemetry client
│       ├── deployment.go   # Main service orchestration
│       └── *_test.go       # Unit tests
├── main.go
├── go.mod
└── go.sum
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v ./internal/deployment/... -cover

# Run tests with coverage report
go test -v ./internal/deployment/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Current test coverage: **93.5%**

### Building

```bash
# Build the binary
go build -o shopmon-cli .

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o shopmon-cli-linux
GOOS=darwin GOARCH=amd64 go build -o shopmon-cli-darwin
GOOS=windows GOARCH=amd64 go build -o shopmon-cli.exe
```

## Architecture

The application follows clean architecture principles with clear separation of concerns:

1. **cmd/**: Contains the CLI command definitions using Cobra
2. **internal/deployment/**: Core business logic separated into:
   - **Interfaces**: `CommandExecutor`, `ComposerReader`, `TelemetryClient` for dependency injection and testing
   - **Service**: Main orchestration layer that coordinates between components
   - **Implementations**: Default implementations of interfaces
   - **Mocks**: Test doubles for unit testing

### Key Design Decisions

- **Dependency Injection**: All major components are injected as interfaces, making the code highly testable
- **Transparent Execution**: Command output is displayed in real-time to the user
- **Error Resilience**: Telemetry failures don't affect command execution
- **Exit Code Preservation**: The CLI exits with the same code as the executed command

## Testing

The codebase includes comprehensive unit tests using the testify library:

- Command parsing and execution
- Composer.json reading and parsing
- Telemetry payload creation and sending
- Service orchestration
- Mock implementations for all interfaces

## License

MIT