.PHONY: run build test tidy generate stop

BINARY := ./alfredo

run:
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi; go run ./cmd/server

build:
	go build -o $(BINARY) ./cmd/server

test:
	go test ./internal/...

tidy:
	go mod tidy

generate:
	go generate ./...

stop:
	@if [ -f alfredo.pid ]; then \
		kill $$(cat alfredo.pid) 2>/dev/null || true; \
		rm alfredo.pid; \
		echo "alfredo stopped."; \
	else \
		echo "No alfredo.pid found."; \
	fi
