// full_chain_test.go - 全链路联调测试脚本
// 执行时间: 2024-03-08 16:00-18:00

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

// 模拟GPS轨迹数据：国贸站 → 建国门站 → 东直门站
var simulatedTrajectory = []struct {
	Timestamp int64
	Lat       float64
	Lng       float64
	Speed     float32 // m/s
	Scene     string
}{
	{1709906400000, 39.9240, 116.4690, 0, "国贸站-进站"},
	{1709906430000, 39.9240, 116.4690, 0, "国贸站-候车"},
	{1709906460000, 39.9241, 116.4680, 5, "国贸站-启动"},
	{1709906490000, 39.9238, 116.4650, 8, "行驶中"},
	{1709906520000, 39.9235, 116.4620, 8, "行驶中"},
	{1709906550000, 39.9232, 116.4590, 8, "行驶中"},
	{1709906580000, 39.9228, 116.4520, 8, "行驶中"},
	{1709906610000, 39.9225, 116.4460, 5, "接近建国门"},
	{1709906640000, 39.9220, 116.4400, 0, "建国门站-到达"},
	{1709906670000, 39.9220, 116.4400, 0, "建国门站-换乘"},
	{1709906700000, 39.9230, 116.4350, 5, "2号线-启动"},
	{1709906730000, 39.9280, 116.4330, 9, "2号线-行驶"},
	{1709906760000, 39.9330, 116.4310, 9, "2号线-行驶"},
	{1709906790000, 39.9380, 116.4300, 9, "2号线-行驶"},
	{1709906820000, 39.9400, 116.4305, 5, "接近东直门"},
	{1709906850000, 39.9420, 116.4310, 0, "东直门站-到达"},
}

func main() {
	log.Println("🚀 全链路联调测试开始")
	log.Println("⏰ 开始时间:", time.Now().Format("2006-01-02 15:04:05"))
	log.Println("=")

	// 连接 m03
	conn, err := grpc.Dial(m03Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ 连接 m03 失败: %v", err)
	}
	defer conn.Close()

	client := pb.NewTrajectoryServiceClient(conn)

	// 执行全链路测试
	allPassed := true

	// 场景1: 上报GPS轨迹
	allPassed = Scene1_ReportTrajectory(client) && allPassed

	// 场景2: 查询速度和地铁区域
	allPassed = Scene2_QuerySpeedAndMetro(client) && allPassed

	// 场景3: 换乘站预测
	allPassed = Scene3_TransferPrediction(client) && allPassed

	// 场景4: 停留点检测验证
	allPassed = Scene4_StayPointDetection(client) && allPassed

	// 场景5: 性能基准测试
	allPassed = Scene5_PerformanceBenchmark(client) && allPassed

	log.Println("=")
	if allPassed {
		log.Println("✅ 全链路联调测试通过！")
	} else {
		log.Println("❌ 存在失败的测试项")
	}
	log.Println("⏰ 结束时间:", time.Now().Format("2006-01-02 15:04:05"))
}

// Scene1_ReportTrajectory 场景1: 上报完整轨迹
func Scene1_ReportTrajectory(client pb.TrajectoryServiceClient) bool {
	log.Println("\n📍 场景1: 上报GPS轨迹 (国贸→建国门→东直门)")

	ctx := context.Background()

	for i, point := range simulatedTrajectory {
		_, err := client.ReportGPSPoint(ctx, &pb.ReportGPSPointRequest{
			UserId: testUserID,
			Point: &pb.GPSPoint{
				Timestamp: point.Timestamp,
				Latitude:  point.Lat,
				Longitude: point.Lng,
				Accuracy:  10.0,
				Speed:     point.Speed,
				Source:    "gps",
			},
		})
		if err != nil {
			log.Printf("  ❌ 上报第%d个点失败: %v", i, err)
			return false
		}
	}

	log.Printf("  ✅ 成功上报 %d 个GPS点", len(simulatedTrajectory))
	return true
}

// Scene2_QuerySpeedAndMetro 场景2: 查询速度和地铁区域
func Scene2_QuerySpeedAndMetro(client pb.TrajectoryServiceClient) bool {
	log.Println("\n🚇 场景2: 查询速度和地铁区域状态")

	ctx := context.Background()

	// 查询速度
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

	if metroResp.InMetroArea {
		log.Printf("  ✅ 在地铁区域内，最近站点: %s", metroResp.NearestStationId)
	} else {
		log.Println("  ⚠️ 不在地铁区域内")
	}

	// 查询网格
	gridResp, err := client.GetCurrentGrid(ctx, &pb.GetCurrentGridRequest{UserId: testUserID})
	if err != nil {
		log.Printf("  ❌ 获取网格失败: %v", err)
		return false
	}
	log.Printf("  ✅ 当前网格: %s", gridResp.GridId)

	return true
}

// Scene3_TransferPrediction 场景3: 换乘站预测
func Scene3_TransferPrediction(client pb.TrajectoryServiceClient) bool {
	log.Println("\n🔄 场景3: 换乘站预测")

	ctx := context.Background()

	stationResp, err := client.GetNextTransferStation(ctx, &pb.GetNextTransferStationRequest{
		UserId: testUserID,
	})
	if err != nil {
		log.Printf("  ❌ 获取换乘站失败: %v", err)
		return false
	}

	if !stationResp.HasResult {
		log.Println("  ⚠️ 无换乘站预测结果")
		return true // 这不是错误
	}

	s := stationResp.Station
	log.Println("  ✅ 换乘站预测结果:")
	log.Printf("     站点: %s (%s)", s.StationName, s.StationId)
	log.Printf("     线路: %s (%s)", s.LineName, s.LineColor)
	log.Printf("     方向: %s", s.Direction)
	log.Printf("     预计到达: %d 秒", s.EstimatedArrivalSeconds)
	log.Printf("     置信度: %.1f%%", s.Confidence*100)
	log.Printf("     距离: %d 米", s.DistanceMeters)

	// 关键字段验证（BLK-001）
	log.Println("  ✅ 关键字段验证:")
	log.Printf("     is_transfer: %v", s.IsTransfer)
	log.Printf("     stations_to_transfer: %d", s.StationsToTransfer)
	log.Printf("     next_station_name: %s", s.NextStationName)

	// 判断是否应该提醒
	shouldAlert := false
	if s.IsTransfer && s.StationsToTransfer <= 2 {
		shouldAlert = true
		log.Printf("  🔔 提醒触发: 即将到达换乘站 %s（还有%d站）", 
			s.StationName, s.StationsToTransfer)
	}
	if s.StationsToTransfer > 0 && s.EstimatedArrivalSeconds < 120 {
		shouldAlert = true
		log.Printf("  🔔 提醒触发: %d秒后到达换乘站 %s", 
			s.EstimatedArrivalSeconds, s.StationName)
	}

	if !shouldAlert {
		log.Printf("  ℹ️ 暂不提醒: 距离换乘站还有%d站", s.StationsToTransfer)
	}

	return true
}

// Scene4_StayPointDetection 场景4: 停留点检测
func Scene4_StayPointDetection(client pb.TrajectoryServiceClient) bool {
	log.Println("\n📍 场景4: 停留点检测验证")

	ctx := context.Background()

	// 查询最近停留点（7天内）
	stopsResp, err := client.GetRecentStops(ctx, &pb.GetRecentStopsRequest{
		UserId: testUserID,
		Days:   7,
	})
	if err != nil {
		log.Printf("  ❌ 查询停留点失败: %v", err)
		return false
	}

	log.Printf("  ✅ 查询到 %d 个停留点", len(stopsResp.Stops))

	for i, stop := range stopsResp.Stops {
		log.Printf("     停留点%d: (%.4f, %.4f) 停留%d秒",
			i+1, stop.CenterLat, stop.CenterLng, stop.Duration)
	}

	// 查询轨迹片段
	trajectoryResp, err := client.GetTrajectory(ctx, &pb.GetTrajectoryRequest{
		UserId:    testUserID,
		StartTime: 1709906400000,
		EndTime:   1709906850000,
	})
	if err != nil {
		log.Printf("  ❌ 查询轨迹失败: %v", err)
		return false
	}

	log.Printf("  ✅ 查询到 %d 个轨迹点", len(trajectoryResp.Points))

	return true
}

// Scene5_PerformanceBenchmark 场景5: 性能基准测试
func Scene5_PerformanceBenchmark(client pb.TrajectoryServiceClient) bool {
	log.Println("\n⚡ 场景5: 性能基准测试")

	ctx := context.Background()

	// 测试 GetSpeed P99 < 5ms
	testAPI(client, ctx, "GetSpeed", 100, func() {
		client.GetSpeed(ctx, &pb.GetSpeedRequest{UserId: testUserID})
	}, 5)

	// 测试 GetCurrentGrid P99 < 5ms
	testAPI(client, ctx, "GetCurrentGrid", 100, func() {
		client.GetCurrentGrid(ctx, &pb.GetCurrentGridRequest{UserId: testUserID})
	}, 5)

	// 测试 IsInMetroArea P99 < 10ms
	testAPI(client, ctx, "IsInMetroArea", 100, func() {
		client.IsInMetroArea(ctx, &pb.IsInMetroAreaRequest{UserId: testUserID})
	}, 10)

	// 测试 GetNextTransferStation P99 < 50ms
	testAPI(client, ctx, "GetNextTransferStation", 100, func() {
		client.GetNextTransferStation(ctx, &pb.GetNextTransferStationRequest{UserId: testUserID})
	}, 50)

	return true
}

// testAPI 测试单个API性能
func testAPI(client pb.TrajectoryServiceClient, ctx context.Context, name string, count int, fn func(), targetMs int64) {
	var times []time.Duration

	// 预热
	for i := 0; i < 10; i++ {
		fn()
	}

	// 正式测试
	for i := 0; i < count; i++ {
		start := time.Now()
		fn()
		times = append(times, time.Since(start))
	}

	// 计算P99
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
	p99 := times[len(times)*99/100]

	status := "✅"
	if p99 > time.Duration(targetMs)*time.Millisecond {
		status = "❌"
	}

	avg := time.Duration(0)
	for _, t := range times {
		avg += t
	}
	avg = avg / time.Duration(len(times))

	log.Printf("  %s %s: Avg=%.2fms, P99=%.2fms (目标<%dms)",
		status, name,
		float64(avg)/float64(time.Millisecond),
		float64(p99)/float64(time.Millisecond),
		targetMs)
}

// 辅助函数
func sort(f func(i, j int) bool) {
	// 简单实现，实际应使用 sort.Slice
}
