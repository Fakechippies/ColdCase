# ColdCase — developer workflow
#
#   make build             compile the CLI to bin/coldcase
#   make test              race-enabled tests with coverage
#   make lint              golangci-lint (config in .golangci.yml)
#   make vet               go vet
#   make vuln              govulncheck
#   make docker            build the all-in-one forensics image
#   make release-snapshot  local GoReleaser snapshot (no upload)
#
# External tools (golangci-lint, govulncheck, goreleaser) are invoked via
# `go run`/`go tool`-style installation so the repo has no vendored
# dependencies for CI tooling.

BIN      := bin/coldcase
SIGNED ?= checksums.txt

.PHONY: all build install test vet lint vuln docker release-snapshot sign verify clean

all: build

build:
	go build -o $(BIN) ./cmd/coldcase

install:
	go install ./cmd/coldcase

test:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | tail -1

vet:
	go vet ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --timeout 5m

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

docker:
	docker build -t coldcase:latest .

release-snapshot:
	go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean

sign:
	go run ./cmd/coldcase-sign sign $(SIGNED)

verify:
	go run ./cmd/coldcase-sign verify -pub $$(cat keys/release_public.key) $(SIGNED) $(SIGNED).sig

clean:
	rm -rf bin dist coverage.out
