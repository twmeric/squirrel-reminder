#!/bin/bash
# ============================================
# Squirrel Services Performance Benchmark
# ============================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

RESULTS_DIR="$PROJECT_ROOT/benchmark_results"
mkdir -p "$RESULTS_DIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULT_FILE="$RESULTS_DIR/benchmark_$TIMESTAMP.txt"

echo "============================================" | tee -a "$RESULT_FILE"
echo "Squirrel Services Benchmark" | tee -a "$RESULT_FILE"
echo "Timestamp: $(date)" | tee -a "$RESULT_FILE"
echo "============================================" | tee -a "$RESULT_FILE"
echo "" | tee -a "$RESULT_FILE"

# 检查服务是否运行
check_services() {
    echo "Checking services..." | tee -a "$RESULT_FILE"
    
    if curl -s http://localhost:8083/health > /dev/null 2>&1; then
        echo "✓ M03 is running" | tee -a "$RESULT_FILE"
    else
        echo "✗ M03 is not running" | tee -a "$RESULT_FILE"
        exit 1
    fi
    
    if curl -s http://localhost:8084/health > /dev/null 2>&1; then
        echo "✓ M04 is running" | tee -a "$RESULT_FILE"
    else
        echo "✗ M04 is not running" | tee -a "$RESULT_FILE"
        exit 1
    fi
    
    echo "" | tee -a "$RESULT_FILE"
}

# M03基准测试
benchmark_m03() {
    echo "============================================" | tee -a "$RESULT_FILE"
    echo "M03 Trajectory Service Benchmark" | tee -a "$RESULT_FILE"
    echo "============================================" | tee -a "$RESULT_FILE"
    
    # GetSpeed 基准测试
    echo "" | tee -a "$RESULT_FILE"
    echo "Testing GetSpeed endpoint..." | tee -a "$RESULT_FILE"
    
    START_TIME=$(date +%s%N)
    for i in {1..100}; do
        curl -s http://localhost:8083/api/v1/speed?user_id=user_$i > /dev/null
    done
    END_TIME=$(date +%s%N)
    
    DURATION=$(( (END_TIME - START_TIME) / 1000000 ))  # ms
    AVG=$(( DURATION / 100 ))
    
    echo "  Total time: ${DURATION}ms" | tee -a "$RESULT_FILE"
    echo "  Average latency: ${AVG}ms" | tee -a "$RESULT_FILE"
    echo "  Requests: 100" | tee -a "$RESULT_FILE"
    echo "  Target: <10ms" | tee -a "$RESULT_FILE"
    
    if [ $AVG -lt 10 ]; then
        echo "  ✓ PASSED" | tee -a "$RESULT_FILE"
    else
        echo "  ✗ FAILED" | tee -a "$RESULT_FILE"
    fi
    
    # ProcessBatch 基准测试
    echo "" | tee -a "$RESULT_FILE"
    echo "Testing ProcessBatch endpoint..." | tee -a "$RESULT_FILE"
    
    # 生成测试数据
    PAYLOAD='{"user_id":"bench_user","locations":['
    for i in {0..99}; do
        LAT=$(echo "22.5431 + $i * 0.0001" | bc -l)
        LNG=$(echo "113.9589 + $i * 0.0001" | bc -l)
        PAYLOAD="${PAYLOAD}{\"lat\":$LAT,\"lng\":$LNG,\"timestamp\":$(date +%s),\"speed\":30}"
        if [ $i -lt 99 ]; then
            PAYLOAD="${PAYLOAD},"
        fi
    done
    PAYLOAD="${PAYLOAD}]}"
    
    START_TIME=$(date +%s%N)
    curl -s -X POST http://localhost:8083/api/v1/batch \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" > /dev/null
    END_TIME=$(date +%s%N)
    
    DURATION=$(( (END_TIME - START_TIME) / 1000000 ))
    
    echo "  Batch size: 100 points" | tee -a "$RESULT_FILE"
    echo "  Total time: ${DURATION}ms" | tee -a "$RESULT_FILE"
    echo "  Target: <50ms" | tee -a "$RESULT_FILE"
    
    if [ $DURATION -lt 50 ]; then
        echo "  ✓ PASSED" | tee -a "$RESULT_FILE"
    else
        echo "  ✗ FAILED" | tee -a "$RESULT_FILE"
    fi
    
    echo "" | tee -a "$RESULT_FILE"
}

# M04基准测试
benchmark_m04() {
    echo "============================================" | tee -a "$RESULT_FILE"
    echo "M04 Insight Engine Benchmark" | tee -a "$RESULT_FILE"
    echo "============================================" | tee -a "$RESULT_FILE"
    
    # 单用户分析
    echo "" | tee -a "$RESULT_FILE"
    echo "Testing single user analysis..." | tee -a "$RESULT_FILE"
    
    START_TIME=$(date +%s%N)
    curl -s http://localhost:8084/api/v1/users/bench_user/profile?days=30 > /dev/null
    END_TIME=$(date +%s%N)
    
    DURATION=$(( (END_TIME - START_TIME) / 1000000 ))
    
    echo "  Analysis period: 30 days" | tee -a "$RESULT_FILE"
    echo "  Total time: ${DURATION}ms" | tee -a "$RESULT_FILE"
    echo "  Target: <200ms" | tee -a "$RESULT_FILE"
    
    if [ $DURATION -lt 200 ]; then
        echo "  ✓ PASSED" | tee -a "$RESULT_FILE"
    else
        echo "  ✗ FAILED" | tee -a "$RESULT_FILE"
    fi
    
    # 批量分析
    echo "" | tee -a "$RESULT_FILE"
    echo "Testing batch analysis (10 users)..." | tee -a "$RESULT_FILE"
    
    PAYLOAD='{"user_ids":['
    for i in {1..10}; do
        PAYLOAD="${PAYLOAD}\"bench_user_$i\""
        if [ $i -lt 10 ]; then
            PAYLOAD="${PAYLOAD},"
        fi
    done
    PAYLOAD="${PAYLOAD}],\"days\":30}"
    
    START_TIME=$(date +%s%N)
    curl -s -X POST "http://localhost:8084/api/v1/batch/analyze?days=30" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" > /dev/null
    END_TIME=$(date +%s%N)
    
    DURATION=$(( (END_TIME - START_TIME) / 1000000 ))
    
    echo "  Batch size: 10 users" | tee -a "$RESULT_FILE"
    echo "  Total time: ${DURATION}ms" | tee -a "$RESULT_FILE"
    echo "  Target: <500ms" | tee -a "$RESULT_FILE"
    
    if [ $DURATION -lt 500 ]; then
        echo "  ✓ PASSED" | tee -a "$RESULT_FILE"
    else
        echo "  ✗ FAILED" | tee -a "$RESULT_FILE"
    fi
    
    echo "" | tee -a "$RESULT_FILE"
}

# 资源使用测试
benchmark_resources() {
    echo "============================================" | tee -a "$RESULT_FILE"
    echo "Resource Usage" | tee -a "$RESULT_FILE"
    echo "============================================" | tee -a "$RESULT_FILE"
    
    echo "" | tee -a "$RESULT_FILE"
    echo "M03 Container Stats:" | tee -a "$RESULT_FILE"
    docker stats --no-stream squirrel-m03 2>/dev/null | tee -a "$RESULT_FILE" || echo "Container not found"
    
    echo "" | tee -a "$RESULT_FILE"
    echo "M04 Container Stats:" | tee -a "$RESULT_FILE"
    docker stats --no-stream squirrel-m04 2>/dev/null | tee -a "$RESULT_FILE" || echo "Container not found"
    
    echo "" | tee -a "$RESULT_FILE"
}

# 负载测试
load_test() {
    echo "============================================" | tee -a "$RESULT_FILE"
    echo "Load Test (if wrk is available)" | tee -a "$RESULT_FILE"
    echo "============================================" | tee -a "$RESULT_FILE"
    
    if command -v wrk &> /dev/null; then
        echo "" | tee -a "$RESULT_FILE"
        echo "Running wrk load test on M03..." | tee -a "$RESULT_FILE"
        wrk -t4 -c100 -d30s http://localhost:8083/health 2>&1 | tee -a "$RESULT_FILE"
        
        echo "" | tee -a "$RESULT_FILE"
        echo "Running wrk load test on M04..." | tee -a "$RESULT_FILE"
        wrk -t4 -c100 -d30s http://localhost:8084/health 2>&1 | tee -a "$RESULT_FILE"
    else
        echo "wrk not installed, skipping load test" | tee -a "$RESULT_FILE"
        echo "Install: brew install wrk (macOS) or apt-get install wrk (Ubuntu)" | tee -a "$RESULT_FILE"
    fi
    
    echo "" | tee -a "$RESULT_FILE"
}

# 主函数
main() {
    check_services
    benchmark_m03
    benchmark_m04
    benchmark_resources
    load_test
    
    echo "============================================" | tee -a "$RESULT_FILE"
    echo "Benchmark Complete!" | tee -a "$RESULT_FILE"
    echo "Results saved to: $RESULT_FILE" | tee -a "$RESULT_FILE"
    echo "============================================" | tee -a "$RESULT_FILE"
}

main "$@"
