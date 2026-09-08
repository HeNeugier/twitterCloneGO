package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HeNeugier/twitterCloneGO/internal/auth"
	"github.com/HeNeugier/twitterCloneGO/internal/database"
)

type newUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) createNewUserHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	user := newUser{}
	err := decoder.Decode(&user)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request", err)
		return
	}

	hashedPw, err := auth.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error occurred when hashing the password.", err)
		return
	}

	dbUser, err := cfg.dbQuery.CreateUser(r.Context(), database.CreateUserParams{
		Email:          user.Email,
		HashedPassword: hashedPw,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error occurred when adding the user to the DB.", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})
}

func (cfg *apiConfig) clearDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Unauthorised action.", nil)
		return
	}

	err := cfg.dbQuery.ClearDatabase(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Problem clearing the DB", err)
		return
	}
	respondWithJSON(w, http.StatusOK, "DB Cleared")
}

func (cfg *apiConfig) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	const expiresIn = time.Duration(time.Hour * 1)
	const refreshExpiresIn = time.Duration(time.Hour * 24 * 60)

	decoder := json.NewDecoder(r.Body)
	user := newUser{}
	err := decoder.Decode(&user)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding request", err)
		return
	}

	dbUser, err := cfg.dbQuery.ReturnUserByEmail(r.Context(), user.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	ok, err := auth.CheckPasswordHash(user.Password, dbUser.HashedPassword)
	if err != nil || !ok {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	tok, err := auth.MakeJWT(
		dbUser.ID,
		cfg.secretString,
		time.Duration(expiresIn),
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error occurred while making an auth token", err)
		return
	}

	refreshTok := auth.MakeRefreshToken()
	dbRefreshToken, err := cfg.dbQuery.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshTok,
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().Add(refreshExpiresIn),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error occurred while adding the new refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        tok,
		RefreshToken: dbRefreshToken.Token,
	})
}

func (cfg *apiConfig) refreshUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	return
}
