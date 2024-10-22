package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"codeberg.org/iklabib/kerat/container"
	"codeberg.org/iklabib/kerat/model"
	"codeberg.org/iklabib/kerat/util"
	"github.com/labstack/echo/v4"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/net/context"
)

func main() {
	config, err := util.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if config.QueueCap < 0 {
		log.Fatalln("queue cap must be greater than zero")
	}

	queue := make(chan string, config.QueueCap)

	engine, err := container.NewEngine(*config)
	if err != nil {
		log.Fatal(err)
	}

	if err := engine.Check(); err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.POST("/run", func(c echo.Context) error {
		var submission model.Submission

		if err := c.Bind(&submission); err != nil {
			return c.JSON(http.StatusBadRequest, "bad request")
		}

		submissionId, err := gonanoid.New(8)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "internal server error")
		}

		if !engine.IsSupported(submission.Type) {
			msg := fmt.Sprintf("bad request: submission type \"%s\" is unsupported", submission.Type)
			log.Printf("[%s] %s\n", submissionId, msg)
			return c.JSON(http.StatusBadRequest, msg)
		}

		ctx := c.Request().Context()

		select {
		case queue <- submissionId:
			defer func() {
				<-queue
			}()

			containerName := "kerat_" + submissionId
			result, err := engine.Run(ctx, &submission)
			if err != nil {
				log.Printf("[%s] %s\n", submissionId, err.Error())
				return c.JSON(http.StatusInternalServerError, "internal server error")
			}

			// TODO: don't remove container that has compile error. We want to reuse them
			if err := engine.Remove(containerName); err != nil {
				log.Printf("[%s] %s\n", submissionId, err.Error())
				return c.JSON(http.StatusInternalServerError, "internal server error")
			}

			return c.JSON(http.StatusOK, result)

		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return c.JSON(499, "request canceled")
			} else { // ideally this should not happened
				log.Printf("[%s] %s\n", submissionId, ctx.Err().Error())
				return c.JSON(http.StatusInternalServerError, "internal server error")
			}
		}
	})

	address := ":31415"
	if host := os.Getenv("KERAT_HOST"); host != "" {
		address = host
	}

	e.Logger.Fatal(e.Start(address))
}
