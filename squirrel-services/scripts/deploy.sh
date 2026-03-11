#!/bin/bash
# ============================================
# Squirrel Services Deployment Script
# ============================================

set -e

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENVIRONMENT="${1:-staging}"
VERSION="${2:-latest}"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "Checking dependencies..."
    
    command -v docker >/dev/null 2>&1 || { log_error "Docker is required but not installed."; exit 1; }
    command -v docker-compose >/dev/null 2>&1 || { log_error "Docker Compose is required but not installed."; exit 1; }
    
    log_info "All dependencies satisfied"
}

# 构建镜像
build_images() {
    log_info "Building Docker images..."
    
    cd "$PROJECT_ROOT"
    
    # Build M03
    log_info "Building M03 Trajectory Service..."
    docker build \
        -f docker/Dockerfile.m03 \
        -t "squirrel/m03:${VERSION}" \
        -t "squirrel/m03:latest" \
        .
    
    # Build M04
    log_info "Building M04 Insight Engine..."
    docker build \
        -f docker/Dockerfile.m04 \
        -t "squirrel/m04:${VERSION}" \
        -t "squirrel/m04:latest" \
        .
    
    log_info "Images built successfully"
}

# 运行测试
run_tests() {
    log_info "Running tests..."
    
    cd "$PROJECT_ROOT/m03-trajectory"
    go test ./... || { log_error "M03 tests failed"; exit 1; }
    
    cd "$PROJECT_ROOT/m04-insight-engine"
    python -m pytest tests/ || { log_error "M04 tests failed"; exit 1; }
    
    log_info "All tests passed"
}

# 部署到开发环境
deploy_local() {
    log_info "Deploying to local environment..."
    
    cd "$PROJECT_ROOT/docker"
    
    # 加载环境变量
    if [ -f .env ]; then
        export $(cat .env | xargs)
    fi
    
    # 启动服务
    docker-compose down
    docker-compose up -d --build
    
    # 等待服务就绪
    log_info "Waiting for services to be ready..."
    sleep 10
    
    # 健康检查
    check_health
    
    log_info "Local deployment complete!"
    log_info "M03: http://localhost:8083"
    log_info "M04: http://localhost:8084"
    log_info "Grafana: http://localhost:3000"
}

# 健康检查
check_health() {
    log_info "Checking service health..."
    
    # Check M03
    if curl -s http://localhost:8083/health > /dev/null; then
        log_info "M03 is healthy"
    else
        log_error "M03 health check failed"
        return 1
    fi
    
    # Check M04
    if curl -s http://localhost:8084/health > /dev/null; then
        log_info "M04 is healthy"
    else
        log_error "M04 health check failed"
        return 1
    fi
}

# 部署到腾讯云
deploy_tencent() {
    log_info "Deploying to Tencent Cloud TKE..."
    
    # 检查环境变量
    if [ -z "$TENCENT_SECRET_ID" ] || [ -z "$TENCENT_SECRET_KEY" ]; then
        log_error "Tencent Cloud credentials not set"
        exit 1
    fi
    
    # 配置kubectl
    # tkectl configure ...
    
    # 应用K8s配置
    kubectl apply -f k8s/namespace.yaml
    kubectl apply -f k8s/configmap.yaml
    kubectl apply -f k8s/secret.yaml
    kubectl apply -f k8s/m03-deployment.yaml
    kubectl apply -f k8s/m04-deployment.yaml
    kubectl apply -f k8s/service.yaml
    kubectl apply -f k8s/ingress.yaml
    
    # 等待部署完成
    kubectl rollout status deployment/m03-trajectory
    kubectl rollout status deployment/m04-insight
    
    log_info "Tencent Cloud deployment complete!"
}

# 数据库迁移
migrate_database() {
    log_info "Running database migrations..."
    
    # 使用Tidb的MySQL协议执行迁移
    mysql -h "$TIDB_HOST" -P 4000 -u root < "$PROJECT_ROOT/docker/migrations/001_init.sql"
    
    log_info "Database migrations complete"
}

# 回滚
deploy_rollback() {
    log_warn "Rolling back deployment..."
    
    cd "$PROJECT_ROOT/docker"
    docker-compose down
    
    # 恢复上一个版本
    PREVIOUS_VERSION=$(docker images squirrel/m03 --format "{{.Tag}}" | grep -v latest | head -1)
    
    if [ -n "$PREVIOUS_VERSION" ]; then
        log_info "Rolling back to version $PREVIOUS_VERSION"
        docker-compose up -d
    else
        log_error "No previous version found for rollback"
    fi
}

# 清理资源
cleanup() {
    log_info "Cleaning up resources..."
    
    # 清理旧镜像
    docker images --filter "dangling=true" -q | xargs -r docker rmi
    
    # 清理未使用的卷
    docker volume prune -f
    
    log_info "Cleanup complete"
}

# 主函数
main() {
    case "${1:-local}" in
        local)
            check_dependencies
            build_images
            run_tests
            deploy_local
            ;;
        tencent|tke)
            check_dependencies
            build_images
            run_tests
            migrate_database
            deploy_tencent
            ;;
        build)
            check_dependencies
            build_images
            ;;
        test)
            run_tests
            ;;
        rollback)
            deploy_rollback
            ;;
        cleanup)
            cleanup
            ;;
        *)
            echo "Usage: $0 {local|tencent|build|test|rollback|cleanup}"
            exit 1
            ;;
    esac
}

main "$@"
