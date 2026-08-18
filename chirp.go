package main

import (
	"encoding/json"
	"net/http"
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
	decoder := json.NewDecoder(r.Body)
	chirp := Chirp{}
	err := decoder.Decode(&chirp)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	newChirpParams := database.CreateChirpParams{
		Body:   profanityCensor(chirp.Body),
		UserID: chirp.User_id,
	}

	newChirp, err := c.dbQueries.CreateChirp(r.Context(), newChirpParams)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	chirp = Chirp{
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
