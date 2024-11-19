package main

import (
	"log"
	"net/http"
	"os"

	"codeberg.org/iklabib/kerat/server"
)

func main() {
	server, err := server.NewServer("config.yaml")
	if err != nil {
		log.Fatalln(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /submit", server.HandleSubmission)

	address := ":31415"
	if host := os.Getenv("KERAT_HOST"); host != "" {
		address = host
	}

	log.Printf("Server starting on %s\n", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatal(err)
	}
}
