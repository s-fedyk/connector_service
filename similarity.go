package main

import (
	pb "connector/gen" // import path to your generated files
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
)

var similarityClient pb.ImageServiceClient

func init() {
	log.Print("Initializing similarity client connection...")

	similarityURL := os.Getenv("SIMILARITY_SERVICE_URL")
	if similarityURL == "" {
		log.Fatal("SIMILARITY_SERVICE_URL environment variable is not set")
	}

	conn, err := grpc.Dial(similarityURL, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}

	similarityClient = pb.NewImageServiceClient(conn)

	log.Print("Similarity client connection established!")
}

func similarity(w http.ResponseWriter, r *http.Request) {
	log.Print("Similarity request!")

	file, header, err := r.FormFile("image")

	log.Printf("filename is %v", header.Filename)
	log.Printf("filesize is %v", header.Size)

	buffer := make([]byte, header.Size)

	for {
		_, err := file.Read(buffer)

		if err == io.EOF {
			break
		}
	}

	encodedImage := base64.StdEncoding.EncodeToString(buffer)
	store(encodedImage, header.Filename)

	request := &pb.IdentifyRequest{
		BaseImage: &pb.Image{Url: header.Filename},
	}

	context := context.Background()

	res, err := similarityClient.Identify(context, request)

	if err != nil {
		log.Printf("Identify call failed: %v", err)
		http.Error(w, fmt.Sprintf("Identify call failed: %v", err), http.StatusInternalServerError)
	}

	querySimilar(res.Embedding, context)
}
