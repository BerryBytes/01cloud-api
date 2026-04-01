<p align="center">
  <h1 align="center">01cloud API</h1>
  <p align="center">Backend API for the 01cloud platform, focused on Kubernetes operations, application delivery, infrastructure automation, and multi-tenant platform management.</p>
</p>

<p align="center">
  <a href="https://github.com/01cloud/01cloud-api/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.22-00ADD8.svg?logo=go" alt="Go Version"></a>
  <a href="https://github.com/01cloud/01cloud-api/actions"><img src="https://img.shields.io/github/actions/workflow/status/01cloud/01cloud-api/ci.yaml?branch=main&label=CI" alt="CI Status"></a>
  <a href="https://goreportcard.com/report/github.com/01cloud/01cloud-api"><img src="https://goreportcard.com/badge/github.com/01cloud/01cloud-api" alt="Go Report Card"></a>
</p>

## Overview

`01cloud-api` is the core backend service for the [01cloud](https://01cloud.com) platform. It exposes REST APIs and background-processing integrations used to manage Kubernetes clusters, deploy workloads, connect registries and Git providers, orchestrate CI/CD workflows, and handle organization-level platform administration.

This repository is intended to be usable by open-source contributors. If you are exploring the codebase for the first time, start with the quickstart below, then review the contribution and security guides before opening a PR.

## What This Project Does

- Manages cluster and environment lifecycle operations
- Supports application deployment workflows and Helm-based package delivery
- Integrates with Git providers such as GitHub, GitLab, and Bitbucket
- Provides organization, user, role, and access-management APIs
- Connects to external services including PostgreSQL, Redis, RabbitMQ, SMTP, and gRPC services
- Exposes Swagger/OpenAPI documentation for API exploration

## Repository Layout

```text
01cloud-api/
├── api/
│   ├── controllers/      # HTTP handlers and route logic
│   ├── models/           # Data models and persistence logic
│   ├── middlewares/      # Auth and request middleware
│   ├── mailer/           # Email delivery helpers
│   ├── notifications/    # Notification service integration
│   ├── registry/         # Container registry helpers
│   ├── seed/             # Seed data utilities
│   ├── utils/            # Shared helpers and provider integrations
│   ├── websocket/        # WebSocket support
│   └── server.go         # API server bootstrapping
├── docs/                 # Swagger-generated API assets
├── .github/              # CI workflows and contribution templates
├── Dockerfile            # Multi-stage container build
├── Makefile              # Common development tasks
└── main.go               # Application entrypoint
```

## Prerequisites

- Go 1.22 or newer
- PostgreSQL 12 or newer
- Redis for cache-backed features
- RabbitMQ for queue-backed/background features
- Docker if you prefer the containerized workflow

Some features also expect external services such as SMTP and gRPC notification/backup services. For local development, you can often leave optional values unset until you work on those areas.

## Quickstart

```bash
git clone https://github.com/01cloud/01cloud-api.git
cd 01cloud-api
cp .env.sample .env
go mod download
go run main.go
```

The API starts on `http://localhost:8081` by default.

## Configuration

Configuration is driven by environment variables in `.env`.

Important variables include:

| Variable | Description | Default |
| --- | --- | --- |
| `API_SECRET` | JWT signing secret | none |
| `DB_HOST` | PostgreSQL host | `127.0.0.1` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL username | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | none |
| `DB_NAME` | PostgreSQL database name | `fullstack_api` |
| `BASE_URL` | Public API base URL | `http://localhost:8081` |
| `SMTP_HOST` | SMTP server hostname | none |
| `GRPC_NOTIFICATION_SERVER` | Notification gRPC endpoint | `localhost:10080` |
| `GRPC_BACKUP_SERVER` | Backup gRPC endpoint | `localhost:10080` |
| `CACHE_TYPE` | Cache backend selection | none |
| `REDIS_SERVER_URL` | Redis connection string | none |
| `DEBUG` | Enables debug logging when `true` | `false` |

See [`.env.sample`](.env.sample) for the broader list used by the application and test suite.

## Development Workflow

Useful local commands:

```bash
make run-dev             # start the API locally
make test-dev            # run tests with coverage output
make format              # gofmt all Go files
make lint                # run linting in Docker
make unit-test           # run unit tests in Docker
```

The Makefile uses Docker for some CI-style checks and local Go tooling for others. If you are contributing from a fresh machine, install Go first and use Docker-backed targets when you want a closer match to CI.

## API Documentation

Swagger assets live in [`docs/`](docs). With the server running locally, the interactive docs are available at:

```text
http://localhost:8081/swagger/index.html
```

To regenerate Swagger files:

```bash
swag init
```

## Testing

```bash
go test ./...
make test-dev
make unit-test
make unit-test-coverage
```

If your change affects API behavior, request handling, or integrations, add or update tests where practical.

## Open Source Contribution

- Read the [contribution guide](CONTRIBUTING.md) before opening a PR
- Review the [code of conduct](CODE_OF_CONDUCT.md) for community expectations
- Use the [security policy](SECURITY.md) for responsible vulnerability reporting
- Use the [support guide](SUPPORT.md) for questions and help requests

## License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.
