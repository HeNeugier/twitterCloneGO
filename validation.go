package main

import (
	"encoding/json"
	"net/http"
	"strings"
)


func validateChirpsHandler(w http.ResponseWriter, r *http.Request) {
	//--CONSTS
	const maxChirpLength = 140
	var profanity = []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	//-- Define what JSON structure we expect to see
	type chirps struct {
		Body string `json:"body"`
	}
	type validReturn struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := chirps{}
	err := decoder.Decode(&chirp)

	//-- Error checks --
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding chirp", err)
		return
	}
	if len(chirp.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", err)
		return
	}

	//-- Case of success
	respondWithJSON(w, http.StatusOK, validReturn{
		CleanedBody: filterProfanity(chirp.Body, profanity),
	})
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

