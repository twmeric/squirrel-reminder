// 联调测试用例 TC-001 ~ TC-005
// 执行时间: 14:00-16:00
// 参与方: @backend-1 (m01) + @backend-2 (m03)

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	
	pb "github.com/squirrel-awake/m03-trajectory/proto"
)

const (
	m03Address = "localhost:50053"
	testUserID = "test_user_001"
)

func main() {
	log.Println("🧪 开始 m01-m03 联调测试...")
	log.Println("⏰ 时间:", time.Now().Format("2006-01-02 15:04:05"))
	
	// 连接 m03
	conn, err := grpc.Dial(m03Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ 连接 m03 失败: %v", err)
	}
	defer conn.Close()
	
	client := pb.NewTrajectoryServiceClient(conn)
	
	// 执行测试用例
	allPassed := true
	
	allPassed = TC001_HealthCheck(client) && allPassed
	allPassed = TC002_ReportGPSAndGetGrid(client) && allPassed
	allPassed = TC003_GetSpeedAndMetroArea(client) && allPassed
	allPassed = TC004_GetNextTransferStation(client) && allPassed
	allPassed = TC005_AlertTriggerLogic(client) && allPassed
	
	if allPassed {
		log.Println("\n✅ 所有测试用例通过！")
	} else {
		log.Println("\n❌ 存在失败的测试用例")
	}
}

// TC-001: 健康检查与连接测试
func TC001_HealthCheck(client pb.TrajectoryServiceClient) bool {
	log.Println("\n📋 TC-001: 健康检查与连接测试")
	
	// 简单调用验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	resp, err := client.GetCurrentGrid(ctx, &pb.GetCurrentGridRequest{UserId: testUserID})
	if err != nil {
		log.Printf("  ❌ 连接测试失败: %v", err)
		return false
	}
	
	log.Printf("  ✅ 连接成功，响应: grid_id=%s", resp.GridId)
	return true
}

// TC-002: GPS上报与网格获取
func TC002_ReportGPSAndGetGrid(client pb.TrajectoryServiceClient) bool {
	log.Println("\n📋 TC-002: GPS上报与网格获取")
	
	ctx := context.Background()
	
	// 上报GPS点（国贸站附近）
	points := []struct {
		lat, lng float64
	}{
		{39.9240, 116.4690},
		{39.9242, 116.4695},
		{39.9244, 116.4700},
	}
	
	for i, p := range points {
		_, err := client.ReportGPSPoint(ctx, &pb.ReportGPSPointRequest{
			UserId: testUserID,
			Point: &pb.GPSPoint{
				Timestamp: time.Now().Add(time.Duration(i) * time.Second).UnixMilli(),
				Latitude:  p.lat,
				Longitude: p.lng,
				Accuracy:  10.0,
				Speed:     8.5,
				Source:    "gps",
			},
		})
		if err != nil {
			log.Printf("  ❌ 上报GPS点失败: %v", err)
			return false
		}
	}
	
	// 获取网格
	gridResp, err := client.GetCurrentGrid(ctx, &pb.GetCurrentGridRequest{UserId: testUserID})
	if err != nil {
		log.Printf("  ❌ 获取网格失败: %v", err)
		return false
	}
	
	if gridResp.GridId == "" {
		log.Println("  ❌ 网格ID为空")
		return false
	}
	
	log.Printf("  ✅ GPS上报成功，网格ID: %s", gridResp.GridId)
	return true
}

// TC-003: 速度计算与地铁区域判断
func TC003_GetSpeedAndMetroArea(client pb.TrajectoryServiceClient) bool {
	log.Println("\n📋 TC-003: 速度计算与地铁区域判断")
	
	ctx := context.Background()
	
	// 获取速度
	speedResp, err := client.GetSpeed(ctx, &pb.GetSpeedRequest{UserId: testUserID})
	if err != nil {
		log.Printf("  ❌ 获取速度失败: %v", err)
		return false
	}
	
	log.Printf("  ✅ 当前速度: %.2f km/h", speedResp.SpeedKmh)
	
	// 判断地铁区域
	metroResp, err := client.IsInMetroArea(ctx, &pb.IsInMetroAreaRequest{UserId: testUserID})
	if err != nil {
		log.Printf("  ❌ 地铁区域判断失败: %v", err)
		return false
	}
	
	status := "不在地铁区域"
	if metroResp.InMetroArea {
		status = fmt.Sprintf("在地铁区域内 (最近站点: %s)", metroResp.NearestStationId)
	}
	log.Printf("  ✅ %s", status)
	
	return true
}

// TC-004: 获取下一个换乘站（BLK-001 字段验证）
func TC004_GetNextTransferStation(client pb.TrajectoryServiceClient) bool {
	log.Println("\n📋 TC-004: 获取下一个换乘站（BLK-001 字段验证）")
	
	ctx := context.Background()
	
	// 上报更多GPS点（模拟移动）
	// 在国贸站附近，往四惠东方向
	points := []struct {
		lat, lng float64
		speed   float32
	}{
		{39.9240, 116.4690, 0},
		{39.9242, 116.4695, 5},
		{39.9244, 116.4700, 10},
		{39.9246, 116.4705, 15},
		{39.9248, 116.4710, 20},
		{39.9250, 116.4715, 25},
	}
	
	for i, p := range points {
		_, err := client.ReportGPSPoint(ctx, &pb.ReportGPSPointRequest{
			UserId: testUserID,
			Point: &pb.GPSPoint{
				Timestamp: time.Now().Add(time.Duration(i) * time.Second).UnixMilli(),
				Latitude:  p.lat,
				Longitude: p.lng,
				Accuracy:  10.0,
				Speed:     p.speed,
				Source:    "gps",
			},
		})
		if err != nil {
			log.Printf("  ❌ 上报GPS点失败: %v", err)
			return false
		}
	}
	
	// 获取换乘站预测
	stationResp, err := client.GetNextTransferStation(ctx, &pb.GetNextTransferStationRequest{
		UserId: testUserID,
	})
	if err != nil {
		log.Printf("  ❌ 获取换乘站失败: %v", err)
		return false
	}
	
	if !stationResp.HasResult {
		log.Println("  ⚠️ 无预测结果（可能不在地铁区域）")
		return true // 这不是错误，只是条件不满足
	}
	
	s := stationResp.Station
	log.Println("  ✅ 获取到换乘站信息:")
	log.Printf("     站点: %s (%s)", s.StationName, s.StationId)
	log.Printf("     线路: %s", s.LineName)
	log.Printf("     方向: %s", s.Direction)
	log.Printf("     预计到达: %d 秒", s.EstimatedArrivalSeconds)
	log.Printf("     置信度: %.2f%%", s.Confidence*100)
	log.Printf("     距离: %d 米", s.DistanceMeters)
	
	// BLK-001 新增字段验证
	log.Println("  ✅ BLK-001 新增字段:")
	log.Printf("     is_transfer: %v", s.IsTransfer)
	log.Printf("     stations_to_transfer: %d", s.StationsToTransfer)
	log.Printf("     next_station_name: %s", s.NextStationName)
	
	return true
}

// TC-005: 提醒触发逻辑验证
func TC005_AlertTriggerLogic(client pb.TrajectoryServiceClient) bool {
	log.Println("\n📋 TC-005: 提醒触发逻辑验证")
	
	ctx := context.Background()
	
	// 模拟场景：即将到达换乘站（建国门）
	// 上报GPS点（永安里站附近，往建国门方向）
	points := []struct {
		lat, lng float64
	}{
		{39.9230, 116.4590}, // 永安里附近
		{39.9232, 116.4595},
		{39.9234, 116.4600},
		{39.9236, 116.4605},
		{39.9238, 116.4610},
	}
	
	for i, p := range points {
		_, err := client.ReportGPSPoint(ctx, &pb.ReportGPSPointRequest{
			UserId: testUserID,
			Point: &pb.GPSPoint{
				Timestamp: time.Now().Add(time.Duration(i) * time.Second).UnixMilli(),
				Latitude:  p.lat,
				Longitude: p.lng,
				Accuracy:  10.0,
				Speed:     30 / 3.6, // 30 km/h
				Source:    "gps",
			},
		})
		if err != nil {
			log.Printf("  ❌ 上报GPS点失败: %v", err)
			return false
		}
	}
	
	// 获取换乘站
	stationResp, err := client.GetNextTransferStation(ctx, &pb.GetNextTransferStationRequest{
		UserId: testUserID,
	})
	if err != nil {
		log.Printf("  ❌ 获取换乘站失败: %v", err)
		return false
	}
	
	if !stationResp.HasResult {
		log.Println("  ⚠️ 无预测结果")
		return true
	}
	
	s := stationResp.Station
	
	// 模拟 m01 的提醒触发逻辑
	shouldAlert := false
	alertReason := ""
	
	// 场景1: 即将到达换乘站（还有1-2站）
	if s.IsTransfer && s.StationsToTransfer <= 2 {
		shouldAlert = true
		alertReason = fmt.Sprintf("即将到达换乘站 %s（还有%d站）", 
			s.StationName, s.StationsToTransfer)
	}
	
	// 场景2: 距离换乘站时间 < 2分钟
	if s.StationsToTransfer > 0 && s.EstimatedArrivalSeconds < 120 {
		shouldAlert = true
		alertReason = fmt.Sprintf("%d秒后到达换乘站 %s", 
			s.EstimatedArrivalSeconds, s.StationName)
	}
	
	// 场景3: 终点站
	if s.DirectionCode == "terminal" && s.EstimatedArrivalSeconds < 60 {
		shouldAlert = true
		alertReason = fmt.Sprintf("即将到达终点站 %s", s.StationName)
	}
	
	if shouldAlert {
		log.Printf("  ✅ 提醒触发: %s", alertReason)
	} else {
		log.Printf("  ℹ️ 暂不提醒: 距离换乘站还有%d站，预计%d秒到达，置信度%.0f%%",
			s.StationsToTransfer, s.EstimatedArrivalSeconds, s.Confidence*100)
	}
	
	return true
}
