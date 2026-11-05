# Changelog

## 2026-11 — refactoring

- Fixed SQL injection: all queries now use placeholders instead of string formatting
- Fixed swapped title/announcement caused by `SELECT *` and column order; queries list columns explicitly
- Errors are handled and logged instead of `panic`; proper 400/404/405/500 responses
- Database connection is opened once at startup with a connection pool and startup retry
- Templates are parsed once at startup and embedded with static files into the binary
- Switched routing from gorilla/mux to the Go 1.22 standard library ServeMux; removed unused dependencies
- Config: JSON file plus environment variable overrides; added `cfg/db.cfg.example`
- Form validation (required fields, length limits) with error message on the form
- Graceful shutdown, HTTP server timeouts, `/healthz` endpoint, `log/slog` logging
- Schema: `INT UNSIGNED` id (was `TINYINT`, max 255 rows), utf8mb4, no destructive `DROP TABLE`
- Removed unused `init/mysql.go`; fixed relative CSS path on article pages; removed dead `/contact` link
- Added unit and integration tests, Dockerfile, docker compose stack and GitHub Actions CI (amd64 + arm64 builds)

## 2021-04 — 2021-05 — initial version

 - added mysql module
 - created initial db schema file
 - added html page with form to add new article
 - added link to add article page
 - implemented showing of all article list in index page
 - implemented function to read articles from db
 - implemented storing articles functionality
