package server

import (
	"context"
	"errors"
	"log"

	"codeberg.org/iklabib/kerat/model"
	"github.com/labstack/echo/v4"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var ALPHABET string = "abcdefghijklmnopqrstuvwxyz0123456789"

type HTTPServer struct {
	processor *SubmissionProcessor
	queue     chan string
}

func NewHTTPServer(processor *SubmissionProcessor, queueCap int) *HTTPServer {
	return &HTTPServer{
		processor: processor,
		queue:     make(chan string, queueCap),
	}
}

func (s *HTTPServer) HandleSubmission(c echo.Context) error {
	var submission model.Submission

	if err := c.Bind(&submission); err != nil {
		return c.JSON(400, "bad request")
	}

	submissionId, err := gonanoid.Generate(ALPHABET, 8)
	if err != nil {
		log.Printf("[error] failed to generate submission ID: %v\n", err)
		return c.JSON(500, "internal server error")
	}

	ctx := c.Request().Context()

	select {
	case s.queue <- submissionId:
		defer func() { <-s.queue }()

		result, err := s.processor.ProcessSubmission(ctx, submission, submissionId)
		if err != nil {
			log.Printf("[%s] processing error: %v\n", submissionId, err)
			return c.JSON(500, "internal server error")
		}

		return c.JSON(200, result)

	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.Canceled) {
			return c.JSON(499, "request canceled")
		} else {
			log.Printf("[%s] %s\n", submissionId, ctx.Err().Error())
			return c.JSON(500, "internal server error")
		}
	}
}
