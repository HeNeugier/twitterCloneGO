package main

import (
	"encoding/json"
	"db"
	"net/http"
)

func (cfg *apiConfig) createNewUserHandler(w http.ResponseWriter, r *http.Request) {
	type newUser struct{
		Email		string
	}
	
	decoder := json.NewDecoder(r.Body)
	user := newUser{}
	err := decoder.Decode(&user)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request", err)
		return
	}

	dbUser, err := cfg.dbQuery.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error occurred when adding the user to the DB.", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:					dbUser.ID,
		CreatedAt:	dbUser.CreatedAt,
		UpdatedAt:	dbUser.UpdatedAt,
		Email:			dbUser.Email,
	})
}

func (cfg *apiConfig) clearDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Unauthorised action.", nil)
		return
	}

	_, err := cfg.dbQuery.ClearDatabase(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Problem clearing the DB", err)
		return
	}
	respondWithJSON(w, http.StatusAccepted, "DB Cleared")
}
