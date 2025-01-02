package server

import "codeberg.org/iklabib/kerat/container"

type SubmitResult struct {
	Success bool                   `json:"success"`
	Build   string                 `json:"build"`
	Tests   []container.TestResult `json:"tests"`
	Metrics container.Metrics      `json:"metrics"`
}
