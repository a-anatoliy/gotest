# gotest

A small article board written in Go: list articles, read one, add a new one.
Data is stored in MySQL. It started as my first Go project and was later
cleaned up to follow common Go practices.

## Features

- HTTP server on the standard library (`net/http`, Go 1.22 routing patterns such as `GET /post/{id}`)
- HTML pages with `html/template` (automatic escaping of user content)
- MySQL access through `database/sql` with parameterized queries only
- Templates and static files embedded into the binary (`embed`)
- Configuration from a JSON file and/or environment variables
- Input validation, proper HTTP status codes (400, 404, 405, 500), structured logging (`log/slog`)
- Graceful shutdown on SIGINT/SIGTERM, server timeouts, `/healthz` endpoint
- Unit tests for handlers and config; integration tests against a real MySQL
- Docker image (multi-stage, distroless, non-root) and a docker compose stack
- GitHub Actions: gofmt, go vet, tests with a MySQL service container, builds for linux/amd64 and linux/arm64

## Run with Docker

```sh
docker compose up --build
```

Then open http://localhost:8080. The schema from `init/db_schema.sql` is loaded automatically.

## Run locally

1. Create a database and load the schema:

   ```sh
   mysql -u <user> -p <db_name> < init/db_schema.sql
   ```

2. Copy the example config and fill it in (`cfg/db.cfg` is git-ignored):

   ```sh
   cp cfg/db.cfg.example cfg/db.cfg
   ```

3. Start the app:

   ```sh
   go run . -config cfg/db.cfg
   ```

Every setting can also be given as an environment variable, which takes
precedence over the file: `DB_USERNAME`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`,
`DB_NAME`, `LISTEN_ADDR` (default `:8080`).

## Tests

```sh
# unit tests (no database needed)
go test -race ./...

# integration tests (need MySQL with the schema loaded)
DB_USERNAME=... DB_PASSWORD=... DB_NAME=... go test -tags integration ./...
```

## Build for ARM64

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o gotest-linux-arm64 .
# or a multi-arch image:
docker buildx build --platform linux/amd64,linux/arm64 -t gotest .
```

## Layout

| File | Purpose |
|---|---|
| `main.go` | startup, DB connection, HTTP server, graceful shutdown |
| `config.go` | config file + environment variables, MySQL DSN |
| `store.go` | `Store` interface and the MySQL implementation |
| `server.go` | routes, handlers, validation, template rendering |
| `*_test.go` | unit and integration tests |
| `templates/`, `static/` | HTML templates and CSS (embedded) |
| `init/db_schema.sql` | database schema |
