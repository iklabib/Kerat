package main

import (
	"log"
	"net/http"
	"os"

	"codeberg.org/iklabib/kerat/container"
	"codeberg.org/iklabib/kerat/server"
	"codeberg.org/iklabib/kerat/util"
)

func main() {
	config, err := util.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if config.QueueCap < 0 {
		log.Fatalln("queue cap must be greater than zero")
	}

	engine, err := container.NewEngine(*config)
	if err != nil {
		log.Fatal(err)
	}

	if err := engine.Check(); err != nil {
		log.Fatal(err)
	}

	server := server.NewServer(engine, config.QueueCap)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /run", server.HandleRun)

	address := ":31415"
	if host := os.Getenv("KERAT_HOST"); host != "" {
		address = host
	}

	log.Printf("Server starting on %s\n", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatal(err)
	}
}
