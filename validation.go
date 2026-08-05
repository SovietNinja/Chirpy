package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

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
		Cleaned_body string `json:"cleaned_body"`
	}

	resp := validResponse{
		Cleaned_body: profanityCensor(chirp.Body),
	}
	respondWithJSON(w, 200, resp)
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
