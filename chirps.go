package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/SovietNinja/Chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	Id         uuid.UUID `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Body       string    `json:"body"`
	User_id    uuid.UUID `json:"user_id"`
}

func (c *apiConfig) handleChirp(w http.ResponseWriter, r *http.Request) {
	user_id, err := c.validateUser(r)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	decoder := json.NewDecoder(r.Body)
	type chirpRequest struct {
		Body string `json:"body"`
	}
	chirpReq := chirpRequest{}
	err = decoder.Decode(&chirpReq)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	if len(chirpReq.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	newChirpParams := database.CreateChirpParams{
		Body:   profanityCensor(chirpReq.Body),
		UserID: user_id,
	}

	newChirp, err := c.dbQueries.CreateChirp(r.Context(), newChirpParams)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	chirp := Chirp{
		Id:         newChirp.ID,
		Created_at: newChirp.CreatedAt,
		Updated_at: newChirp.UpdatedAt,
		Body:       newChirp.Body,
		User_id:    newChirp.UserID,
	}

	respondWithJSON(w, 201, chirp)
}

var prohibited_list = []string{"kerfuffle", "sharbert", "fornax"}

func profanityCensor(text string) string {
	lcase_text := strings.ToLower(text)
	orig_slice := strings.Split(text, " ")
	lcase_text_slice := strings.Split(lcase_text, " ")
	for idx, word := range lcase_text_slice {
		if isProhibited(word) {
			orig_slice[idx] = "****"
		}
	}
	return strings.Join(orig_slice, " ")
}

func isProhibited(text string) bool {
	for _, word := range prohibited_list {
		if text == word {
			return true
		}
	}
	return false
}

func (c *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	var chirps []database.Chirp
	var err error
	asc := true
	authorQuery := r.URL.Query().Get("author_id")
	if authorQuery != "" {
		userId, parseErr := uuid.Parse(authorQuery)
		if parseErr != nil {
			respondWithError(w, 400, "invalid author_id")
			return
		}
		chirps, err = c.dbQueries.GetChirpsByUserID(r.Context(), userId)
	} else {
		chirps, err = c.dbQueries.GetChirps(r.Context())
	}
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	sortQuery := r.URL.Query().Get("sort")
	if sortQuery == "desc" {
		asc = false
	}
	export := make([]Chirp, len(chirps))
	for idx, chirp := range chirps {
		export_chirp := Chirp{
			Id:         chirp.ID,
			Created_at: chirp.CreatedAt,
			Updated_at: chirp.UpdatedAt,
			Body:       chirp.Body,
			User_id:    chirp.UserID,
		}
		export[idx] = export_chirp
	}
	if !asc {
		sort.Slice(export, func(i, j int) bool { return export[i].Created_at.After(export[j].Created_at) })
	}
	respondWithJSON(w, 200, export)
}

func (c *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	chirp, err := c.dbQueries.GetChirpByID(r.Context(), id)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	export := Chirp{
		Id:         chirp.ID,
		Created_at: chirp.CreatedAt,
		Updated_at: chirp.UpdatedAt,
		Body:       chirp.Body,
		User_id:    chirp.UserID,
	}
	respondWithJSON(w, 200, export)
}

func (c *apiConfig) handlerDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	userId, err := c.validateUser(r)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	chirp, err := c.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	if chirp.UserID != userId {
		respondWithError(w, 403, "unauthorized")
		return
	}
	err = c.dbQueries.DeleteChirpById(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	w.WriteHeader(204)
}
