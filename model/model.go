package model

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
