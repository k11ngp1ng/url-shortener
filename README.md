# URL Shortener API

A production-oriented RESTful URL shortening API built with **Go**, designed to create short links, handle redirects, track basic access statistics, and expose operational metrics.

The project demonstrates backend API development with **PostgreSQL persistence, Docker, database migrations, rate limiting, structured logging, Prometheus metrics, and CI with GitHub Actions**.

## Features

- Create short URLs from long URLs
- Redirect short codes to their original URLs
- Track the number of accesses for each shortened URL
- Optional URL expiration using RFC 3339 timestamps
- PostgreSQL persistence
- In-memory storage fallback for local development
- Versioned SQL database migrations
- Per-IP rate limiting
- Structured application logging
- Health check endpoint
- Prometheus-compatible metrics
- Docker and Docker Compose support
- Automated CI pipeline with GitHub Actions

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/urls` | Creates a shortened URL |
| `GET` | `/{code}` | Redirects to the original URL |
| `GET` | `/api/v1/urls/{code}` | Returns URL information and access count |
| `GET` | `/health` | Checks API and database health |
| `GET` | `/metrics` | Exposes Prometheus-compatible metrics |

## Usage Example

Create a shortened URL:

```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev","expires_at":"2030-01-01T00:00:00Z"}'
```

The `expires_at` field is optional and must contain a future timestamp in **RFC 3339** format.

Once a URL expires, redirect requests return:

```text
410 Gone
```

Expired URLs do not increment the access counter.

## Running Locally

### Prerequisites

- Go 1.23+

Run the API:

```bash
go run ./cmd/api
```

Run the test suite:

```bash
go test ./...
```

If `DATABASE_URL` is not configured, the API automatically uses in-memory storage.

> In-memory storage is intended for development and testing only. Data is lost when the application stops.

## Docker and PostgreSQL

The application can be started with PostgreSQL using Docker Compose:

```bash
docker compose up --build
```

The `migrate` service automatically applies the versioned SQL migrations located in:

```text
migrations/
```

before the API starts.

To stop the containers and remove the local database volume:

```bash
docker compose down -v
```

Then start the environment again with:

```bash
docker compose up --build
```

## Configuration

The application can be configured using environment variables.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port used by the API |
| `DATABASE_URL` | Empty | PostgreSQL connection string; enables persistent storage |
| `RATE_LIMIT_PER_MINUTE` | `60` | Maximum requests allowed per IP per minute; `0` disables rate limiting |

### Example

```bash
PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/urlshortener
RATE_LIMIT_PER_MINUTE=60
```

## Storage

The application supports two storage modes.

### In-Memory

When `DATABASE_URL` is not configured, the application uses an in-memory repository.

This mode is useful for:

- Local development
- Automated tests
- Quick experimentation

Data is not persisted between application restarts.

### PostgreSQL

When `DATABASE_URL` is configured, the application uses PostgreSQL for persistent storage.

Database schema changes are managed through versioned SQL migrations stored in:

```text
migrations/
```

## Rate Limiting

The API includes an in-memory per-IP rate limiter.

The default limit is:

```text
60 requests per minute per IP
```

When the limit is exceeded, the API responds with:

```text
429 Too Many Requests
```

and includes the `Retry-After` header.

Rate limiting can be disabled by setting:

```text
RATE_LIMIT_PER_MINUTE=0
```

## Observability

The application provides basic production observability through structured logs, health checks, and Prometheus-compatible metrics.

### Health Check

```http
GET /health
```

This endpoint can be used to verify the status of the API and its configured database.

### Metrics

```http
GET /metrics
```

The endpoint exposes application metrics in a format compatible with **Prometheus**.

## CI/CD

The project includes a **GitHub Actions** workflow that automatically validates the application.

The CI pipeline runs:

- `go vet`
- Automated tests
- Tests with Go's race detector
- Docker image build

This helps ensure that changes pushed to the repository continue to compile and pass the project's quality checks.

## Production-Oriented Features

The project includes several features commonly used in production backend services:

- PostgreSQL persistent storage
- Versioned SQL migrations
- Per-IP rate limiting
- `429 Too Many Requests` responses
- `Retry-After` headers
- Structured logging
- Health checks
- Prometheus metrics
- Docker containerization
- Docker Compose orchestration
- Automated CI validation
- Go race detector testing

## Tech Stack

- Go 1.23+
- PostgreSQL
- Docker
- Docker Compose
- Prometheus-compatible metrics
- GitHub Actions
- REST API

## Project Goals

This project was created to demonstrate practical experience with:

- RESTful API development in Go
- HTTP routing and redirects
- Repository-based persistence
- PostgreSQL integration
- Database migrations
- Containerized development environments
- API rate limiting
- Application observability
- Automated testing
- Continuous Integration
- Production-oriented backend development

## License

This project currently does not include a license.

Unless a license is added, the source code should not be assumed to be available for unrestricted reuse or redistribution.
