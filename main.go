package main

import (
	pb "connector/gen" // import path to your generated files
	"context"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/grpc"
)

// server implements ImageServiceServer (from imageservice_grpc.pb.go)
type server struct {
	pb.UnimplementedImageServiceServer
}

func similarity(w http.ResponseWriter, r *http.Request) {
	log.Print("Similarity request")
	context := context.Background()

	conn, err := grpc.Dial("similarity-service.default.svc.cluster.local:80", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}

	imageClient := pb.NewImageServiceClient(conn)
	request := &pb.IdentifyRequest{
		BaseImage: &pb.Image{Url: "hello"},
	}

	res, err := imageClient.Identify(context, request)

	if err != nil {
		http.Error(w, fmt.Sprintf("Identify call failed: %v", err), http.StatusInternalServerError)
	}

	querySimilar(res.Embedding, context)
}

func main() {
	http.HandleFunc("/similarity", similarity)
	fmt.Println("Starting server...")
	err := http.ListenAndServe(":80", nil)

	if err != nil {
		fmt.Println(err)
	}

	return
}
