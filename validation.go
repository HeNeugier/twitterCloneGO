package main

import (
	"encoding/json"
	"net/http"
)


func validateChirpsHandler(w http.ResponseWriter, r *http.Request) {
	//--CONSTS
	const maxChirpLength = 140

	//-- Define what JSON structure we expect to see
	type chirps struct {
		Body string `json:"body"`
	}
	type validReturn struct {
		Valid bool `json:"valid"`
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
		Valid: true,
	})
}

