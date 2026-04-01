# Contributing to 01cloud API

Thanks for your interest in contributing to `01cloud-api`. This guide is meant to help outside contributors get productive quickly and submit changes that are easy to review.

## Before You Start

- Read the [Code of Conduct](CODE_OF_CONDUCT.md)
- Check existing issues and pull requests before starting duplicate work
- Open an issue first for large features, behavior changes, or refactors that may affect maintainers' roadmap

## Development Setup

### Requirements

- Go 1.22+
- PostgreSQL 12+
- Docker for container-based lint/test targets
- Redis and RabbitMQ if your work touches cache- or queue-backed features

### Local Setup

```bash
git clone https://github.com/01cloud/01cloud-api.git
cd 01cloud-api
cp .env.sample .env
go mod download
make run-dev
```

If you do not yet have all optional services running, focus your changes on areas that can be tested locally and document any known setup gaps in your PR.

## Common Commands

```bash
make run-dev
make test-dev
make format
make lint
make unit-test
make unit-test-coverage
```

## Contribution Workflow

1. Fork the repository and branch from `main`.
2. Choose a descriptive branch name such as `fix/session-timeout` or `docs/readme-setup`.
3. Make the smallest change that solves the problem cleanly.
4. Add or update tests when the change affects behavior.
5. Update documentation when APIs, config, or workflows change.
6. Run the relevant checks before opening your PR.
7. Open a pull request with context, testing notes, and linked issues.

## Pull Request Expectations

Please aim for pull requests that are easy to review:

- Keep the scope focused on one problem
- Explain the user or maintainer impact
- Link related issues using `Fixes #123` when applicable
- Include screenshots or request/response examples when the change benefits from them
- Call out follow-up work explicitly instead of sneaking unrelated cleanup into the same PR

## Code Style

- Format Go code with `gofmt` via `make format`
- Keep functions and handlers readable; prefer clear naming over clever shortcuts
- Add tests for new logic where practical
- Avoid unrelated refactors in the same PR unless they are required to support the fix

## Commit Guidance

- Use short, imperative commit messages such as `Add registry validation` or `Fix Redis cache nil handling`
- Keep the first line concise
- Reference issues in the commit body when it adds useful context

## Reporting Bugs

When opening a bug report, include:

- What happened
- What you expected to happen
- Clear reproduction steps
- Relevant environment details
- Logs, screenshots, or sample payloads when available

Use the provided GitHub issue template whenever possible.

## Suggesting Features

Feature requests are most useful when they explain:

- The problem being solved
- Why existing behavior is insufficient
- The rough shape of the proposed solution
- Any alternatives already considered

## Security Reports

Do not open public issues for vulnerabilities. Please follow the process in [SECURITY.md](SECURITY.md).
