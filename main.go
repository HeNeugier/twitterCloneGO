package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/HeNeugier/twitterCloneGO/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQuery *database.Queries
}

func main() {
	//-- Import .env into values in code --
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL .env variable must be set")
	}

	//-- Open a connection to our DB --
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Database failed to open: %s", err)
		return
	}
	dbQueries := database.New(db)

	//-- Define Constants --
	const filepathRoot = "."
	const port = "8080"

	//-- Create our traffic director (mux: multiplexer) and config struct --
	myMux := http.NewServeMux()
	apiCfg := &apiConfig{
		fileserverHits: 	atomic.Int32{},
		dbQuery: 					dbQueries,
	}

	//-- Muxxing Handlers start here --
	//---------------------------------
	myMux.HandleFunc("GET /api/healthz", readinessHandler)
	myMux.HandleFunc("POST /api/validate_chirp", validateChirpsHandler)

	myMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	myMux.HandleFunc("POST /admin/reset", apiCfg.metricsResetHandler)
	
	//- Helpers
	fileServer := http.FileServer(http.Dir(filepathRoot))
	appHandler := http.StripPrefix("/app", fileServer)

	//- wraps with middleware for metrics
	myMux.Handle(
		"/app/", 
		apiCfg.metricsIncMiddleware(appHandler),
	)

	//-----Mux-ends-here---------------

	//-- Define our server using the struct --
	server := &http.Server{
		Addr:				":" + port,
		Handler:		myMux,
	}

	//-- Begin listening for HTTP --
	server.ListenAndServe()

}

func readinessHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

