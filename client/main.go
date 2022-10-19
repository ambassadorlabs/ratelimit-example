package main

import (
	"context"
	"fmt"
	"log"
	"time"

	// This import path is based on the name declaration in the go.mod,
	// and the gen/proto/go output location in the buf.gen.yaml.
	ratelimitv3 "github.com/emissary-ingress/ratelimit-example/gen/proto/go/envoy/service/ratelimit/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fmt.Println("Ok trying to connect!")
	connectTo := "127.0.0.1:5000"
	conn, err := grpc.Dial(connectTo, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to RatelimitService on %s: %w", connectTo, err)
	}
	log.Println("Connected to", connectTo)

	ratelimitClient := ratelimitv3.NewRateLimitServiceClient(conn)

	request := &ratelimitv3.RateLimitRequest{
		Domain: "test-client",
	}

	// keep making calls every 2 seconds and record response
	fmt.Println("I'm going to make request every 2 seconds until you stop me!!")
	for {

		resp, err := ratelimitClient.ShouldRateLimit(context.Background(), request)
		if err != nil {
			return fmt.Errorf("failed to ShouldRateLimit: %w", err)
		}

		fmt.Printf("Ratelimit code applied is: %v\n", resp.GetOverallCode().String())
		time.Sleep(500 * time.Millisecond)
	}
}
