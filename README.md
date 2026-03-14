# Shopmon CLI

A command-line tool for monitoring and managing Shopware applications with deployment telemetry.

## Features

- Track deployment commands and their execution details to Shopmon monitoring service

## Installation

### Composer

```bash
composer require frosh/shopmon-cli

# Downloads on the fly shopmon-cli and executes the command
vendor/bin/shopmon-cli deploy -- <command>
```

### Docker

```dockerfile
COPY --from=ghcr.io/friendsofshopware/shopmon-cli:0.0.6 /shopmon-cli /usr/local/bin/shopmon-cli
```

## Usage

### Basic Usage

```bash
./shopmon-cli deploy -- <command>
```

### Examples

```bash
# Execute a simple command
./shopmon-cli deploy -- vendor/bin/shopware-deployment-helper run
```

### Environment Variables

- `SHOPMON_BASE_URL`: Override the default monitoring service URL (default: `https://shopmon.fos.gg`)
- `SHOPMON_API_KEY`: Authorization token for the monitoring service (required, sent as Bearer token)
- `SHOPMON_SHOP_ID`: Shop ID to include in telemetry
- `SHOPMON_DEPLOYMENT_NAME`: Custom name for the deployment (if not set, the server generates a random name)
- `SHOPMON_DEPLOYMENT_VERSION_REFERENCE`: Override the version reference (defaults to `git rev-parse HEAD`)

```bash
# With custom URL
SHOPMON_BASE_URL="http://localhost:8080" \
SHOPMON_API_KEY="your-secret-token" \
./shopmon-cli deploy -- vendor/bin/shopware-deployment-helper run

# With authorization token
SHOPMON_API_KEY="your-secret-token" ./shopmon-cli deploy -- vendor/bin/shopware-deployment-helper run
```

## License

MIT
