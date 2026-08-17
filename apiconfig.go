package main

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"github.com/SovietNinja/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
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

func (c *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type createUserRequest struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	var req createUserRequest
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	dbuser, err := c.dbQueries.CreateUser(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	user := User{
		ID:        dbuser.ID,
		CreatedAt: dbuser.CreatedAt,
		UpdatedAt: dbuser.UpdatedAt,
		Email:     dbuser.Email,
	}
	respondWithJSON(w, 201, user)
}

func (c *apiConfig) handlerReset() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.platform != "dev" {
			respondWithError(w, 403, "Platform is not dev!")
			return
		}
		c.resetHits()
		err := c.dbQueries.ResetUsers(r.Context())
		if err != nil {
			respondWithError(w, 500, err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("Reset hits metric and users"))
	})
}
