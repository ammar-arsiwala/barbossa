package service

import (
	"context"
	"log"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/keshavchand/barbossa/models"
)

type Startup struct {
	ContainerName string
}

func NewStartup(req models.StartupRequest) []Startup {
	var sd []Startup
	for _, i := range req.Info {
		sd = append(sd, Startup{
			ContainerName: i.Name,
		})
	}
	return sd
}

func (s *Startup) Perform(ctx context.Context, c *client.Client) error {
	ctr, err := c.ContainerList(ctx, container.ListOptions{
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
		log.Println("startup container:", err)
		return FnErrApiError(err)
	}

	for _, ctr := range ctr {
		if err = c.ContainerStart(ctx, ctr.ID, container.StartOptions{}); err != nil {
			return FnErrApiError(err)
		}
	}

	return nil
}
