package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SovietNinja/Chirpy/internal/auth"
	"github.com/google/uuid"
)

type UserUpgradeRequest struct {
	Event string      `json:"event"`
	Data  RequestData `json:"data"`
}

type RequestData struct {
	UserID uuid.UUID `json:"user_id"`
}

func (c *apiConfig) handlePolka(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != c.polkaKey {
		w.WriteHeader(401)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var req UserUpgradeRequest
	err = decoder.Decode(&req)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	if req.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	_, err = c.dbQueries.SetUserRed(r.Context(), req.Data.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, 404, "user not found")
		return
	} else if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	w.WriteHeader(204)
}
