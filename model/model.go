package model

import (
	"os"
	"time"
)

type Submission struct {
	Type           string       `json:"type"`
	SourceCodeTest string       `json:"src_test"`
	SourceFiles    []SourceFile `json:"src"`
}

type SourceFile struct {
	Filename   string `json:"filename"`
	Path       string `json:"path,omitempty"`
	SourceCode string `json:"src"`
}

type BuildError struct {
	Filename string `json:"filename"`
	Message  string `json:"message"`
	Line     int    `json:"line"`
}

type SourceCode struct {
	SourceCodeTest string       `json:"src_test"`
	SourceCodes    []SourceFile `json:"src"`
}

type Metrics struct {
	Signal   os.Signal     `json:"signal"`
	ExitCode int           `json:"exit_code"`
	SysTime  time.Duration `json:"sys_time"`
	UserTime time.Duration `json:"time"`
	WallTime time.Duration `json:"wall_time"`
	Memory   int64         `json:"memory"`
}

type Result struct {
	Stdout  string  `json:"stdout"`
	Stderr  string  `json:"stderr"`
	Message string  `json:"message"`
	Metric  Metrics `json:"metric"`
}

type SubmissionConfig struct {
	Id         string         `json:"id" yaml:"cpus"`
	Cpus       float64        `json:"cpus" yaml:"cpus"`
	MaxPids    int            `json:"max_pids" yaml:"max_pids"`
	MaxSwap    int            `json:"max_swap" yaml:"max_swap"`     // MiB
	MaxMemory  int            `json:"max_memory" yaml:"max_memory"` // MiB
	Timeout    int            `json:"timeout" yaml:"timeout"`       // wall-time in seconds
	Privileged bool           `json:"privileged" yaml:"privileged"`
	Ulimits    map[string]int `json:"ulimits" yaml:"ulimits"`
}

type GlobalConfig struct {
	QueueCap int      `json:"queue_cap" yaml:"queue_cap"`
	Enables  []string `json:"enables" yaml:"enables"`
	Engine   string   `json:"engine" yaml:"engine"`
	Runtime  string   `json:"runtime" yaml:"runtime"`
}
