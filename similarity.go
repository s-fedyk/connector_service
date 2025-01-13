package main

import (
	"bytes"
	analyzer "connector/gen/analyzer"         // import path to your generated files
	embedder "connector/gen/embedder"         // import path to your generated files
	preprocessor "connector/gen/preprocessor" // import path to your generated files
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
	"sync"
	"time"

	"github.com/beevik/guid"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var similarityClient embedder.ImageServiceClient
var analyzerClient analyzer.AnalyzerClient
var preprocessorClient preprocessor.PreprocessorClient

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

	similarityClient = embedder.NewImageServiceClient(conn)

	log.Print("Similarity client connection established!")

	log.Print("Initializing analyzer client connection...")

	analyzerURL := os.Getenv("ANALYZER_SERVICE_URL")
	if analyzerURL == "" {
		log.Fatal("ANALYZER_SERVICE_URL environment variable is not set")
	}

	conn, err = grpc.NewClient(analyzerURL, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("failed to connect to analyzer Service gRPC server: %v", err)
	}

	analyzerClient = analyzer.NewAnalyzerClient(conn)

	log.Print("analyzer client connection established!")

	log.Print("Initializing preprocessor client connection...")

	preprocessorURL := os.Getenv("PREPROCESSOR_SERVICE_URL")
	if preprocessorURL == "" {
		log.Fatal("PREPROCESSOR_SERVICE_URL environment variable is not set")
	}

	conn, err = grpc.NewClient(preprocessorURL, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("failed to connect to Preprocessor Service gRPC server: %v", err)
	}

	preprocessorClient = preprocessor.NewPreprocessorClient(conn)

	log.Print("preprocessor client connection established!")
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
	Analysis    Analysis   `json:"analysis"`
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

type Analysis struct {
	EMOTION string `json:"emotion"`
	GENDER  string `json:"gender"`
	RACE    string `json:"race"`
	AGE     string `json:"age"`
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
		http.Error(w, fmt.Sprintf("Failed to convert image to JPEG", err), http.StatusInternalServerError)
	}

	tempName := guid.NewString()
	context := context.Background()

	storeS3(buffer.Bytes(), tempName)

	preprocessRequest := &preprocessor.PreprocessRequest{
		BaseImage: &preprocessor.Image{Url: tempName},
	}

	preprocessResponse, preprocessErr := preprocessorClient.Preprocess(context, preprocessRequest)

	if preprocessErr != nil {
		log.Printf("Preprocess call failed, err=(%v)", preprocessErr)
		http.Error(w, "Failed to analyze image, please wait and try again", http.StatusInternalServerError)
		return
	}

	processedURL := preprocessResponse.ProcessedImage.Url

	similarityRequest := &embedder.IdentifyRequest{
		BaseImage: &embedder.Image{Url: processedURL},
	}

	analysisRequest := &analyzer.AnalyzeRequest{
		BaseImage: &analyzer.Image{Url: processedURL},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var analysisRes *analyzer.AnalyzeResponse
	var identifyRes *embedder.IdentifyResponse
	var analysisErr, identifyErr error

	go func() {
		defer wg.Done()
		analysisStart := time.Now()
		analysisRes, analysisErr = analyzerClient.Analyze(context, analysisRequest)
		analysisDuration := time.Since(analysisStart)
		analysisHistogram.With(prometheus.Labels{}).Observe(analysisDuration.Seconds())
	}()

	go func() {
		defer wg.Done()
		embeddingStart := time.Now()
		identifyRes, identifyErr = similarityClient.Identify(context, similarityRequest)
		embeddingDuration := time.Since(embeddingStart)
		modelHistogram.With(prometheus.Labels{}).Observe(embeddingDuration.Seconds())
	}()

	wg.Wait()

	go func() {
		deleteS3(tempName)
	}()

	if analysisErr != nil {
		log.Printf("Analyze call failed, err=(%v)", analysisErr)
		http.Error(w, "Failed to analyze image, please wait and try again", http.StatusInternalServerError)
		return
	}

	if identifyErr != nil {
		log.Printf("Identify call failed, err=(%v)", identifyErr)
		http.Error(w, "Failed to create face embedding, please wait and try again", http.StatusInternalServerError)
		return
	}

	databaseStart := time.Now()
	similarURLs := querySimilar(identifyRes.Embedding, context)
	databaseDuration := time.Since(databaseStart)
	databaseHistogram.With(prometheus.Labels{}).Observe(databaseDuration.Seconds())

	scale_inv_y := 1.0 / preprocessResponse.ScaleH
	scale_inv_x := 1.0 / preprocessResponse.ScaleW

	left_eye := Eye{
		X: int32(float32(identifyRes.FacialArea.LeftEye.X) * scale_inv_x),
		Y: int32(float32(identifyRes.FacialArea.LeftEye.Y) * scale_inv_y),
	}

	right_eye := Eye{
		X: int32(float32(identifyRes.FacialArea.RightEye.X) * scale_inv_x),
		Y: int32(float32(identifyRes.FacialArea.RightEye.Y) * scale_inv_y),
	}

	identifyResponse := SimilarityResponse{
		SimilarURLs: similarURLs,
		FacialArea: FacialArea{
			X:         int32(float32(identifyRes.FacialArea.X) * scale_inv_x),
			Y:         int32(float32(identifyRes.FacialArea.Y) * scale_inv_y),
			W:         int32(float32(identifyRes.FacialArea.W) * scale_inv_x),
			H:         int32(float32(identifyRes.FacialArea.H) * scale_inv_y),
			LEFT_EYE:  left_eye,
			RIGHT_EYE: right_eye,
		},
		Analysis: Analysis{
			GENDER:  analysisRes.Gender,
			AGE:     analysisRes.Age,
			RACE:    analysisRes.Race,
			EMOTION: analysisRes.Emotion,
		},
	}

	jsonWriter := json.NewEncoder(w)
	err = jsonWriter.Encode(identifyResponse)

	if err != nil {
		log.Printf("Response encoding failure!, err=(%v)", err)
		http.Error(w, "Failed to create response, please wait and try again", http.StatusInternalServerError)
		return
	}
}
