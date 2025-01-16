package main

import (
	"log"
	"os"

	"codeberg.org/iklabib/kerat/server"
	"codeberg.org/iklabib/kerat/util"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config, err := util.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	processor, err := server.NewSubmissionProcessor(config)
	if err != nil {
		log.Fatal(err)
	}

	httpServer := server.NewHTTPServer(processor, config.QueueCap)

	address := ":31415"
	if host := os.Getenv("KERAT_HOST"); host != "" {
		address = host
	}

	log.Printf("Server starting on %s\n", address)
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))
	e.POST("/submit", httpServer.HandleSubmission)
	e.Logger.Fatal(e.Start(address))
}
