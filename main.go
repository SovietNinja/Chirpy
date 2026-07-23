package main

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
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

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	mux.Handle("GET /metrics", apiCfg.middlewareMetricsPrint())

	mux.Handle("POST /reset", apiCfg.middlewareMetricsReset())

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (c *apiConfig) increment() {
	c.fileserverHits.Add(1)
}

func (c *apiConfig) resetHits() {
	c.fileserverHits.Store(0)
}

func (c *apiConfig) printHits() string {
	counter := c.fileserverHits.Load()
	Text := "Hits: " + strconv.Itoa(int(counter))
	return Text
}

func (c *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.increment()
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) middlewareMetricsPrint() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(c.printHits()))
	})
}

func (c *apiConfig) middlewareMetricsReset() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.resetHits()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("Reset hits metric"))
	})
}
