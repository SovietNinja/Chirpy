package main

import (
	"fmt"
	"net/http"
)

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
