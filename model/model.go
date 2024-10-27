package model

import (
	"time"
)

type Submission struct {
	Type   string     `json:"type"`
	Source SourceCode `json:"source"`
}

type SourceCode struct {
	SrcTest []SourceFile `json:"src_test"`
	Src     []SourceFile `json:"src"`
}

type SourceFile struct {
	Filename   string `json:"filename"`
	SourceCode string `json:"src"`
}

type BuildError struct {
	Filename string `json:"filename"`
	Message  string `json:"message"`
	Line     int    `json:"line"`
}

type Build struct {
	Success bool
	Bin     []byte
	BinName string
	Stderr  string
}

type RunPayload struct {
	Type string
	Bin  []byte
}

type Run struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

type RunResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

type Metrics struct {
	ExitCode int           `json:"exit_code"`
	SysTime  time.Duration `json:"sys_time"`
	UserTime time.Duration `json:"time"`
	WallTime time.Duration `json:"wall_time"`
	Memory   int64         `json:"memory"`
}

type SubmissionConfig struct {
	Id        string           `json:"id" yaml:"id"`
	CPUPeriod int64            `json:"cpu_period" yaml:"cpu_period"`
	CPUQuota  int64            `json:"cpu_quota" yaml:"cpu_quota"`
	MaxPids   int64            `json:"max_pids" yaml:"max_pids"`
	MaxSwap   int64            `json:"max_swap" yaml:"max_swap"`     // MiB
	MaxMemory int64            `json:"max_memory" yaml:"max_memory"` // MiB
	Timeout   int              `json:"timeout" yaml:"timeout"`       // wall-time in seconds
	Ulimits   map[string]int64 `json:"ulimits" yaml:"ulimits"`
}

type Config struct {
	QueueCap          int                `json:"queue_cap" yaml:"queue_cap"`
	Engine            string             `json:"engine" yaml:"engine"`
	Runtime           string             `json:"runtime" yaml:"runtime"`
	SubmissionConfigs []SubmissionConfig `json:"submission_configs" yaml:"submission_configs"`
}
