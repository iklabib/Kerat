package main

import (
	"log"
	"net/http"

	"codeberg.org/iklabib/kerat/container"
	"codeberg.org/iklabib/kerat/model"
	"codeberg.org/iklabib/kerat/util"
	"github.com/labstack/echo/v4"
)

func main() {
	config, err := util.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// https://github.com/google/gvisor/issues/190
	// so we should fail it for now
	if config.Runtime == "runsc" {
		log.Fatalln("runsc is unsupported")
	}

	engine := container.NewEngine(*config)
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

		if submission.Id == "" {
			submission.Id = util.RandomString()
		}

		result, err := engine.Run(&submission)
		if err != nil {
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, "internal error")
		}

		return c.JSON(http.StatusOK, result)
	})

	e.Logger.Fatal(e.Start(":31415"))
}
