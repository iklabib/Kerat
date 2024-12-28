package container

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"codeberg.org/iklabib/kerat/model"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
)

type Engine struct {
	client            *client.Client
	runtime           string
	hostConfigs       map[string]container.HostConfig
	submissionConfigs map[string]model.SubmissionConfig
}

func NewEngine(config model.Config) (*Engine, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed start container engine")
	}

	if config.Runtime == "" {
		config.Runtime = "runc"
	}

	engine := &Engine{
		client:            cli,
		runtime:           config.Runtime,
		hostConfigs:       make(map[string]container.HostConfig),
		submissionConfigs: make(map[string]model.SubmissionConfig),
	}

	for _, v := range config.SubmissionConfigs {
		engine.submissionConfigs[v.Id] = v
		engine.hostConfigs[v.Id] = engine.buildHostConfig(v.Id)
	}

	return engine, nil
}

func (e *Engine) buildHostConfig(subType string) container.HostConfig {
	config := e.submissionConfigs[subType]
	resources := container.Resources{
		Memory:     config.MaxMemory * 1024 * 1024,
		CPUPeriod:  config.CPUPeriod,
		CPUQuota:   config.CPUQuota,
		MemorySwap: config.MaxSwap * 1024 * 1024,
		PidsLimit:  &config.MaxPids,
	}

	hostConfig := container.HostConfig{
		AutoRemove: false,
		Resources:  resources,
		Runtime:    e.runtime,
	}

	// posible values
	// core, cpu, data, fsize, locks,
	// memlock, msgqueue, nice, nofile,
	// nproc, rss, rtprio, rttime,
	// sigpending, stack
	for k, v := range config.Ulimits {
		ulim, err := units.ParseUlimit(k)
		if err != nil {
			panic(fmt.Errorf("failed to parse ulimit"))
		}
		ulim.Soft = v
		ulim.Hard = -1
		hostConfig.Ulimits = append(hostConfig.Ulimits, ulim)
	}

	return hostConfig
}

func (e *Engine) IsSupported(subType string) bool {
	_, ok := e.submissionConfigs[subType]
	return ok
}

func (e *Engine) Check() error {
	_, err := e.client.Ping(context.Background())
	return err
}

func (e *Engine) Create(ctx context.Context, subType string) (string, error) {
	submissionConfig := e.submissionConfigs[subType]
	hostConfig := e.hostConfigs[subType]

	var containerConfig = container.Config{
		Hostname:        "box",
		Domainname:      "box",
		NetworkDisabled: true,
		Image:           submissionConfig.ContainerImage,
	}

	resp, err := e.client.ContainerCreate(context.Background(), &containerConfig, &hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("error create container: %w", err)
	}

	return resp.ID, nil
}

func (e *Engine) Copy(ctx context.Context, payload CopyPayload) error {
	opt := container.CopyToContainerOptions{}
	return e.client.CopyToContainer(ctx, payload.ContainerId, payload.Dest, payload.Content, opt)
}

func (e *Engine) Run(ctx context.Context, payload RunPayload) (ContainerResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer e.Remove(payload.ContainerId)

	var res ContainerResult
	if err := e.client.ContainerStart(ctx, payload.ContainerId, container.StartOptions{}); err != nil {
		return res, fmt.Errorf("error start container: %w", err)
	}

	statusCh, errCh := e.client.ContainerWait(ctx, payload.ContainerId, container.WaitConditionNotRunning)
	select {
	case <-ctx.Done(): // cancelled
		return res, ctx.Err()

	case err := <-errCh:
		if err != nil {
			return res, fmt.Errorf("error waiting for container: %w", err)
		}

	case containerStat := <-statusCh:
		if containerStat.Error != nil {
			return res, fmt.Errorf("container %s exited with status code %d error message: %s", payload.ContainerId[:8], containerStat.StatusCode, containerStat.Error.Message)
		} else if containerStat.StatusCode != 0 {
			return res, fmt.Errorf("container %s exited with status code %d", payload.ContainerId[:8], containerStat.StatusCode)
		}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	out, err := e.client.ContainerLogs(ctx, payload.ContainerId, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return res, fmt.Errorf("error reading container output: %w", err)
	}

	_, err = StdCopy(&stdout, &stderr, out)
	if err != nil {
		return res, fmt.Errorf("error reading container output: %w", err)
	}

	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		return res, fmt.Errorf("error deserialize output: %s", err.Error())
	}

	return res, nil
}

func (e *Engine) Remove(id string) error {
	return e.client.ContainerRemove(context.Background(), id, container.RemoveOptions{Force: true})
}

func (e *Engine) Kill(id string) error {
	return e.client.ContainerKill(context.Background(), id, "SIGKILL")
}
