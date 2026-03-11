package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	pb "github.com/squirrelawake/m03-trajectory/proto"
)

func main() {
	conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewLocationServiceClient(conn)

	// 并发测试
	var wg sync.WaitGroup
	concurrency := 100
	requests := 1000

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < requests/concurrency; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_, err := client.GetSpeed(ctx, &pb.SpeedRequest{UserId: fmt.Sprintf("user_%d_%d", id, j)})
				cancel()
				if err != nil {
					log.Printf("Error: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("Completed %d requests in %v\n", requests, duration)
	fmt.Printf("RPS: %.2f\n", float64(requests)/duration.Seconds())
}
