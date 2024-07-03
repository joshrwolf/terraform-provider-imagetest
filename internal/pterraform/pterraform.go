package pterraform

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

type Pterraform struct {
	// source is the path to the directory where the source files are located
	source fs.FS

	// workingDir is the path to the directory where the copy of sourceDir is
	// placed, and the working directory for terraform operations
	workingDir string
}

type Options func(*Pterraform)

func NewPterraform(source fs.FS, opts ...Options) *Pterraform {
	p := &Pterraform{
		source: source,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p Pterraform) Apply(ctx context.Context) error {
	if err := p.setup(ctx); err != nil {
		return err
	}

	// terraform apply
	{
		_, err := p.exec(ctx, "terraform", []string{"apply", "-auto-approve", "-input=false"}...)
		if err != nil {
			return fmt.Errorf("terraform apply: %w", err)
		}
	}

	// terraform output json
	{
		_, err := p.exec(ctx, "terraform", []string{"output", "-json"}...)
		if err != nil {
			return fmt.Errorf("terraform output: %w", err)
		}
	}

	return nil
}

func (p Pterraform) Destroy(ctx context.Context) error {
	// terraform destroy
	{
		_, err := p.exec(ctx, "terraform", []string{"destroy", "-auto-approve", "-input=false"}...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pterraform) setup(ctx context.Context) error {
	tdir, err := os.MkdirTemp("", "imagetest-pterraform")
	if err != nil {
		return err
	}
	p.workingDir = tdir

	// Copy the source directory to the working directory, skipping symlinks
	if err := fs.WalkDir(p.source, ".", func(path string, d fs.DirEntry, _ error) error {
		targ := filepath.Join(p.workingDir, filepath.FromSlash(path))
		if d.IsDir() {
			if err := os.MkdirAll(targ, 0755); err != nil {
				return err
			}
			return nil
		}

		r, err := p.source.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()

		info, err := r.Stat()
		if err != nil {
			return err
		}

		w, err := os.OpenFile(targ, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
		if err != nil {
			return err
		}
		if _, err := io.Copy(w, r); err != nil {
			return err
		}

		return w.Close()
	}); err != nil {
		return err
	}

	// terraform init
	if _, err := p.exec(ctx, "terraform", []string{"init", "-input=false"}...); err != nil {
		return fmt.Errorf("terraform init: %w", err)
	}

	return nil
}

// exec runs a terraform command and captures the output
func (p Pterraform) exec(ctx context.Context, command string, args ...string) (io.Reader, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = os.Stdin
	cmd.Dir = p.workingDir

	// TODO: Append more env vars
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	var cout bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &cout)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		_, _ = io.Copy(mw, stdout)
	}()

	go func() {
		_, _ = io.Copy(mw, stderr)
	}()

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("command exited with error: %w\n%s", err, &cout)
	}

	return &cout, nil
}
