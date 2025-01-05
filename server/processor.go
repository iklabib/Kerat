package server

import (
	"context"
	"fmt"
	"os"

	"codeberg.org/iklabib/kerat/container"
	"codeberg.org/iklabib/kerat/memo"
	"codeberg.org/iklabib/kerat/model"
	"codeberg.org/iklabib/kerat/toolchains"
	"codeberg.org/iklabib/kerat/util"
)

type SubmissionProcessor struct {
	engine *container.Engine
	config *model.Config
}

func NewSubmissionProcessor(config *model.Config) (*SubmissionProcessor, error) {
	engine, err := container.NewEngine(*config)
	if err != nil {
		return nil, err
	}

	if err := engine.Check(); err != nil {
		return nil, err
	}

	return &SubmissionProcessor{
		engine: engine,
		config: config,
	}, nil
}

func (p *SubmissionProcessor) ProcessSubmission(ctx context.Context, submission model.Submission, submissionId string) (SubmissionResult, error) {
	if !p.engine.IsSupported(submission.Type) {
		return SubmissionResult{}, fmt.Errorf("submission type %q is unsupported", submission.Type)
	}

	switch submission.Type {
	case "python":
		return p.processInterpretedSubmission(ctx, submission)
	case "csharp":
		return p.processCompiledSubmission(ctx, submission)
	default:
		return SubmissionResult{}, fmt.Errorf("unknown submission type: %s", submission.Type)
	}
}

func (p *SubmissionProcessor) processInterpretedSubmission(ctx context.Context, submission model.Submission) (SubmissionResult, error) {
	containerId, err := p.engine.Create(context.Background(), submission.Type)
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("container creation error: %v", err)
	}

	defer func() {
		go p.engine.Remove(containerId)
	}()

	content, err := util.TarSources(submission.Source)
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("creating tar error: %v", err)
	}

	copyPayload := container.CopyPayload{ContainerId: containerId, Dest: "/workspace", Content: &content}
	if err := p.engine.Copy(context.Background(), copyPayload); err != nil {
		return SubmissionResult{}, fmt.Errorf("copying tar error: %v", err)
	}

	ret, err := p.engine.Run(ctx, container.RunPayload{ContainerId: containerId, SubmissionType: submission.Type})
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("run error: %v", err)
	}

	return SubmissionResult{
		Success: ret.Success,
		Build:   ret.Message,
		Tests:   ret.Output,
		Metrics: ret.Metrics,
	}, nil
}

func (p *SubmissionProcessor) processCompiledSubmission(ctx context.Context, submission model.Submission) (SubmissionResult, error) {
	caches := memo.NewBoxCaches(p.config.CleanInterval)
	exerciseId := submission.ExerciseId

	tc, ok := caches.LoadToolchain(exerciseId)
	if !ok {
		var err error
		tc, err = toolchains.NewToolchain(submission, p.config.Repository)
		if err != nil {
			return SubmissionResult{}, fmt.Errorf("failed to create toolchain: %v", err)
		}
		caches.AddToolchain(exerciseId, tc)
	}

	if err := tc.Prep(); err != nil {
		return SubmissionResult{}, fmt.Errorf("prep error: %v", err)
	}

	build, err := tc.Build()
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("build error: %v", err)
	}

	if !build.Success {
		return SubmissionResult{
			Build: string(build.Stdout),
			Tests: []container.TestResult{},
		}, nil
	}

	bin, err := os.ReadFile(build.BinPath)
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("failed to read binary: %v", err)
	}

	containerId, err := p.engine.Create(context.Background(), submission.Type)
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("container creation error: %v", err)
	}

	defer func() {
		go p.engine.Remove(containerId)
	}()

	content, err := util.TarBinary("box", bin)
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("creating tar error: %v", err)
	}

	copyPayload := container.CopyPayload{ContainerId: containerId, Dest: "/workspace", Content: &content}
	if err := p.engine.Copy(context.Background(), copyPayload); err != nil {
		return SubmissionResult{}, fmt.Errorf("copying tar error: %v", err)
	}

	ret, err := p.engine.Run(ctx, container.RunPayload{ContainerId: containerId, SubmissionType: submission.Type})
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("runtime error: %v", err)
	}

	return SubmissionResult{
		Success: ret.Success,
		Tests:   ret.Output,
		Metrics: ret.Metrics,
	}, nil
}
