package main

import (
    "context"
    "fmt"
    "log"
    "net"

    pb "my-backend/proto" // import path to your generated files

    "google.golang.org/grpc"
)

// server implements ImageServiceServer (from imageservice_grpc.pb.go)
type server struct {
    pb.UnimplementedImageServiceServer
}

func (s *server) Identify(ctx context.Context, req *pb.IdentifyRequest) (*pb.IdentifyResponse, error) {
    log.Printf("Identify called with base_image url: %s", req.BaseImage.Url)

    // Just returning a dummy embedding for demonstration
    return &pb.IdentifyResponse{
        Embedding: []float32{0.1, 0.2, 0.3},
    }, nil
}

func main() {
    // Listen on port 50051 (typical gRPC port)
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    // Create a gRPC server
    grpcServer := grpc.NewServer()

    // Register our service implementation with the gRPC server
    pb.RegisterImageServiceServer(grpcServer, &server{})

    log.Println("Server listening on port 50051...")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
