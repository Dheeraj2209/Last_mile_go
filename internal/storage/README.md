# Storage stubs

This package provides:
- MongoDB + Redis client helpers.
- Storage interfaces with in-memory implementations for user/station services (used by default).
- Mongo/Redis-backed stores for user/station services (enabled via env).

Mongo:
- env: `MONGO_URI`, optional `MONGO_TIMEOUT` (default 10s)
- store config: `MONGO_DB`, `MONGO_RIDER_COLLECTION`, `MONGO_DRIVER_COLLECTION`, `MONGO_STATION_COLLECTION`

Redis:
- env: `REDIS_ADDR`, optional `REDIS_PASSWORD`, `REDIS_DB`, `REDIS_TIMEOUT` (default 5s)
- store config: `REDIS_KEY_PREFIX`

These are optional until a service wires them in.

In-memory stores:
- `NewMemoryUserStore()` implements Rider/Driver stores.
- `NewMemoryStationStore()` implements Station store.

Mongo stores:
- `NewMongoUserStore()` implements Rider/Driver stores.
- `NewMongoStationStore()` implements Station store.

Redis stores:
- `NewRedisUserStore()` implements Rider/Driver stores.
- `NewRedisStationStore()` implements Station store (sorted set index).
