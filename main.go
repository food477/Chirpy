package main

import (
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
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	// 1. Create a new http.ServeMux
	// This is essentially a router or "multiplexer".
	// It's responsible for mapping incoming HTTP requests to their appropriate "handlers".
	// Think of it as the front desk of a magical hotel, directing guests (requests)
	// to the correct room (handler) based on their destination.
	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	// 2. Create a new http.Server struct.
	// This struct represents the actual HTTP server itself.
	// It holds configuration like the address it should listen on (`:8080`)
	// and the `Handler` which is the `ServeMux` we just created.
	// The `ServeMux` will then decide how to respond to requests.
	srv := &http.Server{
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
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
