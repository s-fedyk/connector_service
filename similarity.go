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

type SimilarityResponse struct {
	SimilarURLs []string   `json:"similar_urls"`
	FacialArea  FacialArea `json:"facial_area"`
}

type Eye struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

type FacialArea struct {
	X         int32 `json:"x"`
	Y         int32 `json:"y"`
	W         int32 `json:"w"`
	H         int32 `json:"h"`
	LEFT_EYE  Eye   `json:"left_eye"`
	RIGHT_EYE Eye   `json:"right_eye"`
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

	file, _, err := r.FormFile("image")

	buffer, err := toJPEG(file)

	if err != nil {
		log.Printf("toJPEG call failure, err=(%v)", err)
		http.Error(w, fmt.Sprintf("toJPEG call err=(%v)", err), http.StatusInternalServerError)
	}

	tempName := guid.NewString()

	storeS3(buffer.Bytes(), tempName)

	request := &pb.IdentifyRequest{
		BaseImage: &pb.Image{Url: tempName},
	}

	context := context.Background()

	embeddingStart := time.Now()
	res, err := similarityClient.Identify(context, request)

	deleteS3(tempName)
	embeddingDuration := time.Since(embeddingStart)
	modelHistogram.With(prometheus.Labels{}).Observe(embeddingDuration.Seconds())

	if err != nil {
		log.Printf("Identify call failed, err=(%v)", err)
		http.Error(w, fmt.Sprintf("Identify call failed, err=(%v)", err), http.StatusInternalServerError)
		return
	}

	print("Response: %v", &res.FacialArea)

	databaseStart := time.Now()
	similarURLs := querySimilar(res.Embedding, context)
	databaseDuration := time.Since(databaseStart)
	databaseHistogram.With(prometheus.Labels{}).Observe(databaseDuration.Seconds())

	left_eye := Eye{
		X: res.FacialArea.LeftEye.X,
		Y: res.FacialArea.LeftEye.Y,
	}

	right_eye := Eye{
		X: res.FacialArea.RightEye.X,
		Y: res.FacialArea.RightEye.Y,
	}

	similarityResponse := SimilarityResponse{
		SimilarURLs: similarURLs,
		FacialArea: FacialArea{
			X:         res.FacialArea.X,
			Y:         res.FacialArea.Y,
			W:         res.FacialArea.W,
			H:         res.FacialArea.H,
			LEFT_EYE:  left_eye,
			RIGHT_EYE: right_eye,
		},
	}

	jsonWriter := json.NewEncoder(w)
	err = jsonWriter.Encode(similarityResponse)

	if err != nil {
		log.Printf("Response encoding failure!, err=(%v)", err)
		http.Error(w, fmt.Sprintf("Response encoding failure!, err=(%v)", err), http.StatusInternalServerError)
		return
	}
}
