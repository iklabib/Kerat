package box

import (
	"fmt"
	"os"
	"path/filepath"

	"codeberg.org/iklabib/kerat/model"
	"codeberg.org/iklabib/kerat/toolchains"
)

type Box struct {
	workdir   string
	source    model.SourceCode
	toolchain toolchains.Toolchain
}

func NewBox(submission model.Submission) (*Box, error) {
	tc, err := toolchains.NewToolchain(submission.Type)
	if err != nil {
		return nil, err
	}

	workdir, err := os.MkdirTemp("", "box_")
	if err != nil {
		return nil, err
	}

	source := submission.Source
	box := &Box{
		toolchain: tc,
		source:    source,
		workdir:   workdir,
	}

	if err := box.Prep(); err != nil {
		return nil, err
	}

	return box, nil
}

// when we want to reuse box
func LoadBox(workdir string, submission model.Submission) (*Box, error) {
	if _, err := os.Stat(workdir); os.IsNotExist(err) {
		return nil, fmt.Errorf("error to load box because workdir %s not exist", workdir)
	}

	box, err := NewBox(submission)
	if err != nil {
		return nil, err
	}

	return box, nil
}

func (box *Box) Prep() error {
	for _, v := range box.source.Src {
		filePath := filepath.Join(box.workdir, v.Filename)
		err := os.WriteFile(filePath, []byte(v.SourceCode), 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (box *Box) Build() (model.Build, error) {
	var files []string
	sources := append(box.source.Src, box.source.SrcTest...)
	for _, v := range sources {
		files = append(files, v.Filename)
	}

	return box.toolchain.Build(box.workdir, files)
}

func (box *Box) Clean() error {
	return box.toolchain.Clean(box.workdir)
}
