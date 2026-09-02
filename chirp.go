package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	
	"github.com/google/uuid"
	"github.com/HeNeugier/twitterCloneGO/internal/database"
)

type Chirp struct {
	ID					uuid.UUID		`json:"id"`
	CreatedAt		time.Time		`json:"created_at"`
	UpdatedAt		time.Time		`json:"updated_at"`
	Body 				string 			`json:"body"`
	UserID			uuid.UUID		`json:"user_id"`
}

func (cfg *apiConfig) postValidChirpHandler(w http.ResponseWriter, r *http.Request) {
	//-- Define JSON structures we expect to see
	type parameters struct {
		Body 		string 			`json:"body"`
		UserID 	uuid.UUID		`json:"user_id"`
	}

	//-- Decode into our parameters struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
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
			Body:		cleaned_text, 
			UserID:	params.UserID,
		})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error occurred when adding the chirp to the DB.", err)
		return
	}

	//-- Case of success
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID: 				dbChirp.ID,
		CreatedAt:	dbChirp.CreatedAt,
		UpdatedAt:	dbChirp.UpdatedAt,
		Body:				dbChirp.Body,
		UserID:			dbChirp.UserID,
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
		return "", errors.New("Chirp is too long.")
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
