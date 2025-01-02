package container

import "io"

type RunPayload struct {
	ContainerId    string
	SubmissionType string
}

type CopyPayload struct {
	ContainerId string
	Dest        string
	Content     io.Reader // SourceCode as TAR
}

type TestResult struct {
	Passed     bool   `json:"passed"`
	Name       string `json:"name"`
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace"`
}

type ContainerResult struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Output  []TestResult `json:"output"`
	Metrics `json:"metrics"`
}

type Metrics struct {
	ExitCode int     `json:"exit_code"`
	WallTime float64 `json:"wall_time"` // running wall time (s)
	CpuTime  uint64  `json:"cpu_time"`  // total CPU time consumed (ns)
	Memory   uint64  `json:"memory"`    // peak memory recorded (bytes)
}
