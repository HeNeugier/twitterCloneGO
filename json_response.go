package main

import (
	"encoding/json"
	"log"
	"net/http"
)


func respondWithError(w http.ResponseWriter, responseCode int, msg string, err error) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	if err != nil {
		log.Println(err)
	}
	if responseCode >= 500 {
		log.Printf("5XX error type: %s", msg)
	}

	respondWithJSON(w, responseCode, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, responseCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(responseCode)
	w.Write(dat)
}
