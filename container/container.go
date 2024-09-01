package container

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"codeberg.org/iklabib/kerat/model"
)

type Engine struct {
	runtime string
	path    string
	args    map[string]string
}

func NewEngine(config model.GlobalConfig, submissionConfigs []model.SubmissionConfig) Engine {
	e := Engine{
		path:    config.Engine,
		runtime: config.Runtime,
		args:    map[string]string{},
	}

	for _, v := range submissionConfigs {
		e.args[v.Id] = strings.Join(e.buildEngineArgs(v), " ")
	}

	return e
}

func (e Engine) IsSupported(id string) bool {
	_, exist := e.args[id]
	return exist
}

// docker run arguments based on config
// I don't feel like to bother with Docker SDK
func (e Engine) buildEngineArgs(config model.SubmissionConfig) []string {
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

	if e.runtime != "" {
		args = append(args, "--runtime", e.runtime)
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

func (e Engine) Run(ctx context.Context, containerName string, submission *model.Submission) (*model.Result, error) {
	stdin, err := json.Marshal(submission)
	if err != nil {
		return nil, err
	}

	args := []string{e.args[submission.Type], "--name", containerName, "kerat:" + submission.Type}

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, e.path, args...)
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

func (e Engine) Kill(id string) {
	cmd := exec.Command(e.path, "kill", id)
	cmd.Run()
}
