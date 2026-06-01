# Contributing

Thank you for contributing to Shopmon CLI. This project is a small Go command-line tool that is distributed as a native binary, Docker image, and Composer package wrapper.

## Getting Started

1. Fork the repository and create a branch from `main`.
2. Install Go `1.26` or newer.
3. Download dependencies:

   ```bash
   go mod download
   ```

4. Run the test suite:

   ```bash
   go test ./...
   ```

## Development

- Keep changes focused on one problem or feature.
- Prefer small pull requests with a clear description of the behavior change.
- Add or update tests when changing behavior.
- Run `gofmt` on changed Go files before opening a pull request.
- Keep user-facing documentation in `README.md` in sync with CLI behavior and environment variables.

## Local Usage

Build the CLI locally:

```bash
go build -o shopmon-cli .
```

Run a deployment command through the CLI:

```bash
SHOPMON_API_KEY="your-token" ./shopmon-cli deploy -- vendor/bin/shopware-deployment-helper run
```

Use `SHOPMON_BASE_URL` when testing against a local or staging Shopmon service.

## Pull Requests

Before requesting review, make sure:

- `go test ./...` passes.
- Any relevant documentation is updated.
- The pull request explains the motivation, implementation, and testing performed.

Maintainers may squash or reword commits when merging.

## Releases

Releases are created by maintainers by pushing tags. GoReleaser builds the native archives and Docker image from the tagged commit.
