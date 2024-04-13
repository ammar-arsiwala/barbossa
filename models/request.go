package models

import "fmt"

// swagger:parameters shutdown
type ShutdownRequest struct {
	Info []struct {
		Name     string `json:"name"`
		Graceful bool   `json:"graceful"`
	} `json:"info"`
}

func (s *ShutdownRequest) Verify() error {
	for idx, i := range s.Info {
		if i.Name == "" {
			return fmt.Errorf("name at index %d is empty", idx)
		}
	}

	return nil
}

// swagger:parameters startup
type StartupRequest struct {
	Info []struct {
		Name string `json:"name"`
	} `json:"info"`
}

func (s *StartupRequest) Verify() error {
	for idx, i := range s.Info {
		if i.Name == "" {
			return fmt.Errorf("name at index %d is empty", idx)
		}
	}

	return nil
}

// swagger:parameters partition
type PartitionRequest struct {
	Info []struct {
		ContainerName string `json:"container_name"`
		NetworkName   string `json:"network_name"`
		Force         bool   `json:"force"`
	} `json:"info"`
}

func (s *PartitionRequest) Verify() error {
	for idx, i := range s.Info {
		if i.ContainerName == "" {
			return fmt.Errorf("container name at index %d is empty", idx)
		}
		if i.NetworkName == "" {
			return fmt.Errorf("network name at index %d is empty", idx)
		}
	}
	return nil
}

// swagger:parameters connect
type ConnectRequest struct {
	Info []struct {
		ContainerName string `json:"container_name"`
		NetworkName   string `json:"network_name"`
	} `json:"info"`
}

func (s *ConnectRequest) Verify() error {
	for idx, i := range s.Info {
		if i.ContainerName == "" {
			return fmt.Errorf("container name at index %d is empty", idx)
		}
		if i.NetworkName == "" {
			return fmt.Errorf("network name at index %d is empty", idx)
		}
	}
	return nil
}

// swagger:parameters status
type StatusRequest struct {
	// in:query
	Name string `json:"name"`
}
