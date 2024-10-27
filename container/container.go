package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"codeberg.org/iklabib/kerat/model"
	"codeberg.org/iklabib/kerat/util"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
)

type Engine struct {
	client            *client.Client
	config            model.Config
	hostConfigs       map[string]container.HostConfig
	submissionConfigs map[string]model.SubmissionConfig
}

func NewEngine(config model.Config) (*Engine, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed start container engine")
	}

	if config.Runtime == "" {
		config.Runtime = cli.ClientVersion()
	}

	engine := &Engine{
		client:            cli,
		config:            config,
		hostConfigs:       make(map[string]container.HostConfig),
		submissionConfigs: make(map[string]model.SubmissionConfig),
	}

	for _, v := range config.SubmissionConfigs {
		engine.submissionConfigs[v.Id] = v
	}

	return engine, nil
}

func (e *Engine) buildHostConfig(subType string) container.HostConfig {
	if c, ok := e.hostConfigs[subType]; ok {
		return c
	}

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
		Runtime:    e.config.Runtime,
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

	// TODO: don't mutate here
	e.hostConfigs[subType] = hostConfig

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

func (e *Engine) Run(ctx context.Context, runPayload model.RunPayload) (*model.Run, error) {
	submissionConfig := e.submissionConfigs[runPayload.Type]
	hostConfig := e.buildHostConfig(runPayload.Type)

	var containerConfig = container.Config{
		Hostname:        "box",
		Domainname:      "box",
		OpenStdin:       true,
		StdinOnce:       true,
		NetworkDisabled: true,
		Image:           "kerat:box",
		Env:             []string{fmt.Sprintf("TIMEOUT=%d", submissionConfig.Timeout)},
	}

	resp, err := e.client.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("error create container: %w", err)
	}

	if err := e.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("error start container: %w", err)
	}

	// TODO: container might in condition of not running after this point, check before hijack
	hijackedResponse, err := e.client.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	})

	if err != nil {
		return nil, fmt.Errorf("error to attach to container: %w", err)
	}

	_, err = hijackedResponse.Conn.Write(append(runPayload.Bin, '\n'))
	if err != nil {
		return nil, fmt.Errorf("error write stdin: %w", err)
	}
	hijackedResponse.CloseWrite()

	statusCh, errCh := e.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("error waiting for container: %w", err)
		}
	case containerStat := <-statusCh:
		if containerStat.Error != nil {
			return nil, fmt.Errorf("container %s exited with status code %d error message: %s", resp.ID, containerStat.StatusCode, containerStat.Error.Message)
		} else if containerStat.StatusCode != 0 {
			return nil, fmt.Errorf("container %s exited with status code %d", resp.ID, containerStat.StatusCode)
		}
	}

	out, err := e.client.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		return nil, fmt.Errorf("error container logging %s", resp.ID)
	}

	output, err := io.ReadAll(out)
	if err != nil {
		return nil, fmt.Errorf("error reading container output: %w", err)
	}

	// I think that hijackedResponse sending control characters to container
	// so we got those in stdout, sanitize the output
	result := model.Run{}
	err = json.Unmarshal(util.SanitizeStdout(output), &result)
	if err != nil {
		return nil, err
	}

	go e.removeInBackground(resp.ID)

	return &result, nil
}

func (e *Engine) removeInBackground(id string) {
	err := e.Remove(id)
	if err != nil {
		log.Printf("failed to remove container %s: %s\n", id, err.Error())
	}
}

func (e *Engine) Remove(id string) error {
	return e.client.ContainerRemove(context.Background(), id, container.RemoveOptions{Force: true})
}

func (e *Engine) Kill(id string) error {
	return e.client.ContainerKill(context.Background(), id, "SIGKILL")
}
