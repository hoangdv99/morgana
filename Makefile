VERSION := $(shell cat VERSION)
COMMIT_HASH := $(shell git rev-parse --short HEAD)
PROJECT_NAME := morgana

all: generate build-all

generate:
	protoc -I=. \
		--go_out=internal/generated/ \
		--go-grpc_out=internal/generated/ \
		--grpc-gateway_out=internal/generated \
		--grpc-gateway_opt generate_unbound_methods=true \
		--openapiv2_out . \
		--openapiv2_opt generate_unbound_methods=true \
		--validate_out="lang=go:internal/generated" \
		api/morgana.proto
	wire internal/wiring/wire.go
	
.PHONY: build-linux-amd64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_linux_amd64 cmd/$(PROJECT_NAME)/*.go

.PHONY: build-linux-arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_linux_arm64 cmd/$(PROJECT_NAME)/*.go

.PHONY: build-macos-amd64
build-macos-amd64:
	GOOS=darwin GOARCH=amd64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_macos_amd64 cmd/$(PROJECT_NAME)/*.go

.PHONY: build-macos-arm64
build-macos-arm64:
	GOOS=darwin GOARCH=arm64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_macos_arm64 cmd/$(PROJECT_NAME)/*.go

.PHONY: build-windows-amd64
build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_windows_amd64.exe cmd/$(PROJECT_NAME)/*.go

.PHONY: build-windows-arm64
build-windows-arm64:
	GOOS=windows GOARCH=amd64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_windows_arm64.exe cmd/$(PROJECT_NAME)/*.go

.PHONY: build-all
build-all:
	make build-linux-amd64
	make build-linux-arm64
	make build-macos-amd64
	make build-macos-arm64
	make build-windows-amd64
	make build-windows-arm64

.PHONY: build
build:
	go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME) \
		cmd/$(PROJECT_NAME)/*.go

.PHONY: clean
clean:
	rm -rf build/

.PHONY: run-server
run-server:
	go run cmd/$(PROJECT_NAME)/*.go server

.PHONY: lint
lint:
	golangci-lint run ./... 

.PHONY: docker-compose-dev-up
docker-compose-dev-up:
	docker-compose -f deployments/docker-compose.dev.yaml up -d

.PHONY: docker-compose-dev-down
docker-compose-dev-down:
	docker-compose -f deployments/docker-compose.dev.yaml down
