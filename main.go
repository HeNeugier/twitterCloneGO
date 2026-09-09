package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/HeNeugier/twitterCloneGO/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQuery        *database.Queries
	platform       string
	secretString   string
}

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func main() {
	//-- Import .env into values in code --
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL .env variable must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM .env variable must be set")
	}
	secretString := os.Getenv("SECRET_STRING")
	if secretString == "" {
		log.Fatal("SECRET_STRING .env variable must be set")
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
		fileserverHits: atomic.Int32{},
		dbQuery:        dbQueries,
		platform:       platform,
		secretString:   secretString,
	}

	//-- Muxxing Handlers start here --
	//---------------------------------
	myMux.HandleFunc("GET /api/healthz", readinessHandler)

	myMux.HandleFunc("POST /api/users", apiCfg.createNewUserHandler)
	myMux.HandleFunc("PUT /api/users", apiCfg.updateUserCredentialsHandler)

	myMux.HandleFunc("POST /api/refresh", apiCfg.refreshUserTokenHandler)
	myMux.HandleFunc("POST /api/revoke", apiCfg.revokeUserTokenHandler)
	myMux.HandleFunc("POST /api/login", apiCfg.loginUserHandler)

	myMux.HandleFunc("POST /api/polka/webhooks", apiCfg.userUpgradeHandler)

	myMux.HandleFunc("GET /api/chirps", apiCfg.retrieveAllChirpsHandler)
	myMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.retrieveChirp)
	myMux.HandleFunc("POST /api/chirps", apiCfg.postValidChirpHandler)
	myMux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirpHandler)

	myMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	myMux.HandleFunc("POST /admin/reset", apiCfg.clearDatabaseHandler)

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
		Addr:    ":" + port,
		Handler: myMux,
	}

	//-- Begin listening for HTTP --
	server.ListenAndServe()

}

func readinessHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
