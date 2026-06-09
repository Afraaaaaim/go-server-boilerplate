# go-server-boilerplate

A production-ready Go API server template.

## What's included

- HTTP server with [Chi](https://github.com/go-chi/chi) router
- Structured logging (`log/slog`) with OpenTelemetry bridge
- API key authentication middleware
- Rate limiting middleware
- Standardized JSON error responses
- Request ID propagation
- Graceful shutdown
- Docker + Docker Compose (app, Redis, OTel Collector, Jaeger)

## Quick start

```bash
cp .env.example .env
# edit .env and add your API keys
go run ./cmd/api
```

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server port |
| `ENV` | `development` | Environment name |
| `LOG_LEVEL` | `info` | `debug` `info` `warn` `error` |
| `LOG_FORMAT` | `text` | `text` or `json` |
| `LOG_FILE` | stdout | File path for logs |
| `API_KEYS` | — | Comma-separated valid API keys |
| `RATE_LIMIT_RPS` | `100` | Requests per second |
| `RATE_LIMIT_BURST` | `20` | Burst size |
| `OTLP_ENDPOINT` | disabled | OTel collector endpoint e.g. `localhost:4317` |
| `SERVICE_NAME` | `go-server-boilerplate` | Service name in traces/logs |

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/healthz` | No | Liveness probe |
| GET | `/readyz` | No | Readiness probe |
| GET | `/api/example` | Yes | Example protected route |

## Auth

Pass your API key in either header:

```
X-API-Key: your-key
Authorization: Bearer your-key
```

## Run with Docker

```bash
docker compose -f docker/docker-compose.yml up --build -d
```

Jaeger UI available at `http://localhost:16686`

## Make targets

```
make run          run locally
make build        build binary to ./bin/
make test         run tests
make tidy         go mod tidy
make docker-up    start full stack
make docker-down  stop stack
make docker-logs  tail app logs
```