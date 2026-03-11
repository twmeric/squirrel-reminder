# Makefile for m03-trajectory

.PHONY: build build-optimized test benchmark docker run clean

# 变量
BINARY_NAME=m03-trajectory
DOCKER_IMAGE=squirrel/m03-trajectory:v1.2

# 构建
default: build

build:
	go build -o bin/$(BINARY_NAME) ./cmd/server

build-optimized:
	CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o bin/$(BINARY_NAME) ./cmd/server

# 测试
test:
	go test -v ./...

benchmark:
	go test -bench=. -benchmem ./internal/algorithm/

stress-test:
	go test -v ./tests/stress_test.go

# Docker
docker-build:
	docker build -t $(DOCKER_IMAGE) -f Dockerfile.optimized .

docker-run:
	docker-compose -f docker-compose.prod.yml up -d

docker-stop:
	docker-compose -f docker-compose.prod.yml down

# 性能分析
profile-cpu:
	go test -cpuprofile=cpu.prof -bench=. ./internal/algorithm/
	go tool pprof cpu.prof

profile-mem:
	go test -memprofile=mem.prof -bench=. ./internal/algorithm/
	go tool pprof mem.prof

# 清理
clean:
	rm -rf bin/
	rm -f *.prof
	go clean

# 代码检查
lint:
	golangci-lint run

fmt:
	go fmt ./...

# 帮助
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-optimized- Build optimized binary"
	@echo "  test           - Run unit tests"
	@echo "  benchmark      - Run benchmark tests"
	@echo "  stress-test    - Run stress tests"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run production environment"
	@echo "  clean          - Clean build artifacts"
