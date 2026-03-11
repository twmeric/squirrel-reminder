package main

import (
	"log"
	"os"

	grpcserver "squirrel-m01/internal/grpc"
)

func main() {
	port := os.Getenv("M01_PORT")
	if port == "" {
		port = ":50051"
	}

	log.Printf("[m01] Starting State Engine on %s", port)
	
	if err := grpcserver.StartServer(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
