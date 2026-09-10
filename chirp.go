package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/HeNeugier/twitterCloneGO/internal/auth"
	"github.com/HeNeugier/twitterCloneGO/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) postValidChirpHandler(w http.ResponseWriter, r *http.Request) {
	//-- Define JSON structures we expect to see
	type parameters struct {
		Body string `json:"body"`
	}

	// Validate the user's login status using the JWT
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Login first", err)
		return
	}
	bearerUUID, err := auth.ValidateJWT(bearerToken, cfg.secretString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Login first", err)
		return
	}

	//-- Decode into our parameters struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding chirp", err)
		return
	}

	cleaned_text, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	dbChirp, err := cfg.dbQuery.CreateChirp(
		r.Context(), database.CreateChirpParams{
			Body:   cleaned_text,
			UserID: bearerUUID,
		})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error occurred when adding the chirp to the DB.", err)
		return
	}

	//-- Case of success
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func validateChirp(body string) (string, error) {
	//-- CONSTANTS --
	const maxChirpLength = 140
	var profanity = []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	if len(body) > maxChirpLength {
		return "", errors.New("chirp is too long")
	}

	return filterProfanity(body, profanity), nil
}

func filterProfanity(message string, profanity []string) string {
	words := strings.Fields(message)

	for i, word := range words {
		for _, badWord := range profanity {
			if strings.ToLower(word) == badWord {
				words[i] = "****"
				break
			}
		}
	}
	return strings.Join(words, " ")
}

func (cfg *apiConfig) retrieveAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	// Setup our variables
	var dbChirps []database.Chirp
	var err error

	// If we have the optional 'author_id'
	authorID := r.URL.Query().Get("author_id")
	if authorID != "" {
		uuidID, err := uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Malformed author_id", err)
			return
		}

		dbChirps, err = cfg.dbQuery.GetAllUserChirps(r.Context(), uuidID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondWithError(w, http.StatusNotFound, "User not found", err)
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Database error while retrieving user chirps", err)
			return
		}
	} else {
		dbChirps, err = cfg.dbQuery.RetrieveChirps(r.Context())
		if err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"An error occurred when retrieving the chirps.",
				err,
			)
			return
		}
	}

	//-- Cast our dbChirp type to our expected form Chirp
	chirps := make([]Chirp, 0, len(dbChirps))
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) retrieveChirp(w http.ResponseWriter, r *http.Request) {
	//-- Check our PathValue
	parsedUUID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		// Handle the error (for example, return a 404 or bad request status)
		respondWithError(w, http.StatusBadRequest, "The provider uuid is not valid", err)
		return
	}
	dbChirp, err := cfg.dbQuery.RetrieveChirp(r.Context(), parsedUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "The chirp was not found.", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	accessTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorised access", err)
		return
	}

	// Get the user ID of the request user
	userID, err := auth.ValidateJWT(accessTok, cfg.secretString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorised access", err)
		return
	}

	parsedUUID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		// Handle the error (for example, return a 404 or bad request status)
		respondWithError(w, http.StatusBadRequest, "The provided uuid is not valid a valid format", err)
		return
	}
	dbChirp, err := cfg.dbQuery.RetrieveChirp(r.Context(), parsedUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "The chirp was not found.", err)
		return
	}

	// Check the requesting User ID matches that of the chirp retrieved
	if userID != dbChirp.UserID {
		respondWithError(w, http.StatusForbidden, "This is not your chirp", err)
		return
	}

	// All checks pass, now delete
	deletedChirp, err := cfg.dbQuery.DeleteChirp(r.Context(), dbChirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "There was a problem deleting the chirp", err)
		return
	} else if deletedChirp != dbChirp.ID {
		respondWithError(w,
			http.StatusInternalServerError,
			fmt.Sprintf("Chirp deleted was: %s, wanted: %s", deletedChirp, dbChirp.ID),
			err,
		)
		return
	}

	// Success case
	w.WriteHeader(http.StatusNoContent)
}
