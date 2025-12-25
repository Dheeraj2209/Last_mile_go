# Repository Guidelines

## Project Structure & Module Organization
Implementation layout:
- API definitions live in `api/` (Buf workspace rooted here).
- Generated code and OpenAPI output go to `gen/`.
- Service binaries live in `cmd/<service>/`.
- Service code stays isolated (for example, `services/matching/`, `services/driver/`).
- Shared libraries stay in `internal/`.
- Deployment assets live in `deploy/` (Kubernetes manifests under `deploy/k8s/`).
- Keep frontend (React + Vite + Tailwind) in `web/` if added.

Go module path: `github.com/Dheeraj2209/Last_mile_go`. Protos reference generated Go packages under `gen/go/lastmile/v1`.

## Stack & Setup Commands
Required stack (from `project-statement/early-idea.txt`): Go 1.22, grpc-go, buf, grpc-gateway + `protoc-gen-openapiv2`, OpenTelemetry, MongoDB, Redis, Vault + External Secrets Operator, and Kubernetes manifests with HPA. Loki + Grafana are expected for logging/observability.

Suggested local tooling installs:
- `go version` should report `go1.22.x`.
- Install the Protocol Buffers compiler (`protoc`) via your OS package manager.
- `go install github.com/bufbuild/buf/cmd/buf@latest`
- `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
- `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`
- `go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest`
- `go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest`
Ensure Go bin on PATH (for tooling + buf plugins):
- `export PATH="$PATH:$(go env GOPATH)/bin"`

## Build, Test, and Development Commands
Current automation:
- `make buf-update` (update `api/buf.lock` after dependency changes)
- `make lint` (Buf lint for protos)
- `make proto` (generate Go + gRPC + gateway + OpenAPI)

Run a service locally (gRPC + gateway in one binary):
- `go run ./cmd/user --grpc-listen :9090 --grpc-endpoint localhost:9090 --http-addr :8080`

Runtime configuration (env or flags):
- `GRPC_LISTEN_ADDR` / `--grpc-listen` (default `:9090`)
- `GRPC_ENDPOINT` / `--grpc-endpoint` (default `localhost:9090`)
- `HTTP_ADDR` / `--http-addr` (default `:8080`)
- `OTEL_EXPORTER_OTLP_ENDPOINT` / `--otel-endpoint` (default empty, enables OTLP)
- `OTEL_EXPORTER_OTLP_INSECURE` / `--otel-insecure` (default `true`)
- `.env.example` contains local defaults (copy to `.env` if needed).
Note: `.env` usage is optional for now; can be added later if needed.

Storage configuration (optional until services wire them in):
- `MONGO_URI`, `MONGO_TIMEOUT` (default `10s`)
- `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` (default `0`), `REDIS_TIMEOUT` (default `5s`)
Storage clients:
- `NewMongoClient` / `NewRedisClient` return errors when required fields are missing. Services should call them only when they actually need the dependency.

Health checks:
- gRPC health service enabled (standard `grpc.health.v1.Health`).
- HTTP: `GET /healthz` and `GET /readyz` return `200 OK`.

No build/test/run scripts exist yet for services. Once code lands, keep a minimal command surface (for example, `make build`, `make test`, `make run`) and document the exact commands here.

## Next Build Steps (choose one)
- Kubernetes manifests for one service (Deployment + Service + HPA + probes).
- Service config expansion (env validation + `.env.example` for local dev).
- Basic logging + request tracing interceptors (gRPC + HTTP).
- Storage wiring stubs (MongoDB + Redis clients in `internal/`).
- Minimal service logic (start with User + Station CRUD).

Observability:
- gRPC uses OpenTelemetry stats handler + logging interceptors.
- HTTP requests are wrapped with tracing + access logs.

Minimal service logic (in-memory):
- `UserService` supports Create/Get for rider and driver profiles (IDs auto-generated if missing).
- `StationService` supports Upsert/Get/List with simple offset pagination (`page_token` as offset string).

## Coding Style & Naming Conventions
Use standard formatters (`gofmt`, `prettier`) and keep service names aligned to the spec (`user`, `driver`, `rider`, `matching`, `trip`, `notification`, `location`, `station`). Prefer lowercase directory names such as `services/rider/`.

## Testing Guidelines
No tests exist yet. As code is added, keep unit tests next to the code they cover and follow framework defaults (`*_test.go`, `*.spec.ts`). Document how to run them alongside build commands.

## Commit & Pull Request Guidelines
History has a single commit (“first commit”), so no convention exists. Use short, imperative messages (for example, “Add matching service scaffold”). PRs should explain what/why, link specs, and include test/run notes plus any Kubernetes steps.

## Architecture Notes
Keep service boundaries aligned to the spec and communicate via explicit APIs rather than shared databases.

## Agent-Specific Instructions
If you add scripts or automation, update this file with new commands and paths for quick validation.
