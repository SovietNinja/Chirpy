package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	mux.Handle("/", http.FileServer(http.Dir(".")))

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
