package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func health(w http.ResponseWriter, r *http.Request) {
	log.Print("health check")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Healthy"))
}

func main() {
	http.HandleFunc("/similarity", similarity)
	http.HandleFunc("/checkJob", checkJob)
	http.HandleFunc("/health", health)
	http.Handle("/metrics", promhttp.Handler()) // Expose metrics endpoint

	fmt.Println("Starting server...")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		fmt.Println(err)
	}
}
