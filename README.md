# SCD Abstraction in Go with GORM

## Setup

1. **Spin up Postgres** (Docker Compose example):

```
docker-compose up -d
```

2. **Install dependencies:**

```
go mod tidy
```

3. **Run migrations and seed data:**

```
go run main.go
```

4. **Run benchmarks:**

```
go test -bench . ./benchmark
```

## Performance Analysis

- Wrap repo calls with timers and log to CSV for latency analysis.
- Use `EXPLAIN (ANALYZE, BUFFERS)` in Postgres for query plans.
- Add indexes as needed for performance.
- Use `hey` or `wrk` for load testing.
- Use Go's `pprof` for CPU/memory profiling.

## Repo Structure

- `models/` - GORM models
- `repos/` - Repository pattern and helpers
- `benchmark/` - Benchmarks and seeding
- `main.go` - Entry point
- `README.md` - This file 