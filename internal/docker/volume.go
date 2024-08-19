package docker

import (
	"context"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
)

type VolumeRequest struct {
	// Name is the name of the volume to create, if empty a random name will be
	// generated by the daemon
	Name   string
	Target string
	Labels map[string]string
}

func (d *Client) CreateVolume(ctx context.Context, req *VolumeRequest) (mount.Mount, error) {
	if req.Labels == nil {
		req.Labels = make(map[string]string)
	}

	v, err := d.cli.VolumeCreate(ctx, volume.CreateOptions{
		Name:   req.Name,
		Labels: req.Labels,
	})
	if err != nil {
		return mount.Mount{}, err
	}

	return mount.Mount{
		Type:   mount.TypeVolume,
		Source: v.Name,
		Target: req.Target,
		VolumeOptions: &mount.VolumeOptions{
			Labels: d.withDefaultLabels(req.Labels),
		},
	}, nil
}

func (d *Client) RemoveVolume(ctx context.Context, v mount.Mount) error {
	return d.cli.VolumeRemove(ctx, v.Source, true)
}
