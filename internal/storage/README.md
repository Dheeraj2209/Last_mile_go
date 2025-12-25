# Storage stubs

This package provides:
- MongoDB + Redis client helpers.
- Storage interfaces with in-memory implementations for user/station services (used by default).

Mongo:
- env: `MONGO_URI`, optional `MONGO_TIMEOUT` (default 10s)

Redis:
- env: `REDIS_ADDR`, optional `REDIS_PASSWORD`, `REDIS_DB`, `REDIS_TIMEOUT` (default 5s)

These are optional until a service wires them in.

In-memory stores:
- `NewMemoryUserStore()` implements Rider/Driver stores.
- `NewMemoryStationStore()` implements Station store.
