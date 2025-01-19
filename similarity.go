package main

import (
	"bytes"
	analyzer "connector/gen/analyzer"
	embedder "connector/gen/embedder"
	preprocessor "connector/gen/preprocessor"
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

var similarityClient embedder.EmbedderClient
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

	similarityClient = embedder.NewEmbedderClient(conn)

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
	JobID       string     `json:"job_id"`
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

type Analysis struct {
	EMOTION string `json:"emotion"`
	GENDER  string `json:"gender"`
	RACE    string `json:"race"`
	AGE     string `json:"age"`
}

var analysisStore = sync.Map{}

type AnalysisJob struct {
	Status      string
	CreatedAt   time.Time
	Result      string
	ErrorString string
}

func storeInitialAnalysisJob(jobID string) {
	analysisStore.Store(jobID, &AnalysisJob{
		Status:    "pending",
		CreatedAt: time.Now(),
	})
}

func setAnalysisField(jobID, value string) {
	val, ok := analysisStore.Load(jobID)
	if !ok {
		return
	}
	job := val.(*AnalysisJob)

	job.Status = "ready"
	job.Result = value

	analysisStore.Store(jobID, job)
}

func similarity(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v\nStack trace:\n%s", r, debug.Stack())
		}
	}()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Print("Similarity request!")
	requestsCounter.With(prometheus.Labels{}).Inc()

	similarityStart := time.Now()
	defer func() {
		similarityDuration := time.Since(similarityStart)
		requestHistogram.With(prometheus.Labels{}).Observe(similarityDuration.Seconds())
	}()

	file, _, err := r.FormFile("image")

	buffer, err := toJPEG(file)

	if err != nil {
		log.Printf("toJPEG call failure, err=(%v)", err)
		http.Error(w, "Failed to convert image to expected format. Only PNG and JPEG are supported.", http.StatusInternalServerError)
	}

	jobID := guid.NewString()
	context := context.Background()

	storeS3(buffer.Bytes(), jobID)

	preprocessRequest := &preprocessor.PreprocessRequest{
		BaseImage: &preprocessor.Image{Url: jobID},
	}

	preprocessResponse, preprocessErr := preprocessorClient.Preprocess(context, preprocessRequest)

	if preprocessErr != nil {
		log.Printf("Preprocess call failed, err=(%v)", preprocessErr)
		http.Error(w, "Failed to preprocess image, please wait and try again", http.StatusInternalServerError)
		return
	}

	processedID := preprocessResponse.ProcessedImage.Url

	similarityRequest := &embedder.EmbedRequest{
		BaseImage: &embedder.Image{Url: processedID},
	}

	storeInitialAnalysisJob(fmt.Sprintf("%v-age", jobID))
	storeInitialAnalysisJob(fmt.Sprintf("%v-race", jobID))
	storeInitialAnalysisJob(fmt.Sprintf("%v-gender", jobID))
	storeInitialAnalysisJob(fmt.Sprintf("%v-emotion", jobID))

	genderAnalysisRequest := &analyzer.AnalyzeRequest{
		BaseImage: &analyzer.Image{Url: processedID},
		Model:     []string{"gender"},
	}
	ageAnalysisRequest := &analyzer.AnalyzeRequest{
		BaseImage: &analyzer.Image{Url: processedID},
		Model:     []string{"age"},
	}
	raceAnalysisRequest := &analyzer.AnalyzeRequest{
		BaseImage: &analyzer.Image{Url: processedID},
		Model:     []string{"race"},
	}
	emotionAnalysisRequest := &analyzer.AnalyzeRequest{
		BaseImage: &analyzer.Image{Url: processedID},
		Model:     []string{"emotion"},
	}

	var embedRes *embedder.EmbedResponse
	var embedErr error
	var similarURLs []string

	go func() {
		analysisStart := time.Now()
		ageAnalysisRes, ageAnalysisErr := analyzerClient.Analyze(context, ageAnalysisRequest)
		analysisDuration := time.Since(analysisStart)
		analysisHistogram.With(prometheus.Labels{}).Observe(analysisDuration.Seconds())

		if ageAnalysisErr != nil {
			log.Printf("Analyze call failed, err=(%v)", ageAnalysisErr)
			http.Error(w, "Failed to analyze age, please wait and try again", http.StatusInternalServerError)
			return
		}
		setAnalysisField(fmt.Sprintf("%v-age", jobID), ageAnalysisRes.Results[0].Result)
	}()
	go func() {
		analysisStart := time.Now()
		raceAnalysisRes, raceAnalysisErr := analyzerClient.Analyze(context, raceAnalysisRequest)
		analysisDuration := time.Since(analysisStart)
		analysisHistogram.With(prometheus.Labels{}).Observe(analysisDuration.Seconds())

		if raceAnalysisErr != nil {
			log.Printf("Analyze call failed, err=(%v)", raceAnalysisErr)
			http.Error(w, "Failed to analyze race, please wait and try again", http.StatusInternalServerError)
			return
		}
		setAnalysisField(fmt.Sprintf("%v-race", jobID), raceAnalysisRes.Results[0].Result)
	}()
	go func() {
		analysisStart := time.Now()
		emotionAnalysisRes, emotionAnalysisErr := analyzerClient.Analyze(context, emotionAnalysisRequest)
		analysisDuration := time.Since(analysisStart)
		analysisHistogram.With(prometheus.Labels{}).Observe(analysisDuration.Seconds())

		if emotionAnalysisErr != nil {
			log.Printf("Analyze call failed, err=(%v)", emotionAnalysisErr)
			http.Error(w, "Failed to analyze emotion, please wait and try again", http.StatusInternalServerError)
			return
		}
		setAnalysisField(fmt.Sprintf("%v-emotion", jobID), emotionAnalysisRes.Results[0].Result)
	}()

	go func() {
		analysisStart := time.Now()
		genderAnalysisRes, genderAnalysisErr := analyzerClient.Analyze(context, genderAnalysisRequest)
		analysisDuration := time.Since(analysisStart)
		analysisHistogram.With(prometheus.Labels{}).Observe(analysisDuration.Seconds())

		if genderAnalysisErr != nil {
			log.Printf("Analyze call failed, err=(%v)", genderAnalysisErr)
			return
		}
		setAnalysisField(fmt.Sprintf("%v-gender", jobID), genderAnalysisRes.Results[0].Result)
	}()

	embeddingStart := time.Now()
	embedRes, embedErr = similarityClient.Embed(context, similarityRequest)
	embeddingDuration := time.Since(embeddingStart)
	modelHistogram.With(prometheus.Labels{}).Observe(embeddingDuration.Seconds())

	if embedErr == nil {
		databaseStart := time.Now()
		similarURLs = querySimilar(embedRes.Embedding, context)
		databaseDuration := time.Since(databaseStart)
		databaseHistogram.With(prometheus.Labels{}).Observe(databaseDuration.Seconds())
	}

	go func() {
		deleteS3(jobID)
		deleteS3(processedID)
	}()

	if embedErr != nil {
		log.Printf("Embed call failed, err=(%v)", embedErr)
		http.Error(w, "Failed to create face embedding, please wait and try again", http.StatusInternalServerError)
		return
	}

	left_eye := Eye{
		X: int32(float32(preprocessResponse.FacialArea.LeftEye.X)),
		Y: int32(float32(preprocessResponse.FacialArea.LeftEye.Y)),
	}

	right_eye := Eye{
		X: int32(float32(preprocessResponse.FacialArea.RightEye.X)),
		Y: int32(float32(preprocessResponse.FacialArea.RightEye.Y)),
	}

	response := SimilarityResponse{
		JobID:       jobID,
		SimilarURLs: similarURLs,
		FacialArea: FacialArea{
			X:         int32(float32(preprocessResponse.FacialArea.X)),
			Y:         int32(float32(preprocessResponse.FacialArea.Y)),
			W:         int32(float32(preprocessResponse.FacialArea.W)),
			H:         int32(float32(preprocessResponse.FacialArea.H)),
			LEFT_EYE:  left_eye,
			RIGHT_EYE: right_eye,
		},
	}

	jsonWriter := json.NewEncoder(w)
	err = jsonWriter.Encode(response)

	if err != nil {
		log.Printf("Response encoding failure!, err=(%v)", err)
		http.Error(w, "Failed to create response, please wait and try again", http.StatusInternalServerError)
		return
	}
}

type JobResponse struct {
	JobID  string `json:"job_id"`
	Result string `json:"result"`
	Error  string `json:"error"`
}

func checkJob(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	jobID := r.URL.Query().Get("jobID")

	log.Printf("Checking on job %v", jobID)

	if jobID == "" {
		http.Error(w, "Missing jobID param", http.StatusBadRequest)
		return
	}

	response := JobResponse{
		JobID:  jobID,
		Result: "--",
		Error:  "",
	}

	result, ok := analysisStore.Load(jobID)

	if !ok {
		log.Printf("No job found, %v", jobID)
		response.Error = "No job found"
	} else {
		job := result.(*AnalysisJob)

		log.Printf("job found! %v", jobID)
		if job.Status == "pending" {
			log.Printf("Job still pending")
		} else {
			log.Printf("Job has result, %v", job.Result)
			response.Result = job.Result
			analysisStore.Delete(jobID)
		}
	}

	log.Printf("Result is %v", response)
	jsonWriter := json.NewEncoder(w)
	err := jsonWriter.Encode(response)

	if err != nil {
		log.Printf("Response encoding failure!, err=(%v)", err)
		http.Error(w, "Failed to create response, please wait and try again", http.StatusInternalServerError)
		return
	}
}
