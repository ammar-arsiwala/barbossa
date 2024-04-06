package service

import (
	"context"

	"github.com/docker/docker/client"
)

type Command interface {
	Perform(context.Context, *client.Client) error
}

func NewClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return cli, nil
}
