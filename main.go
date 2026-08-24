package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	//-- Define Constants --
	const filepathRoot = "."
	const port = "8080"

	//-- Create our traffic director (mux: multiplexer) and config struct --
	myMux := http.NewServeMux()
	apiCfg := &apiConfig{}

	//-- Muxxing Handlers start here --
	//---------------------------------
	myMux.HandleFunc("/healthz", readinessHandler)
	//- middleware muxxing
	myMux.HandleFunc("/metrics", apiCfg.metricsHandler)
	myMux.HandleFunc("/reset", apiCfg.metricsResetHandler)
	
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

