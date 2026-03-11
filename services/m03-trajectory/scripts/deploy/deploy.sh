#!/bin/bash
# deploy.sh - m03-trajectory 部署脚本

set -e

SERVICE_NAME="m03-trajectory"
HTTP_PORT=8083
GRPC_PORT=50053
METRICS_PORT=9090

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 启动服务
start() {
    log_info "Starting $SERVICE_NAME..."
    
    # 检查依赖
    check_dependencies
    
    # 启动基础设施
    log_info "Starting infrastructure (TiDB, Redis)..."
    docker-compose -f docker-compose.prod.yml up -d tidb redis
    
    # 等待依赖就绪
    wait_for_dependencies
    
    # 启动应用
    log_info "Starting $SERVICE_NAME application..."
    docker-compose -f docker-compose.prod.yml up -d m03
    
    # 等待服务就绪
    sleep 5
    
    # 健康检查
    if health_check; then
        log_success "$SERVICE_NAME started successfully!"
        show_status
    else
        log_error "Failed to start $SERVICE_NAME"
        exit 1
    fi
}

# 检查依赖
check_dependencies() {
    log_info "Checking dependencies..."
    
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker not found"
        exit 1
    fi
    
    # 检查 docker-compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "docker-compose not found"
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

# 等待依赖就绪
wait_for_dependencies() {
    log_info "Waiting for dependencies..."
    
    # 等待 TiDB
    for i in {1..30}; do
        if curl -s http://localhost:10080/status &> /dev/null; then
            log_success "TiDB is ready"
            break
        fi
        echo -n "."
        sleep 2
    done
    
    # 等待 Redis
    for i in {1..30}; do
        if docker exec m03-redis-prod redis-cli ping &> /dev/null; then
            log_success "Redis is ready"
            break
        fi
        echo -n "."
        sleep 2
    done
    
    echo ""
}

# 健康检查
health_check() {
    log_info "Performing health check..."
    
    local max_retries=10
    local retry_count=0
    
    while [ $retry_count -lt $max_retries ]; do
        # HTTP 健康检查
        if curl -s http://localhost:$HTTP_PORT/health | grep -q "ok"; then
            log_success "HTTP health check passed"
            
            # 详细健康检查
            local health_detail=$(curl -s http://localhost:$HTTP_PORT/ready)
            log_info "Health detail: $health_detail"
            
            return 0
        fi
        
        retry_count=$((retry_count + 1))
        log_warn "Health check attempt $retry_count/$max_retries failed, retrying..."
        sleep 3
    done
    
    return 1
}

# 停止服务
stop() {
    log_info "Stopping $SERVICE_NAME..."
    docker-compose -f docker-compose.prod.yml down
    log_success "$SERVICE_NAME stopped"
}

# 重启服务
restart() {
    stop
    start
}

# 显示状态
status() {
    show_status
}

show_status() {
    echo ""
    echo "======================================"
    echo "  $SERVICE_NAME Status"
    echo "======================================"
    
    # 容器状态
    echo ""
    echo "Docker Containers:"
    docker-compose -f docker-compose.prod.yml ps
    
    # 服务地址
    echo ""
    echo "Service Endpoints:"
    echo "  HTTP API:     http://localhost:$HTTP_PORT"
    echo "  gRPC API:     localhost:$GRPC_PORT"
    echo "  Metrics:      http://localhost:$METRICS_PORT/metrics"
    echo "  Health:       http://localhost:$HTTP_PORT/health"
    
    # 健康状态
    echo ""
    echo "Health Status:"
    curl -s http://localhost:$HTTP_PORT/health 2>/dev/null || echo "  Unable to connect"
    
    echo ""
    echo "======================================"
}

# 查看日志
logs() {
    docker-compose -f docker-compose.prod.yml logs -f m03
}

# 性能测试
benchmark() {
    log_info "Running performance benchmark..."
    
    echo ""
    echo "Testing API latency..."
    
    # 测试 GetSpeed
    echo "Testing GetSpeed..."
    for i in {1..100}; do
        curl -s -o /dev/null -w "%{time_total}\n" \
            http://localhost:$HTTP_PORT/api/v1/speed?user_id=test_user
    done | sort -n | tail -1 | xargs -I {} echo "  P99: {}s"
    
    log_success "Benchmark completed"
}

# 使用说明
usage() {
    echo "Usage: $0 {start|stop|restart|status|health|logs|benchmark}"
    echo ""
    echo "Commands:"
    echo "  start     - Start the service"
    echo "  stop      - Stop the service"
    echo "  restart   - Restart the service"
    echo "  status    - Show service status"
    echo "  health    - Perform health check"
    echo "  logs      - View service logs"
    echo "  benchmark - Run performance benchmark"
    echo ""
}

# 主逻辑
case "${1:-}" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    health)
        health_check
        ;;
    logs)
        logs
        ;;
    benchmark)
        benchmark
        ;;
    *)
        usage
        exit 1
        ;;
esac
