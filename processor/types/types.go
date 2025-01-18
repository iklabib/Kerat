package types

import "io"

type SubmissionConfig struct {
	Id             string           `json:"id" yaml:"id"`
	CPUPeriod      int64            `json:"cpu_period" yaml:"cpu_period"`
	CPUQuota       int64            `json:"cpu_quota" yaml:"cpu_quota"`
	MaxPids        int64            `json:"max_pids" yaml:"max_pids"`
	MaxSwap        int64            `json:"max_swap" yaml:"max_swap"`     // MiB
	MaxMemory      int64            `json:"max_memory" yaml:"max_memory"` // MiB
	Timeout        int              `json:"timeout" yaml:"timeout"`       // wall-time in seconds
	Ulimits        map[string]int64 `json:"ulimits" yaml:"ulimits"`
	ContainerImage string           `json:"container_image" yaml:"container_image"`
}

type Config struct {
	Repository        string             `json:"repository" yaml:"repository"`
	QueueCap          int                `json:"queue_cap" yaml:"queue_cap"`
	CleanInterval     int                `json:"clean_interval" yaml:"clean_interval"`
	Engine            string             `json:"engine" yaml:"engine"`
	Runtime           string             `json:"runtime" yaml:"runtime"`
	SubmissionConfigs []SubmissionConfig `json:"submission_configs" yaml:"submission_configs"`
}

type SourceCode struct {
	SrcTest []SourceFile `json:"src_test"`
	Src     []SourceFile `json:"src"`
}

type SourceFile struct {
	Filename   string `json:"filename"`
	SourceCode string `json:"src"`
}

type Submission struct {
	ExerciseId string     `json:"id"`
	Type       string     `json:"subtype"`
	Source     SourceCode `json:"source"`
}

type Build struct {
	Success bool
	BinPath string
	Stderr  []byte
	Stdout  []byte
}

type Runtime struct {
	Stdout []byte `json:"stdout"`
	Stderr []byte `json:"stderr"`
}

type SubmissionResult struct {
	Success bool         `json:"success"`
	Build   string       `json:"build"`
	Tests   []TestResult `json:"tests"`
	Metrics Metrics      `json:"metrics"`
}

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
	ExitCode int64   `json:"exit_code"`
	WallTime float64 `json:"wall_time"` // running wall time (s)
	CpuTime  uint64  `json:"cpu_time"`  // total CPU time consumed (ns)
	Memory   uint64  `json:"memory"`    // peak memory recorded (bytes)
}
