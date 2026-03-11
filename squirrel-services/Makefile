# Squirrel Services Makefile
# 统一构建脚本

.PHONY: all build test lint clean docker-up docker-down proto

# 变量
M03_DIR := m03-trajectory
M04_DIR := m04-insight-engine
DOCKER_DIR := docker

# 默认目标
all: build

# 构建所有服务
build: build-m03 build-m04

# 构建 m03 (Go)
build-m03:
	@echo "Building m03-trajectory..."
	cd $(M03_DIR) && go build -ldflags="-w -s" -o bin/m03-trajectory ./cmd/server

# 构建 m04 (Python)
build-m04:
	@echo "Building m04-insight-engine..."
	cd $(M04_DIR) && pip install -r requirements.txt

# 运行测试
test: test-m03 test-m04

test-m03:
	@echo "Testing m03-trajectory..."
	cd $(M03_DIR) && go test -v -race -coverprofile=coverage.out ./...
	cd $(M03_DIR) && go tool cover -html=coverage.out -o coverage.html

test-m04:
	@echo "Testing m04-insight-engine..."
	cd $(M04_DIR) && python -m pytest tests/ -v --cov=src --cov-report=html

# 代码检查
lint: lint-m03 lint-m04

lint-m03:
	cd $(M03_DIR) && golangci-lint run ./...

lint-m04:
	cd $(M04_DIR) && flake8 src/ && black --check src/

# 生成 protobuf 代码
proto:
	@echo "Generating protobuf code..."
	cd $(M03_DIR) && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/trajectory.proto
	cd $(M04_DIR) && python -m grpc_tools.protoc \
		-I../$(M03_DIR)/proto \
		--python_out=./src/proto \
		--grpc_python_out=./src/proto \
		trajectory.proto

# Docker 操作
docker-up:
	@echo "Starting all services..."
	cd $(DOCKER_DIR) && docker-compose up -d --build

docker-down:
	@echo "Stopping all services..."
	cd $(DOCKER_DIR) && docker-compose down

docker-logs:
	cd $(DOCKER_DIR) && docker-compose logs -f

docker-clean:
	cd $(DOCKER_DIR) && docker-compose down -v --rmi all

# 数据库迁移
migrate-up:
	@echo "Running database migrations..."
	cd $(DOCKER_DIR) && docker-compose exec tidb mysql -h localhost -P 4000 -e "source /migrations/001_init.sql"

# 性能测试
benchmark-m03:
	cd $(M03_DIR) && go test -bench=. -benchmem ./internal/algorithm/

# 安全扫描
security-scan:
	@echo "Running security scan..."
	cd $(M03_DIR) && gosec ./...
	cd $(M04_DIR) && bandit -r src/

# 清理
clean:
	rm -rf $(M03_DIR)/bin
	rm -rf $(M03_DIR)/coverage.out $(M03_DIR)/coverage.html
	rm -rf $(M04_DIR)/htmlcov
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete

# CI/CD 目标
ci-test: lint test
ci-build: clean build
ci-deploy: docker-up

# 帮助
help:
	@echo "Available targets:"
	@echo "  build        - Build all services"
	@echo "  test         - Run all tests"
	@echo "  lint         - Run code linting"
	@echo "  proto        - Generate protobuf code"
	@echo "  docker-up    - Start services with Docker Compose"
	@echo "  docker-down  - Stop services"
	@echo "  clean        - Clean build artifacts"
