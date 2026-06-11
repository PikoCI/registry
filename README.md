# PikoCI Registry

> **Warning:** This project is under active development and not yet ready for production use. APIs, CLI commands, and data formats may change without notice.

A versioned, searchable, tagged registry for PikoCI's five pluggable abstraction types (resource_type, runner_type, service_type, secret_type, notification_type).

## Quickstart

```bash
# Development with SQLite
make serve

# With PostgreSQL
docker-compose -f docker/docker-compose.yml up -d
go run . server --db-system=postgresql --db-host=localhost --db-port=5432 --db-user=registry --db-password=registry --db-name=registry --jwt-secret=dev-secret
```

## CLI

```bash
# Login via GitHub OAuth
pikoci-registry login

# Publish a type
pikoci-registry publish ./my-resource-type

# Search
pikoci-registry search postgres --kind=resource_type

# List types in a namespace
pikoci-registry list myorg

# Get type details
pikoci-registry info myorg/postgres
```

## Development

```bash
make help          # Show all targets
make test          # Run all tests
make test-mock     # Run unit tests only
make lint          # Run staticcheck
make gen           # Run go generate
make serve         # Start dev server
```
