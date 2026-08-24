package main

import (
	"net/http"
)

func readinessHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	myMux := http.NewServeMux()
	myMux.HandleFunc("/healthz", readinessHandler)
	myMux.Handle(
		"/app/", 
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	)

	s := &http.Server{
		Addr:				":8080",
		Handler:		myMux,
	}

	s.ListenAndServe()

}
