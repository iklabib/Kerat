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
	config, err := util.LoadGlobalConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// https://github.com/google/gvisor/issues/190
	// so we should fail it for now
	if config.Runtime == "runsc" {
		log.Fatalln("runsc is unsupported")
	}

	if config.QueueCap < 0 {
		log.Fatalln("queue cap must be greater than zero")
	}

	queue := make(chan string, config.QueueCap)
	submissionConfigs, err := util.LoadSubmissionConfigs("configs", config.Enables)

	engine := container.NewEngine(*config, submissionConfigs)
	if err != nil {
		log.Fatal(err)
	}

	if err := engine.Check(); err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatalf("load submission configs error: %s\n", err.Error())
	} else if len(submissionConfigs) == 0 {
		log.Fatalln("no config loaded")
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

		IsSupported := engine.IsSupported(submission.Type)
		if !IsSupported {
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
			result, err := engine.Run(ctx, containerName, &submission)
			if err != nil {
				log.Printf("[%s] %s\n", submissionId, ctx.Err())
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
