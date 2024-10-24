package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"codeberg.org/iklabib/kerat/container"
	"codeberg.org/iklabib/kerat/model"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var ALPHABET string = "_abcdefghijklmnopqrstuvwxyz0123456789"

type Server struct {
	engine *container.Engine
	queue  chan string
}

func NewServer(engine *container.Engine, queueCap int) *Server {
	s := &Server{
		engine: engine,
		queue:  make(chan string, queueCap),
	}
	return s
}

func (s *Server) HandleRun(w http.ResponseWriter, r *http.Request) {
	var submission model.Submission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	submissionId, err := gonanoid.Generate(ALPHABET, 8)
	if err != nil {
		log.Printf("[error] failed to generate submission ID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if !s.engine.IsSupported(submission.Type) {
		msg := fmt.Sprintf("bad request: submission type \"%s\" is unsupported", submission.Type)
		log.Printf("[%s] %s\n", submissionId, msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	select {
	case s.queue <- submissionId:
		defer func() {
			<-s.queue
		}()

		result, err := s.engine.Run(r.Context(), &submission)
		if err != nil {
			log.Printf("[%s] %s\n", submissionId, err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return

	case <-r.Context().Done():
		if errors.Is(r.Context().Err(), context.Canceled) {
			w.WriteHeader(499)
			w.Write([]byte("request canceled"))
		} else {
			log.Printf("[%s] %s\n", submissionId, r.Context().Err().Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
}
