package main

import (
	"bytes"
	pb "connector/gen" // import path to your generated files
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var similarityClient pb.ImageServiceClient

func init() {
	log.Print("Initializing similarity client connection...")

	similarityURL := os.Getenv("SIMILARITY_SERVICE_URL")
	if similarityURL == "" {
		log.Fatal("SIMILARITY_SERVICE_URL environment variable is not set")
	}

	conn, err := grpc.NewClient(similarityURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
  
	if err != nil {
		log.Fatalf("failed to connect to Similarity Service gRPC server: %v", err)
	}

	similarityClient = pb.NewImageServiceClient(conn)

	log.Print("Similarity client connection established!")
}

func toJPEG(file multipart.File) (*bytes.Buffer, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	var jpegBuffer bytes.Buffer
	err = jpeg.Encode(&jpegBuffer, img, nil)
	if err != nil {
    return nil, err
	}

  return &jpegBuffer, nil
}

func similarity(w http.ResponseWriter, r *http.Request) {
	log.Print("Similarity request!")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle OPTIONS method
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	file, header, err := r.FormFile("image")

	buffer,err := toJPEG(file)

  if (err != nil) {
    http.Error(w, fmt.Sprintf("toJPEG call failed: %v", err), http.StatusInternalServerError)
  }

	store(buffer.Bytes(), header.Filename)

	request := &pb.IdentifyRequest{
		BaseImage: &pb.Image{Url: header.Filename},
	}

	context := context.Background()
	res, err := similarityClient.Identify(context, request)

	if err != nil {
		log.Printf("Identify call failed: %v", err)
		http.Error(w, fmt.Sprintf("Identify call failed: %v", err), http.StatusInternalServerError)
    return
	}

  similarURLs := querySimilar(res.Embedding, context)

  jsonWriter := json.NewEncoder(w)
  jsonWriter.Encode(similarURLs)
}
