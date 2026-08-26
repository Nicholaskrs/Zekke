# Zekke

A boilerplate for Go backend services — structured project layout, config management, structured logging, OpenTelemetry tracing/metrics, JWT + internal API-key auth, GORM (MySQL/Postgres), and S3-compatible storage, ready to build on.

## Prerequisites
- Go 1.25+
- MySQL or PostgreSQL

## Project Structure
```
.
├── api/            # Route/handler definitions grouped by domain (auth, health, user)
├── base/           # Shared base DTOs, constants, and helpers (string, time, encryption, pagination, files, json)
├── core/           # App wiring: dependency injection (dependencies.go), DB engine, telemetry (logger/metric/trace)
├── data/
│   ├── enum/       # Domain enums
│   └── model/      # GORM models
├── modules/        # Business logic per domain, each split into handler → svc → repository
│   ├── health/
│   ├── notification/   # Push notification service (Firebase) — not yet wired into core/dependencies.go
│   └── user/
├── server/
│   ├── middleware/ # JWT auth, internal API-key auth
│   └── router/     # Route registration
├── util/
│   ├── config/     # Viper-based env config loader
│   ├── firebase/   # Firebase Cloud Messaging client
│   ├── logtrace/   # Request/trace-id propagation helper
│   └── storage/    # S3-compatible storage client
├── config.env.example
├── go.mod / go.sum
└── main.go         # Entry point; also handles `migrate` / `seed` CLI subcommands
```

## Setup
1. Clone the project
2. Copy the environment file and fill in the values:
    ```
    cp config.env.example config.env
    ```
3. Install dependencies:
    ```
    go mod tidy
    ```

## Run
From the project root:
```
go run main.go
```
The server starts on the port defined by `SERVER_PORT` (default `:9000`).

## Database
- Migrate schema:
    ```
    go run main.go migrate
    ```
- Seed database:
    ```
    go run main.go seed
    ```
- Migrate and seed together:
    ```
    go run main.go migrate seed
    ```
- Both MySQL and PostgreSQL are supported — set `DB_DRIVER` to `mysql` or `postgres`.

## Environment Variables

| Variable | Description |
| --- | --- |
| `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT` | Database connection |
| `DB_DRIVER` | `mysql` or `postgres` |
| `DB_TIMEZONE` | Timezone used for DB connections |
| `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS` | Connection pool tuning |
| `JWT_SECRET`, `JWT_ISSUER` | JWT auth config |
| `API_KEY` | Key required on `X-API-Key` for internal-only routes |
| `HOST_DOMAIN` | Allowed frontend origin(s) — currently unused since CORS defaults to allow-all (see Notes) |
| `SERVER_PORT` | Port the server listens on, e.g. `:9000` |
| `LOG_TYPE` | 	`1` = log all SQL queries (GORM query logger set to Info level); `0` = SQL logging off |
| `STORAGE_REGION`, `STORAGE_BUCKET`, `STORAGE_BASE_PATH`, `STORAGE_BASE_URL` | S3-compatible storage config |
| `OTEL_TRACE_ENDPOINT` | OTLP gRPC collector endpoint |
| `OTEL_METRICS_ENABLED` | Enable/disable OpenTelemetry metrics |

## API Routes
- `/api/v1/*` — public + JWT-authenticated routes (health, user)
- `/api-internal/v1/*` — requires `X-API-Key` header, for internal/service-to-service use (user admin routes)

## Notes
- All timestamps use `time.Now()`, so behavior depends on the host machine's local time.
- CORS is currently configured with `AllowAllOrigins: true` and `AllowCredentials: true` in `main.go`. **Before deploying to production**, switch to `AllowOrigins` using `HOST_DOMAIN` (comma-separated list of allowed origins) — see the TODO comment in `main.go`.
- The `notification`/Firebase module (`modules/notification`, `util/firebase`) is scaffolded but not yet wired into `core/dependencies.go`. To use it, add a Firebase service-account path to `Config`, instantiate `firebase.NewFirebaseClient(...)`, and pass it into `notification.NewNotificationService(...)` in `InitClient`.
