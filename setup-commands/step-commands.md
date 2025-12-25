# Setup Commands

This document records install commands for the required stack (Fedora only).

## Go 1.22
Fedora:
```bash
sudo dnf install -y golang
```
Verify:
```bash
go version
```

## Protocol Buffers Compiler (protoc)
Fedora:
```bash
sudo dnf install -y protobuf-compiler
```

## Go Tooling (gRPC, Buf, Gateway, OpenAPI)
```bash
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```
Ensure Go bin on PATH:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```
Verify:
```bash
buf --version
protoc --version
protoc-gen-go --version
```

## MongoDB (local, systemd)
Fedora:
```bash
sudo tee /etc/yum.repos.d/mongodb-org-7.0.repo >/dev/null <<'EOF2'
[mongodb-org-7.0]
name=MongoDB Repository
baseurl=https://repo.mongodb.org/yum/redhat/9/mongodb-org/7.0/x86_64/
gpgcheck=1
enabled=1
gpgkey=https://pgp.mongodb.com/server-7.0.asc
EOF2
```
```bash
sudo dnf install -y mongodb-org
```
```bash
sudo systemctl start mongod
sudo systemctl enable mongod
```
Verify:
```bash
mongosh
```
Default connection string:
```bash
mongodb://localhost:27017
```

## Redis (Fedora uses Valkey for Redis compatibility)
Fedora:
```bash
sudo dnf install -y valkey valkey-compat-redis
```
Verify:
```bash
redis-server --version || valkey-server --version
```

## Vault (HashiCorp)
Fedora:
```bash
sudo dnf install -y dnf-plugins-core
sudo dnf config-manager --add-repo https://rpm.releases.hashicorp.com/fedora/hashicorp.repo
sudo dnf install -y vault
```
Verify:
```bash
vault --version
```

## Helm
Fedora:
```bash
sudo dnf install -y helm
```
Verify:
```bash
helm version
```

## Minikube (local Kubernetes)
Fedora:
```bash
sudo dnf install -y minikube
```
Verify:
```bash
minikube version
```
