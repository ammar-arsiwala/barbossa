package service

import (
	"context"
	"errors"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/keshavchand/barbossa/models"
)

type Partition struct {
	ContainerName string
	NetworkName   string
	Force         bool

	Storage EndpointInformationStorage
}

func NewPartition(req models.PartitionRequest, storage EndpointInformationStorage) []Partition {
	var par []Partition
	for _, i := range req.Info {
		par = append(par, Partition{
			ContainerName: i.ContainerName,
			NetworkName:   i.NetworkName,
			Force:         i.Force,
			Storage:       storage,
		})
	}
	return par
}

func (s *Partition) Perform(ctx context.Context, c *client.Client) error {
	ctrs, err := c.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "name",
			Value: s.ContainerName,
		}),
	})

	if client.IsErrNotFound(err) {
		return FnErrContainerNotFound(s.ContainerName)
	}

	if err != nil {
		log.Println("network partition:", err)
		return FnErrApiError(err)
	}

	nets, err := c.NetworkList(ctx, types.NetworkListOptions{
		Filters: filters.NewArgs(filters.Arg("name", s.NetworkName)),
	})

	if client.IsErrNotFound(err) {
		return FnErrContainerNotFound(s.ContainerName)
	}

	if err != nil {
		log.Println("network partition:", err)
		return FnErrApiError(err)
	}

	for _, ctr := range ctrs {
		net := nets[0]
		networks, err := c.NetworkInspect(ctx, net.ID, types.NetworkInspectOptions{
			Verbose: true,
		})

		if err != nil {
			return FnErrApiError(err)
		}

		endpointResource, found := networks.Containers[s.ContainerName]
		if !found {
			endpointResource, found = networks.Containers[ctr.ID]
		}

		if !found {
			log.Println("XXX API Error: cant find endpoint info")
			return FnErrApiError(errors.New("cant find endpoint info"))
		}

		s.Storage.Store(net.ID, ctr.ID, endpointResource)

		if err := c.NetworkDisconnect(ctx, net.ID, ctr.ID, s.Force); err != nil {
			return FnErrApiError(err)
		}
	}

	return nil
}
