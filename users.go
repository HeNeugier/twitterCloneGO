package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/HeNeugier/twitterCloneGO/internal/auth"
	"github.com/HeNeugier/twitterCloneGO/internal/database"
	"github.com/google/uuid"
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
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
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

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

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
		ExpiresAt: time.Now().UTC().Add(refreshExpiresIn),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error occurred while adding the new refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:          dbUser.ID,
			CreatedAt:   dbUser.CreatedAt,
			UpdatedAt:   dbUser.UpdatedAt,
			Email:       dbUser.Email,
			IsChirpyRed: dbUser.IsChirpyRed,
		},
		Token:        tok,
		RefreshToken: dbRefreshToken.Token,
	})
}

func (cfg *apiConfig) refreshUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	type returnForm struct {
		Token string `json:"token"`
	}

	// Check our header format
	refreshTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "bad request", err)
		return
	}

	// Lookup the token in the DB
	dbUser, err := cfg.dbQuery.GetUserFromRefreshToken(r.Context(), refreshTok)
	if err != nil || dbUser == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorised access", err)
		return
	}

	// Create a new access token from the returned user and known secret
	newAccessTok, err := auth.MakeJWT(dbUser, cfg.secretString, time.Duration(time.Hour*1))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate provided token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, returnForm{
		Token: newAccessTok,
	})
}

func (cfg *apiConfig) revokeUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Check our header format
	refreshTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Malformed request", err)
		return
	}

	err = cfg.dbQuery.RevokeGivenRefreshToken(r.Context(), refreshTok)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error while revoking", err)
		return
	}

	// Successful return
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) updateUserCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	// NOTE in this case we expect the ACCESS token. NOT the REFRESH
	accessTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorised access", err)
		return
	}

	userID, err := auth.ValidateJWT(accessTok, cfg.secretString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorised access", err)
		return
	}

	// Extract our new email and passwords
	decoder := json.NewDecoder(r.Body)
	user := newUser{}
	err = decoder.Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding request", err)
		return
	}

	// Hash the new password
	hashedPw, err := auth.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Problem hashing password", err)
		return
	}

	updatedUser, err := cfg.dbQuery.UpdateEmailAndHashedPassword(r.Context(), database.UpdateEmailAndHashedPasswordParams{
		ID:             userID,
		HashedPassword: hashedPw,
		Email:          user.Email,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update credentials", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	})
}

func (cfg *apiConfig) userUpgradeHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	// Check the API key
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || cfg.polkaKey != apiKey {
		respondWithError(w, http.StatusUnauthorized, "API key is incorrect", err)
	}

	// Extract our request
	decoder := json.NewDecoder(r.Body)
	req := request{}
	err = decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding request", err)
		return
	}
	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Parse the UUID
	userID, err := uuid.Parse(req.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding request", err)
		return
	}

	_, err = cfg.dbQuery.UpgradeUserToChirpyRed(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "User not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not update user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
