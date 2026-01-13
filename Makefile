.PHONY: lint format test install build clean

# Default build (all languages)
build:
	go build -o bin/topo .

# Build with specific languages only
build-go:
	go build -tags lang_go -o bin/topo-go .

build-python:
	go build -tags lang_python -o bin/topo-python .

build-typescript:
	go build -tags lang_typescript -o bin/topo-typescript .

build-rust:
	go build -tags lang_rust -o bin/topo-rust .

# Build profiles - language combinations for different use cases
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

# Build all profiles
build-profiles: build build-go build-python build-typescript build-rust build-backend build-frontend build-fullstack build-web build-ml
	@echo "Built all profiles in bin/"
	@ls -lh bin/

clean:
	rm -rf bin/

lint:
	golangci-lint run --fix

format:
	gofumpt -w .

# Default test (all languages)
test:
	go test -mod=readonly ./... -count=1

# Test with specific language
test-go:
	go test -tags lang_go -mod=readonly ./... -count=1

install:
	go install .
