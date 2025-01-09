package main

import (
	"bytes"
	pb "connector/gen" // import path to your generated files
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/beevik/guid"

	"github.com/prometheus/client_golang/prometheus"
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

type requestContext struct {
	context     *context.Context
	requestGUID string
}

func newRequestContext(context *context.Context) {
	var thisContext requestContext
	thisContext.context = context
	thisContext.requestGUID = guid.NewString()
}

func similarity(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v\nStack trace:\n%s", r, debug.Stack())
		}
	}()

	log.Print("Similarity request!")
	requestsCounter.With(prometheus.Labels{}).Inc()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle OPTIONS method
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	similarityStart := time.Now()
	defer func() {
		similarityDuration := time.Since(similarityStart)
		requestHistogram.With(prometheus.Labels{}).Observe(similarityDuration.Seconds())
	}()

	file, header, err := r.FormFile("image")

	buffer, err := toJPEG(file)

	if err != nil {
		log.Printf("toJPEG call failure, err=(%v)", err)
		http.Error(w, fmt.Sprintf("toJPEG call err=(%v)", err), http.StatusInternalServerError)
	}

	storeS3(buffer.Bytes(), header.Filename)

	request := &pb.IdentifyRequest{
		BaseImage: &pb.Image{Url: header.Filename},
	}

	context := context.Background()

	embeddingStart := time.Now()
	res, err := similarityClient.Identify(context, request)

	deleteS3(header.Filename)
	embeddingDuration := time.Since(embeddingStart)
	modelHistogram.With(prometheus.Labels{}).Observe(embeddingDuration.Seconds())

	if err != nil {
		log.Printf("Identify call failed, err=(%v)", err)
		http.Error(w, fmt.Sprintf("Identify call failed, err=(%v)", err), http.StatusInternalServerError)
		return
	}

	databaseStart := time.Now()
	similarURLs := querySimilar(res.Embedding, context)
	databaseDuration := time.Since(databaseStart)
	databaseHistogram.With(prometheus.Labels{}).Observe(databaseDuration.Seconds())

	jsonWriter := json.NewEncoder(w)
	err = jsonWriter.Encode(similarURLs)

	if err != nil {
		log.Printf("Response encoding failure!, err=(%v)", err)
		http.Error(w, fmt.Sprintf("Response encoding failure!, err=(%v)", err), http.StatusInternalServerError)
		return
	}
}
