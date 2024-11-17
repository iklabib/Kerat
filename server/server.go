package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"codeberg.org/iklabib/kerat/container"
	"codeberg.org/iklabib/kerat/memo"
	"codeberg.org/iklabib/kerat/model"
	"codeberg.org/iklabib/kerat/toolchains"
	"codeberg.org/iklabib/kerat/util"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// TODO
// log last accessed exercise and clear if there are no activity for N minutes

var ALPHABET string = "abcdefghijklmnopqrstuvwxyz0123456789"

type Server struct {
	engine *container.Engine
	config *model.Config
	queue  chan string
}

func NewServer(configPath string) (*Server, error) {
	config, err := util.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	if config.QueueCap < 0 {
		return nil, fmt.Errorf("queue cap must be greater than zero")
	}

	engine, err := container.NewEngine(*config)
	if err != nil {
		return nil, err
	}

	if err := engine.Check(); err != nil {
		return nil, err
	}

	s := &Server{
		config: config,
		engine: engine,
		queue:  make(chan string, config.QueueCap),
	}
	return s, nil
}

func (s *Server) HandleRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	caches := memo.NewBoxCaches(s.config.CleanInterval)

	select {
	case s.queue <- submissionId:
		defer func() {
			<-s.queue
		}()

		exerciseId := submission.ExerciseId
		tc, ok := caches.LoadToolchain(exerciseId)
		if !ok {
			var err error
			tc, err = toolchains.NewToolchain(submission, s.config.Repository)
			if err != nil {
				log.Printf("[%s] failed to create box %s\n", submissionId, err.Error())
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			caches.AddToolchain(exerciseId, tc)
		}

		if err := tc.Prep(); err != nil {
			log.Printf("[%s] build error: %s\n", submissionId, err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		build, err := tc.Build()
		if err != nil {
			log.Printf("[%s] build error: %s\n", submissionId, err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if !build.Success {
			ret := model.EvalResult{Build: build.Stderr}
			json.NewEncoder(w).Encode(ret)
			return
		}

		// TODO: we could have internal error when running in seperate container, make the system error distinctive
		ret, err := s.engine.Run(r.Context(), model.RunPayload{Type: submission.Type, Bin: build.Bin})
		if err != nil {
			log.Printf("[%s] %s\n", submissionId, err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if !ret.Success {
			log.Printf("[%s] %s\n", submissionId, ret.Message)
		}

		json.NewEncoder(w).Encode(ret)
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
