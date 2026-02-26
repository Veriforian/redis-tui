# Docker Compose Examples

All four configurations use different ports so they can run simultaneously.

## Standalone

`redis:7-alpine` on port `6379`:

```sh
docker compose -f examples/standalone/docker-compose.yml up -d
redis-tui
```

## Standalone (Redis Stack)

`redis-stack-server` on port `6390` (includes RedisJSON, RediSearch, and more):

```sh
docker compose -f examples/standalone-redis-stack/docker-compose.yml up -d
redis-tui -h localhost -p 6390
```

## Cluster

`redis:7-alpine` 6-node cluster (3 masters + 3 replicas) on ports `6380`-`6385`:

```sh
docker compose -f examples/cluster/docker-compose.yml up -d
redis-tui -c localhost:6380
```

## Cluster (Redis Stack)

`redis-stack-server` 6-node cluster on ports `6386`-`6392` (includes RedisJSON):

```sh
docker compose -f examples/cluster-redis-stack/docker-compose.yml up -d
redis-tui -c localhost:6386
```

## Makefile Shortcuts

```sh
make docker-up     # Start all four instances
make docker-down   # Stop all four instances
make docker-seed   # Seed all four instances
```

Individual targets are also available (e.g. `make docker-up-standalone-stack`). Run `make help` for the full list.

## Seed Data

Populate an instance with sample data covering every data type. Native RedisJSON keys are seeded automatically when the module is available (Redis Stack), otherwise they are skipped gracefully.

```sh
# Standalone (localhost:6379)
go run ./examples/seed

# Standalone Redis Stack (localhost:6390)
go run ./examples/seed -addr localhost:6390

# Cluster (localhost:6380)
go run ./examples/seed -addr localhost:6380 -cluster

# Cluster Redis Stack (localhost:6386)
go run ./examples/seed -addr localhost:6386 -cluster

# Flush existing data before seeding (add -flush)
go run ./examples/seed -flush
```

## Cleanup

```sh
make docker-down
```
