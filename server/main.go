package main

import (
	"context"
	"fmt"
	"log"
	// "math"
	"net"
	"time"
	"github.com/honeycombio/leakybucket"

	// This import path is based on the name declaration in the go.mod,
	// and the gen/proto/go output location in the buf.gen.yaml.
	corev3 "github.com/emissary-ingress/ratelimit-example/gen/proto/go/envoy/config/core/v3"
	ratelimitv3 "github.com/emissary-ingress/ratelimit-example/gen/proto/go/envoy/service/ratelimit/v3"

	"google.golang.org/grpc"
)

var bucket leakybucket.Bucket

func main() {
	bucket = leakybucket.Bucket{
		Capacity:    1,
		DrainAmount: 8,
		DrainPeriod: time.Second,
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	listenOn := ":5000"
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenOn, err)
	}

	server := grpc.NewServer()

	ratelimitv3.RegisterRateLimitServiceServer(server, &rateLimitServer{})
	log.Println("Listening on", listenOn)
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

// rateLimitServer implements the v3.RatelimitService API from EnvoyProxy.
type rateLimitServer struct {
	ratelimitv3.UnimplementedRateLimitServiceServer
}

func (rls *rateLimitServer) ShouldRateLimit(ctx context.Context, request *ratelimitv3.RateLimitRequest) (*ratelimitv3.RateLimitResponse, error) {
	//Note: this matches the previous behavior as outlined here:https://github.com/emissary-ingress/emissary/blob/v2.1.0/docker/test-ratelimit/server.js

	allow := false

	response := &ratelimitv3.RateLimitResponse{
		OverallCode:          0,
		Statuses:             make([]*ratelimitv3.RateLimitResponse_DescriptorStatus, len(request.Descriptors)),
		ResponseHeadersToAdd: make([]*corev3.HeaderValue, 0),
	}

	fmt.Println("========>")
	fmt.Println(request.Domain)

	for _, descriptor := range request.Descriptors {
		for _, entry := range descriptor.Entries {
			fmt.Printf("  %s = %s\n", entry.Key, entry.Value)
			if entry.Key == "x-emissary-test-allow" && entry.Value == "true" {
				allow = true
				break
			}		
		}
	}

	if err := bucket.Add(); err == nil {
		fmt.Println("Bucket add true")
		allow = true
	} else {
		fmt.Println("Bucket add false")
	}

	if allow {
		response.OverallCode = ratelimitv3.RateLimitResponse_OK
	} else {
		response.OverallCode = ratelimitv3.RateLimitResponse_OVER_LIMIT
	}
	fmt.Println("<========")
	fmt.Println(response.OverallCode)
	//TODO: print debug json output for debugging
	return response, nil
}

// // PutPet adds the pet associated with the given request into the PetStore.
// func (rls *rateLimitServer) PutPet(ctx context.Context, req *petv1.PutPetRequest) (*petv1.PutPetResponse, error) {
// 	name := req.GetName()
// 	petType := req.GetPetType()
// 	log.Println("Got a request to create a", petType, "named", name)

// 	return &petv1.PutPetResponse{}, nil
// }
