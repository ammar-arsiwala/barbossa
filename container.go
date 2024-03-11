package main

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func StartContainer(ctx context.Context, cli *client.Client, service *Service) string {
	if service.Image[0] == '/' {
		service.Image = "/" + service.Image
	}

	reader, err := cli.ImagePull(ctx, "docker.io/library"+service.Name, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	io.Copy(os.Stdout, reader)
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		if c.Names[0] == "/"+service.Name {
			if c.State == "exited" {
				// Remove the container
				if err := cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{}); err != nil {
					panic(err)
				}
			} else {
				panic("Container already exists and is running")
			}
		}
	}

	mounts := []mount.Mount{}
	for _, mountPoint := range service.MountPoints {
		mounts = append(mounts, mount.Mount{
			Source: mountPoint.Src,
			Target: mountPoint.Dst,
			Type:   mount.TypeBind, //XXX: Dont know what is this
		})
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Hostname: service.Hostname,
		Image:    service.Image,
		Cmd:      service.Exec,
		Tty:      false,
		Env:      service.Env,
	}, &container.HostConfig{
		Mounts: mounts,
	}, nil, nil, service.Name)

	if err != nil {
		panic(err)
	}
	return resp.ID
}

func CreateNetworkIfNotExists(ctx context.Context, cli *client.Client, net *Network) (string, error) {
	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		panic(err)
	}

	for _, n := range networks {
		if n.Name == net.Name {
			if err := cli.NetworkRemove(ctx, n.ID); err != nil {
				return "", err
			}
		}
	}

	n, err := cli.NetworkCreate(ctx, net.Name, types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet:  net.Subnet,
					Gateway: net.Gateway,
				},
			},
		},
	})

	return n.ID, err
}
