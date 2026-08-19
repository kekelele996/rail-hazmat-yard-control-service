# Rail Hazmat Yard Control Service

A Go service for railway hazardous-material yard operations. It coordinates consist manifests, inspection crews, movement permits, clearance deadlines, telemetry ingestion, incident transactions, dispatch state, track occupancy, and audit event delivery.

## Structure

- `cmd/server`: HTTP service entrypoint.
- `internal/*`: domain modules and memory-backed operational services.
- `internal/policy`: reusable industry policy evaluators.
- `web`: a small operations status page served by the Go process.

## Run

```bash
go run ./cmd/server
```

The service listens on `PORT` (default `18080`). Health is available at `/health`; `/api/summary` returns an operational summary.

## Test

```bash
go test ./...
go test -race ./...
```

## Environment

- `PORT`: HTTP port, defaults to `18080`.
- `YARD_CODE`: logical yard identifier, defaults to `NORTH-TRANSFER`.
