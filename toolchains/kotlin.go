package toolchains

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"codeberg.org/iklabib/kerat/model"
)

type Kotlin struct {
	binPath string
}

func NewKotlin() (*Kotlin, error) {
	binPath, err := exec.LookPath("kotlinc-native")
	if err != nil {
		return nil, err
	}

	kt := &Kotlin{
		binPath: binPath,
	}
	return kt, nil
}

func (k Kotlin) PreBuild(workdir string, source model.SourceCode) error {
	// TODO: make sure source code filename as relative path
	for _, v := range source.SourceCodes {
		filePath := filepath.Join(workdir, v.Filename)
		err := os.WriteFile(filePath, []byte(v.SourceCode), 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k Kotlin) Build(workdir string, files []string) (model.Build, error) {
	stderr := bytes.Buffer{}

	// FIXME: either you iterate the whole directory and find all *.kt, cause go exec not expanding wildcard
	// FIXME: static build until I can figure out this lib loading errors
	args := []string{"-nowarn", "-o", "Main"}
	cmd := exec.Command(k.binPath)
	cmd.Args = append(args, files...)
	cmd.Dir = workdir
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		build := model.Build{Stderr: stderr.String()}
		return build, fmt.Errorf("error to start kotlin build")
	}

	cmd.Wait()

	// check if kotlin compiler was killed
	procState := cmd.ProcessState
	if !procState.Exited() {
		var err error

		wt := procState.Sys().(syscall.WaitStatus)
		if !wt.Signaled() {
			err = fmt.Errorf("kotlin compiler stopped working")
		} else {
			err = fmt.Errorf("kotlin compiler stopped working signaled %s", wt.Signal().String())
		}

		build := model.Build{Stderr: stderr.String()}
		return build, err
	}

	// we expect that failed build return 1 as exit code and fill stderr
	if !procState.Success() {
		build := model.Build{Stderr: stderr.String()}
		return build, nil
	}

	binName := "Main.kexe"
	binPath := filepath.Join(workdir, binName)
	bin, err := os.ReadFile(binPath)
	if err != nil {
		return model.Build{}, fmt.Errorf("error to read compiled binary: %s", err.Error())
	}

	build := model.Build{
		Success: true,
		Bin:     bin,
		BinName: binName,
	}

	return build, nil
}
