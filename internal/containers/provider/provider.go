package provider

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path/filepath"
)

type ExecConfig struct {
	// The command to be executed in the provider
	Command string

	// The working directory to be used to execute the command
	WorkingDir string
}

type Provider interface {
	Start(ctx context.Context) error
	Teardown(ctx context.Context) error
	Exec(ctx context.Context, config ExecConfig) (io.Reader, error)
}

type ContainerRequest struct {
	Image      string
	Entrypoint []string
	User       string // uid:gid
	Env        Env
	Cmd        []string
	Networks   []string
	Privileged bool
	Files      []File
}

type Env map[string]string

func (e Env) ToSlice() []string {
	s := make([]string, 0, len(e))
	for k, v := range e {
		s = append(s, k+"="+v)
	}
	return s
}

type File struct {
	Contents io.Reader
	Target   string
	Mode     int64
}

// TODO: Jon pls.
func (f File) tar() (io.Reader, error) {
	cbuf := &bytes.Buffer{}
	size, err := io.Copy(cbuf, f.Contents)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}

	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	hdr := &tar.Header{
		Name: filepath.Base(f.Target),
		Mode: f.Mode,
		Size: size,
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return buf, err
	}

	if _, err := io.Copy(tw, cbuf); err != nil {
		return buf, err
	}

	if err := tw.Close(); err != nil {
		return buf, err
	}

	if err := zr.Close(); err != nil {
		return buf, err
	}

	return buf, nil
}