package service

import (
	"context"
	"errors"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/keshavchand/barbossa/models"
)

type Connect struct {
	ContainerName string
	NetworkName   string
	Force         bool

	Storage EndpointInformationStorage
}

func NewConnect(req models.ConnectRequest, storage EndpointInformationStorage) []Connect {
	var conn []Connect
	for _, i := range req.Info {
		conn = append(conn, Connect{
			ContainerName: i.ContainerName,
			NetworkName:   i.NetworkName,
			Storage:       storage,
		})
	}
	return conn
}

func (s *Connect) Perform(ctx context.Context, c *client.Client) error {
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
		endpoint, found := s.Storage.Get(net.ID, ctr.ID)
		if !found {
			log.Println("XXX network endpoint settings not found")
			return FnErrApiError(errors.New("network endpoint settings not found"))
		}

		// XXX: Dont know whether it will work or not
		// (chand): it does somehow
		if err := c.NetworkConnect(ctx, net.ID, ctr.ID, &network.EndpointSettings{
			MacAddress: endpoint.MacAddress,
			EndpointID: endpoint.EndpointID,
			IPAddress:  endpoint.IPv4Address,
		}); err != nil {
			return FnErrApiError(err)
		}
	}

	return nil
}
