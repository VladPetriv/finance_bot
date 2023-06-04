.PHONY: run
run:
	go run ./cmd/main.go

.PHONY: build
build:
	go build -o finance_bot ./cmd/main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: mock
mock:
	go generate ./...
