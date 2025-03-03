package processor

import (
	"context"
	"fmt"
	"os"

	"codeberg.org/iklabib/kerat/processor/container"
	"codeberg.org/iklabib/kerat/processor/memo"
	"codeberg.org/iklabib/kerat/processor/toolchains"
	"codeberg.org/iklabib/kerat/processor/types"
)

type SubmissionProcessor struct {
	engine *container.Engine
	config *types.Config
}

func NewSubmissionProcessor(config *types.Config) (*SubmissionProcessor, error) {
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

func (p *SubmissionProcessor) ProcessSubmission(ctx context.Context, submission types.Submission, submissionId string) (types.SubmissionResult, error) {
	if !p.engine.IsSupported(submission.Type) {
		return types.SubmissionResult{}, fmt.Errorf("submission type %q is unsupported", submission.Type)
	}

	switch submission.Type {
	case "python":
		return p.processInterpretedSubmission(ctx, submission)
	case "csharp":
		return p.processCompiledSubmission(ctx, submission)
	default:
		return types.SubmissionResult{}, fmt.Errorf("unknown submission type: %s", submission.Type)
	}
}

func (p *SubmissionProcessor) processInterpretedSubmission(ctx context.Context, submission types.Submission) (types.SubmissionResult, error) {
	result := types.SubmissionResult{}

	containerId, err := p.engine.Create(context.Background(), submission.Type)
	if err != nil {
		return result, fmt.Errorf("container creation error: %v", err)
	}

	defer func() {
		go p.engine.Remove(containerId)
	}()

	content, err := TarSources(submission.Source)
	if err != nil {
		return result, fmt.Errorf("creating tar error: %v", err)
	}

	copyPayload := types.CopyPayload{ContainerId: containerId, Dest: "/workspace", Content: &content}
	if err := p.engine.Copy(context.Background(), copyPayload); err != nil {
		return result, fmt.Errorf("copying tar error: %v", err)
	}

	ret, err := p.engine.Run(ctx, types.RunPayload{ContainerId: containerId, SubmissionType: submission.Type})
	if err != nil {
		return result, fmt.Errorf("run error: %v", err)
	}

	result.Success = ret.Success
	result.Build = ret.Message
	result.Tests = ret.Output
	result.Metrics = ret.Metrics

	return result, nil
}

func (p *SubmissionProcessor) processCompiledSubmission(ctx context.Context, submission types.Submission) (types.SubmissionResult, error) {
	caches := memo.NewBoxCaches(p.config.CleanInterval)
	exerciseId := submission.ExerciseId

	result := types.SubmissionResult{}

	tc, ok := caches.LoadToolchain(exerciseId)
	if !ok {
		var err error
		tc, err = toolchains.NewToolchain(submission, p.config.Repository)
		if err != nil {
			return result, fmt.Errorf("failed to create toolchain: %v", err)
		}
		caches.AddToolchain(exerciseId, tc)
	}

	if err := tc.Prep(); err != nil {
		return result, fmt.Errorf("prep error: %v", err)
	}

	build, err := tc.Build()
	if err != nil {
		return result, fmt.Errorf("build error: %v", err)
	}

	if !build.Success {
		result.Build = string(build.Stdout)
		result.Tests = []types.TestResult{}

		return result, nil
	}

	bin, err := os.ReadFile(build.BinPath)
	if err != nil {
		return result, fmt.Errorf("failed to read binary: %v", err)
	}

	containerId, err := p.engine.Create(context.Background(), submission.Type)
	if err != nil {
		return result, fmt.Errorf("container creation error: %v", err)
	}

	defer func() {
		go p.engine.Remove(containerId)
	}()

	content, err := TarBinary("box", bin)
	if err != nil {
		return result, fmt.Errorf("creating tar error: %v", err)
	}

	copyPayload := types.CopyPayload{ContainerId: containerId, Dest: "/workspace", Content: &content}
	if err := p.engine.Copy(context.Background(), copyPayload); err != nil {
		return result, fmt.Errorf("copying tar error: %v", err)
	}

	ret, err := p.engine.Run(ctx, types.RunPayload{ContainerId: containerId, SubmissionType: submission.Type})
	if err != nil {
		return result, fmt.Errorf("runtime error: %v", err)
	}

	result.Success = ret.Success
	result.Tests = ret.Output
	result.Metrics = ret.Metrics

	return result, nil
}
