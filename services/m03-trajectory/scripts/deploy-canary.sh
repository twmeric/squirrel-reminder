#!/bin/bash
# deploy-canary.sh - v1.2.1 灰度发布脚本

set -e

VERSION="v1.2.1"
PHASE=${1:-"help"}
NAMESPACE="squirrel-prod"
DEPLOYMENT="m03-trajectory"

echo "🚀 v1.2.1 灰度发布工具"
echo "======================"
echo ""

show_help() {
    echo "用法: $0 <phase>"
    echo ""
    echo "阶段:"
    echo "  10      - 10% 灰度 (3个副本)"
    echo "  50      - 50% 灰度 (5个副本)"
    echo "  100     - 100% 全量 (10个副本)"
    echo "  status  - 查看当前状态"
    echo "  verify  - 验证发布结果"
    echo "  rollback- 回滚到v1.2.0"
    echo ""
}

check_prerequisites() {
    echo "🔍 检查前置条件..."
    
    # 检查kubectl
    if ! command -v kubectl &> /dev/null; then
        echo "❌ kubectl 未安装"
        exit 1
    fi
    
    # 检查集群连接
    if ! kubectl cluster-info &> /dev/null; then
        echo "❌ 无法连接集群"
        exit 1
    fi
    
    # 检查deployment存在
    if ! kubectl get deployment $DEPLOYMENT -n $NAMESPACE &> /dev/null; then
        echo "❌ Deployment $DEPLOYMENT 不存在"
        exit 1
    fi
    
    echo "✅ 前置条件检查通过"
    echo ""
}

deploy_10() {
    echo "📦 Phase 1: 10% 灰度发布"
    echo "-------------------------"
    
    # 更新镜像
    echo "更新镜像到 $VERSION..."
    kubectl set image deployment/$DEPLOYMENT \
        m03-trajectory=squirrel/m03-trajectory:$VERSION \
        -n $NAMESPACE
    
    # 设置副本数
    echo "设置副本数: 3..."
    kubectl scale deployment/$DEPLOYMENT --replicas=3 -n $NAMESPACE
    
    # 等待就绪
    echo "等待Pod就绪..."
    kubectl rollout status deployment/$DEPLOYMENT -n $NAMESPACE --timeout=300s
    
    echo ""
    echo "✅ Phase 1 完成"
    echo "流量比例: 10%"
    echo "验证命令: ./deploy-canary.sh verify"
    echo ""
}

deploy_50() {
    echo "📦 Phase 2: 50% 灰度发布"
    echo "-------------------------"
    
    # 设置副本数
    echo "设置副本数: 5..."
    kubectl scale deployment/$DEPLOYMENT --replicas=5 -n $NAMESPACE
    
    # 等待就绪
    echo "等待Pod就绪..."
    kubectl rollout status deployment/$DEPLOYMENT -n $NAMESPACE --timeout=300s
    
    echo ""
    echo "✅ Phase 2 完成"
    echo "流量比例: 50%"
    echo "验证命令: ./deploy-canary.sh verify"
    echo ""
}

deploy_100() {
    echo "📦 Phase 3: 100% 全量发布"
    echo "--------------------------"
    
    # 设置副本数
    echo "设置副本数: 10..."
    kubectl scale deployment/$DEPLOYMENT --replicas=10 -n $NAMESPACE
    
    # 等待就绪
    echo "等待Pod就绪..."
    kubectl rollout status deployment/$DEPLOYMENT -n $NAMESPACE --timeout=300s
    
    echo ""
    echo "✅ Phase 3 完成"
    echo "流量比例: 100%"
    echo "版本: $VERSION 已全量发布"
    echo ""
}

show_status() {
    echo "📊 当前发布状态"
    echo "---------------"
    
    kubectl get deployment $DEPLOYMENT -n $NAMESPACE -o wide
    echo ""
    
    kubectl get pods -n $NAMESPACE -l app=m03-trajectory
    echo ""
    
    # 显示镜像版本
    echo "镜像版本:"
    kubectl get deployment $DEPLOYMENT -n $NAMESPACE \
        -o jsonpath='{range .spec.template.spec.containers[*]}{.name}{": "}{.image}{"\n"}{end}'
    echo ""
}

verify_deployment() {
    echo "✅ 验证发布结果"
    echo "--------------"
    
    # 健康检查
    echo "健康检查..."
    kubectl exec -n $NAMESPACE \
        $(kubectl get pods -n $NAMESPACE -l app=m03-trajectory -o jsonpath='{.items[0].metadata.name}') \
        -- curl -s http://localhost:8083/health || echo "❌ 健康检查失败"
    
    # 指标检查
    echo ""
    echo "指标检查 (需端口转发)..."
    echo "kubectl port-forward svc/m03-trajectory 9090:9090 -n $NAMESPACE"
    echo "curl http://localhost:9090/metrics | grep m03_request_total"
    echo ""
}

rollback() {
    echo "⚠️  执行回滚"
    echo "-----------"
    echo "确认回滚? (yes/no)"
    read confirm
    
    if [ "$confirm" = "yes" ]; then
        echo "回滚到上一个版本..."
        kubectl rollout undo deployment/$DEPLOYMENT -n $NAMESPACE
        kubectl rollout status deployment/$DEPLOYMENT -n $NAMESPACE --timeout=300s
        echo "✅ 回滚完成"
    else
        echo "取消回滚"
    fi
}

# 主逻辑
case $PHASE in
    10)
        check_prerequisites
        deploy_10
        ;;
    50)
        check_prerequisites
        deploy_50
        ;;
    100)
        check_prerequisites
        deploy_100
        ;;
    status)
        show_status
        ;;
    verify)
        verify_deployment
        ;;
    rollback)
        rollback
        ;;
    help|*)
        show_help
        ;;
esac
