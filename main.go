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

	apiCfg := apiConfig{}

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /api/validate_chirp", handleChirp)

	mux.Handle("GET /admin/metrics", apiCfg.middlewareMetricsPrint())

	mux.Handle("POST /admin/reset", apiCfg.middlewareMetricsReset())

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
