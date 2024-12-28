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
	"codeberg.org/iklabib/kerat/memo"
	"codeberg.org/iklabib/kerat/model"
	"codeberg.org/iklabib/kerat/toolchains"
	"codeberg.org/iklabib/kerat/util"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

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

	if config.QueueCap <= 0 {
		return nil, fmt.Errorf("queue cap must be greater than zero")
	}

	engine, err := container.NewEngine(*config)
	if err != nil {
		return nil, err
	}

	if err := engine.Check(); err != nil {
		return nil, err
	}

	return &Server{
		config: config,
		engine: engine,
		queue:  make(chan string, config.QueueCap),
	}, nil
}

func (s *Server) decodeAndValidateSubmission(w http.ResponseWriter, r *http.Request) (model.Submission, string, bool) {
	var submission model.Submission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return submission, "", false
	}

	submissionId, err := gonanoid.Generate(ALPHABET, 8)
	if err != nil {
		log.Printf("[error] failed to generate submission ID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return submission, "", false
	}

	if !s.engine.IsSupported(submission.Type) {
		msg := fmt.Sprintf("bad request: submission type \"%s\" is unsupported", submission.Type)
		log.Printf("[%s] %s\n", submissionId, msg)
		http.Error(w, msg, http.StatusBadRequest)
		return submission, submissionId, false
	}

	return submission, submissionId, true
}

func (s *Server) handleContextCancellation(w http.ResponseWriter, r *http.Request, submissionId string) {
	if errors.Is(r.Context().Err(), context.Canceled) {
		w.WriteHeader(499)
		w.Write([]byte("request canceled"))
	} else {
		log.Printf("[%s] %s\n", submissionId, r.Context().Err().Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) handleInterpretedSubmission(w http.ResponseWriter, r *http.Request, submission model.Submission, submissionId string) {
	containerId, err := s.engine.Create(context.Background(), submission.Type)
	if err != nil {
		log.Printf("[%s] container creation error: %v\n", submissionId, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	content, err := util.TarSources(submission.Source)
	if err != nil {
		log.Printf("[%s] creating tar error: %v\n", submissionId, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	copyPayload := container.CopyPayload{ContainerId: containerId, Dest: "/workspace", Content: &content}
	err = s.engine.Copy(context.Background(), copyPayload)
	if err != nil {
		log.Printf("[%s] copying tar error: %v\n", submissionId, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	ret, err := s.engine.Run(r.Context(), container.RunPayload{ContainerId: containerId})
	if err != nil {
		log.Printf("[%s] %s\n", submissionId, err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(ret)
}

func (s *Server) handleCompiledSubmission(w http.ResponseWriter, r *http.Request, submission model.Submission, submissionId string, caches *memo.BoxCaches) {
	exerciseId := submission.ExerciseId

	tc, ok := caches.LoadToolchain(exerciseId)
	if !ok {
		var err error
		tc, err = toolchains.NewToolchain(submission, s.config.Repository)
		if err != nil {
			log.Printf("[%s] failed to create toolchain: %v\n", submissionId, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		caches.AddToolchain(exerciseId, tc)
	}

	if err := tc.Prep(); err != nil {
		log.Printf("[%s] prep error: %v\n", submissionId, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	build, err := tc.Build()
	if err != nil {
		log.Printf("[%s] build error: %v\n", submissionId, err)
		http.Error(w, "build failed", http.StatusInternalServerError)
		return
	}

	if !build.Success {
		log.Printf("[%s] build stderr: %s \n", submissionId, build.Stderr)
		log.Printf("[%s] build stdout: %s \n", submissionId, build.Stdout)
		result := model.SubmitResult{
			Build: string(build.Stderr),
		}
		json.NewEncoder(w).Encode(result)
		return
	}

	bin, err := os.ReadFile(build.BinPath)
	if err != nil {
		log.Printf("[%s] failed to read binary: %v\n", submissionId, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	containerId, err := s.engine.Create(context.Background(), submission.Type)
	if err != nil {
		log.Printf("[%s] container creation error: %v\n", submissionId, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	content, err := util.TarBinary("box", bin)
	if err != nil {
		log.Printf("[%s] creating tar error: %v\n", submissionId, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	copyPayload := container.CopyPayload{ContainerId: containerId, Dest: "/workspace", Content: &content}
	err = s.engine.Copy(context.Background(), copyPayload)
	if err != nil {
		log.Printf("[%s] copying tar error: %v\n", submissionId, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	runPayload := container.RunPayload{ContainerId: containerId}
	ret, err := s.engine.Run(r.Context(), runPayload)
	if err != nil {
		log.Printf("[%s] runtime error: %v\n", submissionId, err)
		http.Error(w, "runtime error", http.StatusInternalServerError)
		return
	}

	result := model.SubmitResult{
		Success: ret.Success,
		Tests:   ret.Output,
	}
	json.NewEncoder(w).Encode(result)
}

func (s *Server) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	submission, submissionId, ok := s.decodeAndValidateSubmission(w, r)
	if !ok {
		return
	}

	select {
	case s.queue <- submissionId:
		defer func() { <-s.queue }()

		if submission.Type == "python" {
			s.handleInterpretedSubmission(w, r, submission, submissionId)
		} else {
			caches := memo.NewBoxCaches(s.config.CleanInterval)
			s.handleCompiledSubmission(w, r, submission, submissionId, &caches)
		}

	case <-r.Context().Done():
		s.handleContextCancellation(w, r, submissionId)
	}
}
