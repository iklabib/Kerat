package processor

import (
	"archive/tar"
	"bytes"
	"time"

	"codeberg.org/iklabib/kerat/processor/types"
)

func TarSources(files types.SourceCode) (bytes.Buffer, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	sources := append(files.Src, files.SrcTest...)

	for _, file := range sources {
		header := &tar.Header{
			Name:    file.Filename,
			Size:    int64(len(file.SourceCode)),
			Mode:    0644,
			ModTime: time.Now(),
		}

		if err := tw.WriteHeader(header); err != nil {
			return buf, err
		}

		if _, err := tw.Write([]byte(file.SourceCode)); err != nil {
			return buf, err
		}
	}

	return buf, nil
}

func TarBinary(filename string, bin []byte) (bytes.Buffer, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	header := &tar.Header{
		Name:    filename,
		Size:    int64(len(bin)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return buf, err
	}

	if _, err := tw.Write(bin); err != nil {
		return buf, err
	}

	return buf, nil
}
