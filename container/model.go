package container

import (
	"io"

	"codeberg.org/iklabib/kerat/model"
)

type RunPayload struct {
	ContainerId    string
	SubmissionType string
}

type CopyPayload struct {
	ContainerId string
	Dest        string
	Content     io.Reader // model.SourceCode as TAR
}

type ContainerResult struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Output  []model.TestResult `json:"output"`
}
