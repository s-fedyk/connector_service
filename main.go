package main

import (
	pb "connector/gen" // import path to your generated files
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"log"
	"net/http"

	"google.golang.org/grpc"
)

// server implements ImageServiceServer (from imageservice_grpc.pb.go)
type server struct {
	pb.UnimplementedImageServiceServer
}

func similarity(w http.ResponseWriter, r *http.Request) {
	log.Print("SimilarityRequest!")

	client, err := client.NewClient(context.Background(),
		client.Config{
			Address: "milvus-demo.default.svc.cluster.local:19530",
		},
	)

	log.Print("Connection to db established!")

	collectionPresent, err := client.HasCollection(context.Background(), "image_embeddings")

	if err != nil {
		log.Fatalf("Failed to connect to milvus db: %v", err)
	}

	log.Printf("Is image collection present? %v", collectionPresent)

	conn, err := grpc.Dial("similarity-service.default.svc.cluster.local:80", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}
	// Typically you'd defer conn.Close() here, but if you want the connection
	// to remain open until the HTTP server shuts down, you might handle that differently.

	// 2. Create the ImageService client
	imageClient := pb.NewImageServiceClient(conn)
	request := &pb.IdentifyRequest{
		BaseImage: &pb.Image{Url: "hello"},
	}

	res, err := imageClient.Identify(context.Background(), request)

	if err != nil {
		http.Error(w, fmt.Sprintf("Identify call failed: %v", err), http.StatusInternalServerError)
	}

	fmt.Println(res)
	fmt.Println(err)
}

func main() {
	http.HandleFunc("/similarity", similarity)
	fmt.Println("Starting...")

	err := http.ListenAndServe(":80", nil)

	if err != nil {
		fmt.Println(err)
	}

	return
}
