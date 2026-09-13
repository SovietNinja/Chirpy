package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SovietNinja/Chirpy/internal/auth"
	"github.com/SovietNinja/Chirpy/internal/database"
	"github.com/google/uuid"
)

var expiriesInAccess = 3600     //access token lifespan, seconds
var expiriesInRefresh = 60 * 24 //refresh token lifespan, hours

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (c *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type createUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	var req createUserRequest
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	dbuser, err := c.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
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

func (c *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	var req LoginRequest
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	dbuser, err := c.dbQueries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	match, err := auth.CheckPasswordHash(req.Password, dbuser.HashedPassword)
	if err != nil || !match {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	access_token, err := auth.MakeJWT(dbuser.ID, c.secret, time.Duration(expiriesInAccess)*time.Second)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	refresh_token_params := database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    dbuser.ID,
		ExpiresAt: time.Now().UTC().Add(time.Duration(expiriesInRefresh) * time.Hour),
	}

	refresh_token, err := c.dbQueries.CreateRefreshToken(r.Context(), refresh_token_params)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	user := User{
		ID:           dbuser.ID,
		CreatedAt:    dbuser.CreatedAt,
		UpdatedAt:    dbuser.UpdatedAt,
		Email:        dbuser.Email,
		AccessToken:  access_token,
		RefreshToken: refresh_token.Token,
	}
	respondWithJSON(w, 200, user)
}

func (c *apiConfig) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	dbToken, err := c.dbQueries.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 401, "token not found")
		return
	}
	if dbToken.RevokedAt.Valid || time.Now().UTC().After(dbToken.ExpiresAt) {
		respondWithError(w, 401, "token is revoked or expired")
		return
	}
	newAccessToken, err := auth.MakeJWT(dbToken.UserID, c.secret, time.Duration(expiriesInAccess)*time.Second)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	type ReturnToken struct {
		Token string `json:"token"`
	}
	respondWithJSON(w, 200, ReturnToken{newAccessToken})
}

func (c *apiConfig) handleTokenRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	err = c.dbQueries.RevokeToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	w.WriteHeader(204)
}
