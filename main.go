package main

import (
	pb "connector/gen" // import path to your generated files
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"

	"google.golang.org/grpc"
)

var redisClient *redis.Client

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
		log.Printf("Identify call failed: %v", err)
		http.Error(w, fmt.Sprintf("Identify call failed: %v", err), http.StatusInternalServerError)
	}

	querySimilar(res.Embedding, context)
}

func main() {
	ctx := context.Background()

	redisClient = redis.NewClient(&redis.Options{
		Addr: "similarity-image-cache.wutpwp.ng.0001.use2.cache.amazonaws.com:6379",
		DB:   0,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to ElastiCache Redis: %v", err)
	} else {
		log.Println("Connected to ElastiCache Redis successfully!")
	}

	http.HandleFunc("/similarity", similarity)
	fmt.Println("Starting server...")
	err = http.ListenAndServe(":80", nil)

	if err != nil {
		fmt.Println(err)
	}

	return
}
