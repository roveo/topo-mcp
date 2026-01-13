.PHONY: lint format test install build clean

# Default build (Go only)
build:
	go build -o bin/topo .

# Build profiles - language combinations for different use cases
build-go:
	go build -tags lang_go -o bin/topo-go .

build-python:
	go build -tags lang_python -o bin/topo-python .

build-typescript:
	go build -tags lang_typescript -o bin/topo-typescript .

build-rust:
	go build -tags lang_rust -o bin/topo-rust .

build-backend:
	go build -tags "lang_go,lang_python,lang_rust" -o bin/topo-backend .

build-frontend:
	go build -tags lang_typescript -o bin/topo-frontend .

build-fullstack:
	go build -tags "lang_go,lang_typescript" -o bin/topo-fullstack .

build-web:
	go build -tags "lang_python,lang_typescript" -o bin/topo-web .

build-ml:
	go build -tags "lang_python,lang_rust" -o bin/topo-ml .

build-all:
	go build -tags lang_all -o bin/topo-all .

# Build all profiles
build-profiles: build-go build-python build-typescript build-rust build-backend build-frontend build-fullstack build-web build-ml build-all
	@echo "Built all profiles in bin/"
	@ls -lh bin/

clean:
	rm -rf bin/

lint:
	golangci-lint run --fix

format:
	gofumpt -w .

test:
	go test -mod=readonly ./... -count=1

test-all:
	go test -tags lang_all -mod=readonly ./... -count=1

install:
	go install .

install-all:
	go install -tags lang_all .
