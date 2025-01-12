package main

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	buckets = []float64{0.1, 0.25, 0.5, 1, 2, 3, 5, 7, 10}

	requestsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "similarity_requests_total",
			Help: "Total number of similarity requests",
		},
		[]string{},
	)

	modelHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "model_request_delay_seconds",
			Help:    "Histogram of delay to embedding models",
			Buckets: buckets,
		},
		[]string{},
	)

	analysisHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "model_request_delay_seconds",
			Help:    "Histogram of delay to embedding models",
			Buckets: buckets,
		},
		[]string{},
	)

	databaseHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_request_delay_seconds",
			Help:    "Histogram of delay to vector database",
			Buckets: buckets,
		},
		[]string{},
	)

	requestHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_request_delay_seconds",
			Help:    "Total similarity request delay",
			Buckets: buckets,
		},
		[]string{},
	)
)

func init() {
	log.Print("Registering prometheus counters...")

	prometheus.MustRegister(requestsCounter)
	prometheus.MustRegister(modelHistogram)
	prometheus.MustRegister(databaseHistogram)
	prometheus.MustRegister(requestHistogram)

	log.Print("Counters registered!")
}
