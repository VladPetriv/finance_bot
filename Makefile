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

.PHONY: docker-up
docker-up:
	docker compose up --build -d

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-logs
docker-logs:
	docker compose logs -f bot
