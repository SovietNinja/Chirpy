package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (c *apiConfig) increment() {
	c.fileserverHits.Add(1)
}

func (c *apiConfig) resetHits() {
	c.fileserverHits.Store(0)
}

// func (c *apiConfig) printHits() string {
// 	counter := c.fileserverHits.Load()
// 	Text := "Hits: " + strconv.Itoa(int(counter))
// 	return Text
// }

func (c *apiConfig) hitsCount() int {
	counter := c.fileserverHits.Load()
	return int(counter)
}

func (c *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.increment()
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) middlewareMetricsPrint() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		metric_template := `
			<html>
				<body>
					<h1>Welcome, Chirpy Admin</h1>
					<p>Chirpy has been visited %d times!</p>
				</body>
			</html>`
		metric_template = fmt.Sprintf(metric_template, c.hitsCount())
		w.Write([]byte(metric_template))
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

type chirp struct {
	Body string `json:"body"`
}

func handleChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirp := chirp{}
	err := decoder.Decode(&chirp)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	type validResponse struct {
		Valid bool `json:"valid"`
	}
	resp := validResponse{
		Valid: true,
	}
	respondWithJSON(w, 200, resp)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	resp := errorResponse{
		Error: msg,
	}
	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}
