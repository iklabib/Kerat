package container

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"codeberg.org/iklabib/kerat/model"
)

type Engine struct {
	path    string
	args    []string
	enables map[string]bool
}

func NewEngine(config model.Config) Engine {
	e := Engine{
		path:    config.Engine,
		args:    buildEngineArgs(config),
		enables: map[string]bool{},
	}

	for _, v := range config.Enables {
		e.enables[v] = true
	}

	return e
}

// docker run arguments based on config
// I don't feel like to bother with Docker SDK
func buildEngineArgs(config model.Config) []string {
	cpus := fmt.Sprintf("%.2f", config.Cpus)
	maxMemory := fmt.Sprintf("%dM", config.MaxMemory)
	maxSwap := fmt.Sprintf("%dM", config.MaxSwap)
	maxPids := fmt.Sprintf("%d", config.MaxPids)
	timeout := fmt.Sprintf("TIMEOUT=%d", config.Timeout)

	args := []string{
		"run",
		"-i",
		"--rm",
		"--memory",
		maxMemory,
		"--cpus",
		cpus,
		"--memory-swap",
		maxSwap,
		"--pids-limit",
		maxPids,
		"--env",
		timeout,
		"--attach",
		"STDOUT",
		"--attach",
		"STDERR",
	}

	if config.Runtime != "" {
		args = append(args, "--runtime", config.Runtime)
	}

	if config.Privileged {
		args = append(args, "--privileged")
	}

	for k, v := range config.Ulimits {
		args = append(args, "--ulimit", fmt.Sprintf("%s=%d", k, v))
	}

	return args
}

func (e Engine) Check() error {
	cmd := exec.Command(e.path, "version")
	return cmd.Run()
}

func (e Engine) Run(submission *model.Submission) (*model.Result, error) {
	if !e.enables[submission.Type] {
		err := fmt.Errorf("submission type \"%s\" is unsupported", submission.Type)
		return nil, err
	}

	stdin, err := json.Marshal(submission)
	if err != nil {
		return nil, err
	}

	args := append(e.args, "--name", "kerat_"+submission.Id, "kerat:"+submission.Type)

	var stdout bytes.Buffer
	cmd := exec.Command(e.path, args...)
	cmd.Stdout = &stdout
	cmd.Stdin = bytes.NewReader(stdin)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	result := model.Result{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
