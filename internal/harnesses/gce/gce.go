package gce

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/chainguard-dev/terraform-provider-imagetest/internal/harnesses/base"
	"github.com/chainguard-dev/terraform-provider-imagetest/internal/pterraform"
	"github.com/chainguard-dev/terraform-provider-imagetest/internal/types"
)

var _ types.Harness = &Gce{}

type Gce struct {
	*base.Base

	pt *pterraform.Pterraform
}

//go:embed tf/*
var tf embed.FS

func NewGce() (types.Harness, error) {
	fsys, err := fs.Sub(tf, "tf")
	if err != nil {
		return nil, err
	}

	pt := pterraform.NewPterraform(fsys)

	return &Gce{
		Base: base.New(),
		pt:   pt,
	}, nil
}

// DebugLogCommand implements types.Harness.
func (h *Gce) DebugLogCommand() string {
	// TODO:
	return ``
}

// Destroy implements types.Harness.
func (h *Gce) Destroy(ctx context.Context) error {
	if err := h.pt.Destroy(ctx); err != nil {
		return fmt.Errorf("destroying gce harness: %v", err)
	}
	return nil
}

// Setup implements types.Harness.
func (h *Gce) Setup() types.StepFn {
	return h.WithCreate(func(ctx context.Context) (context.Context, error) {
		if err := h.pt.Apply(ctx); err != nil {
			return ctx, fmt.Errorf("creating gce harness: %v", err)
		}

		return ctx, nil
	})
}

// StepFn implements types.Harness.
func (h *Gce) StepFn(config types.StepConfig) types.StepFn {
	return func(ctx context.Context) (context.Context, error) {
		return ctx, nil
	}
}
