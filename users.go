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
	Red          bool      `json:"is_chirpy_red"`
}

func (c *apiConfig) validateUser(r *http.Request) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.Nil, err
	}
	userID, err := auth.ValidateJWT(token, c.secret)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
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
	dbUser, err := c.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Red:       dbUser.IsChirpyRed,
	}
	respondWithJSON(w, 201, user)
}

func (c *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := c.validateUser(r)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	type ChangeRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	var req ChangeRequest
	err = decoder.Decode(&req)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	hashedPass, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	dbUser, err := c.dbQueries.UpdateUserMailAndPassword(r.Context(), database.UpdateUserMailAndPasswordParams{ID: userID, Email: req.Email, HashedPassword: hashedPass})
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 200, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Red:       dbUser.IsChirpyRed,
	})
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
	dbUser, err := c.dbQueries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	match, err := auth.CheckPasswordHash(req.Password, dbUser.HashedPassword)
	if err != nil || !match {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	access_token, err := auth.MakeJWT(dbUser.ID, c.secret, time.Duration(expiriesInAccess)*time.Second)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	refresh_token_params := database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().UTC().Add(time.Duration(expiriesInRefresh) * time.Hour),
	}

	refresh_token, err := c.dbQueries.CreateRefreshToken(r.Context(), refresh_token_params)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	user := User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		AccessToken:  access_token,
		RefreshToken: refresh_token.Token,
		Red:          dbUser.IsChirpyRed,
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
