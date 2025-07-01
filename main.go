package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

// go build -o out && ./out  (this runs the server via termenal)

func main() {
	const filepathRoot = "."
	const port = "8080"
	apiCfg := apiConfig{}
	// 1. Create a new http.ServeMux
	// This is essentially a router or "multiplexer".
	// It's responsible for mapping incoming HTTP requests to their appropriate "handlers".
	// Think of it as the front desk of a magical hotel, directing guests (requests)
	// to the correct room (handler) based on their destination.
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /healthz", handlerReadiness)
	mux.HandleFunc("GET /metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /reset", apiCfg.metricsReset)

	// 2. Create a new http.Server struct.
	// This struct represents the actual HTTP server itself.
	// It holds configuration like the address it should listen on (`:8080`)
	// and the `Handler` which is the `ServeMux` we just created.
	// The `ServeMux` will then decide how to respond to requests.
	server := &http.Server{
		Addr:    ":" + port, // This tells the server to listen on port 8080 of your local machine
		Handler: mux,        // This assigns our `ServeMux` as the main request handler for the server
	}

	// 3. Use the server's ListenAndServe method to start the server.
	// This method makes the server start listening for incoming connections
	// on the address you specified (localhost:8080).
	// It's a blocking call, meaning it will run continuously until the program is stopped
	// or an error occurs.
	//
	// It returns an `error` if something goes wrong, like if the port is already in use.
	// We check for that error and use `log.Fatalf` to stop the program if it happens.
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// checks if the servers up
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1) // ✅ increment atomically
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	count := cfg.fileserverHits.Load()
	fmt.Fprintf(w, "Hits: %d", count)
}

func (cfg *apiConfig) metricsReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(w, "Hits have been reset")
}
