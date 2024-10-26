package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"codeberg.org/iklabib/kerat/container"
	"codeberg.org/iklabib/kerat/model"
	"codeberg.org/iklabib/kerat/toolchains"
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

		tc, err := toolchains.NewToolchain(submission.Type)
		if err != nil {
			log.Printf("[%s] %s\n", submissionId, err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		source := model.SourceCode{
			SourceCodeTest: submission.SourceCodeTest,
			SourceCodes:    submission.SourceFiles,
		}

		workdir, err := os.MkdirTemp("", "box_")
		if err != nil {
			log.Printf("[%s] error to create temp dir: %s\n", submissionId, err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if err := tc.PreBuild(workdir, source); err != nil {
			log.Printf("[%s] prebuild error: %s\n", submissionId, err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// FIXME: make sure filenames are not full path
		var files []string
		for _, v := range source.SourceCodes {
			files = append(files, v.Filename)
		}

		build, err := tc.Build(workdir, files)
		if err != nil {
			log.Printf("[%s] build error: %s\n", submissionId, err.Error())
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if !build.Success {
			w.Header().Set("Content-Type", "application/json")
			ret := struct {
				Success bool   `json:"success"`
				Output  string `json:"output"`
			}{}
			json.NewEncoder(w).Encode(ret)
			return
		}

		// TODO: we could have internal error when running in seperate container, make the system error distinctive
		result, err := s.engine.Run(r.Context(), model.RunPayload{Type: submission.Type, Bin: build.Bin})
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
