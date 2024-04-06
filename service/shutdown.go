package service

import (
	"context"
	"log"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/keshavchand/barbossa/models"
)

type Shutdown struct {
	ContainerName string
	Graceful      bool
}

func NewShutdown(req models.ShutdownRequest) []Shutdown {
	var sd []Shutdown
	for _, i := range req.Info {
		sd = append(sd, Shutdown{
			ContainerName: i.Name,
			Graceful:      i.Graceful,
		})
	}

	return sd
}

func (s *Shutdown) Perform(ctx context.Context, c *client.Client) error {
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
		log.Println("shutdown container:", err)
		return FnErrApiError(err)
	}

	signal := "SIGKILL"
	if s.Graceful {
		signal = "SIGSTOP"
	}

	for _, ctr := range ctr {
		err = c.ContainerStop(ctx, ctr.ID, container.StopOptions{
			Signal: signal,
		})

		if err != nil {
			return FnErrApiError(err)
		}
	}

	return nil
}
