package toolchains

import (
	"fmt"

	"codeberg.org/iklabib/kerat/model"
)

type Toolchain interface {
	Build(srcPath string, files []string) (model.Build, error)
	Clean(workdir string) error
}

func NewToolchain(typeName string) (Toolchain, error) {
	switch typeName {
	case "kotlin":
		kotlin, err := NewKotlin()
		return kotlin, err
	}

	return nil, fmt.Errorf("unsupported type \"%s\"", typeName)
}
