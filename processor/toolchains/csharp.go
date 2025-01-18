package toolchains

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"codeberg.org/iklabib/kerat/processor/types"
	"codeberg.org/iklabib/kerat/processor/util"
)

type Csharp struct {
	id       string
	binPath  string
	template string
	workdir  string
	src      []types.SourceFile
	srcTest  []types.SourceFile
}

func NewCsharp(submission types.Submission, repository string) (*Csharp, error) {
	binPath, err := exec.LookPath("dotnet")
	if err != nil {
		return nil, err
	}

	workdir := filepath.Join(os.TempDir(), submission.ExerciseId)
	templateDir := filepath.Join(repository, "csharp")

	cs := &Csharp{
		id:       submission.ExerciseId,
		binPath:  binPath,
		workdir:  workdir,
		template: templateDir,
		src:      submission.Source.Src,
		srcTest:  submission.Source.SrcTest,
	}
	return cs, nil
}

func (cs *Csharp) Prep() error {
	// copy template
	if util.IsNotExist(cs.workdir) {
		err := os.Mkdir(cs.workdir, 0755)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}

		err = os.CopyFS(cs.workdir, os.DirFS(cs.template))
		if err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	// write source codes to workdir
	sources := append(cs.src, cs.srcTest...)
	for _, v := range sources {
		filePath := filepath.Join(cs.workdir, v.Filename)
		err := os.WriteFile(filePath, []byte(v.SourceCode), 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs *Csharp) Build() (types.Build, error) {
	defer cs.cleanSources()

	stderr := bytes.Buffer{}
	stdout := bytes.Buffer{}

	args := []string{
		"publish",
		"-o", "output",
		"box.csproj",
		"--no-restore",
		"--nologo",
		"-v", "q",
	}

	cmd := exec.Command(cs.binPath, args...)
	cmd.Dir = cs.workdir
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		build := types.Build{Stderr: stderr.Bytes()}
		return build, fmt.Errorf("error to start c# build")
	}

	cmd.Wait()

	// check if compiler was killed
	procState := cmd.ProcessState
	if !procState.Exited() {
		var err error

		wt := procState.Sys().(syscall.WaitStatus)
		if !wt.Signaled() {
			err = fmt.Errorf("c# compiler stopped working")
		} else {
			err = fmt.Errorf("c# compiler stopped working signaled %s", wt.Signal().String())
		}

		build := types.Build{Stderr: stderr.Bytes()}
		return build, err
	}

	// we expect that failed build return 1 as exit code and fill stdout
	if !procState.Success() {
		build := types.Build{Stderr: stderr.Bytes(), Stdout: stdout.Bytes()}
		return build, nil
	}

	binName := "box"
	binPath := filepath.Join(cs.workdir, "output", binName)
	build := types.Build{
		Success: true,
		BinPath: binPath,
	}

	return build, nil
}

// delete source codes, leave the caches behind
func (cs *Csharp) cleanSources() error {
	sources := append(cs.src, cs.srcTest...)
	for _, v := range sources {
		filePath := filepath.Join(cs.workdir, v.Filename)
		err := os.Remove(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// nuke workdir
func (cs *Csharp) Clean() error {
	return os.RemoveAll(cs.workdir)
}
