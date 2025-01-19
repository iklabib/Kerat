package main

import (
	"log"
	"net/http"
	"os"

	"codeberg.org/iklabib/kerat/processor"
	"codeberg.org/iklabib/kerat/server"
)

func main() {
	config, err := processor.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	processor, err := processor.NewSubmissionProcessor(config)
	if err != nil {
		log.Fatal(err)
	}

	httpServer := server.NewHTTPServer(processor, config.QueueCap)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /submit", httpServer.HandleSubmission)

	address := ":31415"
	if host := os.Getenv("KERAT_HOST"); host != "" {
		address = host
	}

	log.Printf("Server starting on %s\n", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatal(err)
	}
}
