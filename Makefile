GOBIN ?= $(shell go env GOPATH)/bin
BUF ?= $(GOBIN)/buf

.PHONY: proto lint buf-update

buf-update:
	@PATH="$(GOBIN):$$PATH" $(BUF) dep update api

lint:
	@PATH="$(GOBIN):$$PATH" $(BUF) lint

proto:
	@PATH="$(GOBIN):$$PATH" $(BUF) generate
