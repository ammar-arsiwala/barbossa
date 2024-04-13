package service

import (
	"context"
	"encoding/json"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/keshavchand/barbossa/models"
)

type Status struct {
	Name string
}

func NewStatus(req models.StatusRequest) *Status {
	return &Status{
		Name: req.Name,
	}
}

func (s *Status) GetStatus(ctx context.Context, cli *client.Client) (*models.StatusResponse, error) {
	ctrs, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", s.Name)),
	})
	if err != nil {
		return nil, FnErrApiError(err)
	}

	var resp models.StatusResponse
	for _, ctr := range ctrs {
		name := ctr.Names[0]
		statJson, err := cli.ContainerStatsOneShot(ctx, ctr.ID)

		if err != nil {
			return nil, FnErrApiError(err)
		}

		var containerStatus models.ContainerStatus
		json.NewDecoder(statJson.Body).Decode(&containerStatus)

		cpuDelta := containerStatus.CPUStats.CPUUsage.TotalUsage - containerStatus.PrecpuStats.CPUUsage.TotalUsage
		systemCPUDelta := containerStatus.CPUStats.SystemCPUUsage - containerStatus.PrecpuStats.SystemCPUUsage
		numsCPUs := containerStatus.CPUStats.OnlineCpus

		cpuPercent := 0.0
		if systemCPUDelta > 0.0 {
			cpuPercent = (float64(cpuDelta) / float64(systemCPUDelta)) * float64(numsCPUs) * 100.0
		}

		memoryPercent := (float64(containerStatus.MemoryStats.Usage) / float64(containerStatus.MemoryStats.Limit)) * 100.0

		stat := models.Status{
			ContainerName:  name,
			Running:        ctr.State == "running",
			TotalProcesses: containerStatus.PidsStats.Current,
			CpuPercent:     cpuPercent,
			MemoryPercent:  memoryPercent,
		}

		for _, v := range containerStatus.Networks {
			stat.RxBytes += float64(v.RxBytes)
			stat.TxBytes += float64(v.TxBytes)
		}

		resp.Data = append(resp.Data, stat)
	}
	return &resp, nil
}
