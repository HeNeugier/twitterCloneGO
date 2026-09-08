package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HeNeugier/twitterCloneGO/internal/auth"
	"github.com/HeNeugier/twitterCloneGO/internal/database"
)

type newUser struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
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
	decoder := json.NewDecoder(r.Body)
	user := newUser{}
	err := decoder.Decode(&user)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding request", err)
		return
	}

	// Set our default timeout of 1 hour
	if user.ExpiresInSeconds == 0 || user.ExpiresInSeconds > 3600 {
		user.ExpiresInSeconds = 3600
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
		time.Duration(user.ExpiresInSeconds*int(time.Second)),
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error occurred while making a JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     tok,
	})
}
