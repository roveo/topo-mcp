lint:
	golangci-lint run --fix

format:
	gofumpt -w .

test:
	go test -mod=readonly ./... -count=1


install:
	go install .
