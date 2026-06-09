# ── Variables ─────────────────────────────────────────────────────────────────
BINARY_NAME=server
BUILD_DIR=./bin
CMD_PATH=./cmd/api
MODULE=github.com/Afraaaaaim/go-server-boilerplate

# ── Go commands ───────────────────────────────────────────────────────────────
.PHONY: run
run:
	go run $(CMD_PATH)

.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

.PHONY: test
test:
	go test ./... -v -race -count=1

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

# ── Docker commands ───────────────────────────────────────────────────────────
.PHONY: docker-build
docker-build:
	docker build -f docker/Dockerfile -t $(BINARY_NAME) .

.PHONY: docker-up
docker-up:
	docker compose -f docker/docker-compose.yml up --build -d

.PHONY: docker-down
docker-down:
	docker compose -f docker/docker-compose.yml down

.PHONY: docker-logs
docker-logs:
	docker compose -f docker/docker-compose.yml logs -f app

# ── Dev helpers ───────────────────────────────────────────────────────────────
.PHONY: env
env:
	cp .env.example .env
	@echo ".env created from .env.example"

.PHONY: help
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "  run            Run the server locally"
	@echo "  build          Build binary to ./bin/"
	@echo "  test           Run all tests"
	@echo "  tidy           Run go mod tidy"
	@echo "  lint           Run golangci-lint"
	@echo "  clean          Remove build artifacts"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-up      Start full stack with docker compose"
	@echo "  docker-down    Stop docker compose stack"
	@echo "  docker-logs    Tail app logs from docker compose"
	@echo "  env            Create .env from .env.example"
	@echo ""