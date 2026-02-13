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
- `SHOPMON_API_KEY`: Authorization token for the monitoring service (required, sent as Bearer token)
- `SHOPMON_SHOP_ID`: Shop ID to include in telemetry
- `SHOPMON_DEPLOYMENT_VERSION_REFERENCE`: Override the version reference (defaults to `git rev-parse HEAD`)

```bash
# With custom URL
SHOPMON_BASE_URL="http://localhost:8080" \
SHOPMON_API_KEY="your-secret-token" \
./shopmon-cli deploy -- php artisan migrate

# With authorization token
SHOPMON_API_KEY="your-secret-token" ./shopmon-cli deploy -- composer install
```

## Telemetry Payload

The deploy command sends the following JSON payload to the monitoring service endpoint (`/trpc/cli.createDeployment`):

```json
{
  "shop_id": 1,
  "command": "php foo",
  "return_code": 0,
  "start_date": "2024-06-01T12:00:00Z",
  "end_date": "2024-06-01T12:00:01Z",
  "execution_time": 1.23,
  "composer": {
    "php": ">=8.1",
    "shopware/core": "6.5.0.0"
  },
  "reference": "abc123..."
}
```

When `SHOPMON_API_KEY` is set, the request includes an `Authorization: Bearer <token>` header for authentication.

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
│       └── deployment.go   # Main orchestration
├── main.go
├── go.mod
└── go.sum
```

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

1. **cmd/**: CLI command definitions using Cobra
2. **internal/deployment/**: Core logic split into:
   - **executor.go**: Runs the user's command and captures output, return code, and timing
   - **composer.go**: Reads `composer.json` to extract dependency info
   - **telemetry.go**: Sends deployment data to the monitoring service, uploads compressed output
   - **deployment.go**: Orchestrates the above into a single `Run()` function

### Key Design Decisions

- **Transparent Execution**: Command output is displayed to the user
- **Error Resilience**: Telemetry failures don't affect command execution
- **Exit Code Preservation**: The CLI exits with the same code as the executed command

## License

MIT
