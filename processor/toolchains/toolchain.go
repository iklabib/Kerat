package toolchains

import (
	"fmt"

	"codeberg.org/iklabib/kerat/processor/types"
)

type Toolchain interface {
	Prep() error
	Build() (types.Build, error)
	Clean() error
}

func NewToolchain(submission types.Submission, repository string) (Toolchain, error) {
	switch submission.Type {
	case "csharp":
		cs, err := NewCsharp(submission, repository)
		return cs, err
	}

	return nil, fmt.Errorf("unsupported type \"%s\"", submission.Type)
}
