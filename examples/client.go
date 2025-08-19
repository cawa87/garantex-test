package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/cawa87/garantex-test/gen/go/rate_service.v1"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewRateServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("=== Garantex Rate Service Client ===")

	fmt.Println("\n1. Health Check:")
	healthResp, err := client.HealthCheck(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		log.Fatalf("HealthCheck failed: %v", err)
	}
	fmt.Printf("Status: %s\n", healthResp.Status)
	for key, value := range healthResp.Details {
		fmt.Printf("  %s: %s\n", key, value)
	}

	fmt.Println("\n2. Get Current Rates:")
	ratesResp, err := client.GetRates(ctx, &pb.GetRatesRequest{})
	if err != nil {
		log.Fatalf("GetRates failed: %v", err)
	}
	fmt.Printf("Ask: $%.2f\n", ratesResp.Ask)
	fmt.Printf("Bid: $%.2f\n", ratesResp.Bid)
	fmt.Printf("Timestamp: %s\n", ratesResp.Timestamp.AsTime().Format(time.RFC3339))

	fmt.Println("\n✅ Service is working correctly!")
}
