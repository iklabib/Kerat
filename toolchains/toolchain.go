package toolchains

import (
	"fmt"

	"codeberg.org/iklabib/kerat/model"
)

type Toolchain interface {
	Prep() error
	Build() (model.Build, error)
	Clean() error
}

func NewToolchain(submission model.Submission, repository string) (Toolchain, error) {
	switch submission.Type {
	case "csharp":
		cs, err := NewCsharp(submission, repository)
		return cs, err
	}

	return nil, fmt.Errorf("unsupported type \"%s\"", submission.Type)
}
