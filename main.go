package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/similarity", similarity)
	fmt.Println("Starting server...")
	err := http.ListenAndServe(":80", nil)

	if err != nil {
		fmt.Println(err)
	}

	return
}
