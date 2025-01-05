package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"codeberg.org/iklabib/kerat/model"
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

func (s *HTTPServer) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	submission, submissionId, ok := s.decodeAndValidateSubmission(w, r)
	if !ok {
		return
	}

	select {
	case s.queue <- submissionId:
		defer func() { <-s.queue }()

		result, err := s.processor.ProcessSubmission(r.Context(), submission, submissionId)
		if err != nil {
			log.Printf("[%s] processing error: %v\n", submissionId, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(result)

	case <-r.Context().Done():
		s.handleContextCancellation(w, r, submissionId)
	}
}

func (s *HTTPServer) decodeAndValidateSubmission(w http.ResponseWriter, r *http.Request) (model.Submission, string, bool) {
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

	return submission, submissionId, true
}

func (s *HTTPServer) handleContextCancellation(w http.ResponseWriter, r *http.Request, submissionId string) {
	if errors.Is(r.Context().Err(), context.Canceled) {
		w.WriteHeader(499)
		w.Write([]byte("request canceled"))
	} else {
		log.Printf("[%s] %s\n", submissionId, r.Context().Err().Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
