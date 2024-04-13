package models

import (
	"fmt"
	"strings"
	"time"
)

type Status struct {
	ContainerName  string  `json:"container_name"`
	Running        bool    `json:"running"`
	TotalProcesses int     `json:"total_processes"`
	CpuPercent     float64 `json:"cpu_percent"`
	MemoryPercent  float64 `json:"memory_percent"`

	// Network
	RxBytes float64 `json:"rx_bytes"`
	TxBytes float64 `json:"tx_bytes"`
}

// swagger:response statusResponse
type StatusResponse struct {
	Data []Status `json:"data"`
}

func (s *StatusResponse) String() string {
	// Pretty Format

	var res strings.Builder
	for _, v := range s.Data {
		res.WriteString(fmt.Sprintf("Container Name: %s\n", v.ContainerName))
		res.WriteString(fmt.Sprintf("Running: %t\n", v.Running))
		res.WriteString(fmt.Sprintf("Total Processes: %d\n", v.TotalProcesses))
		res.WriteString(fmt.Sprintf("CPU Percent: %.2f\n", v.CpuPercent))
		res.WriteString(fmt.Sprintf("Memory Percent: %.2f\n", v.MemoryPercent))
		res.WriteString(fmt.Sprintf("Rx Bytes: %.2f\n", v.RxBytes))
		res.WriteString(fmt.Sprintf("Tx Bytes: %.2f\n", v.TxBytes))
		res.WriteString("\n")
	}

	return res.String()
}

type Network struct {
	RxBytes   int `json:"rx_bytes"`
	RxDropped int `json:"rx_dropped"`
	RxErrors  int `json:"rx_errors"`
	RxPackets int `json:"rx_packets"`
	TxBytes   int `json:"tx_bytes"`
	TxDropped int `json:"tx_dropped"`
	TxErrors  int `json:"tx_errors"`
	TxPackets int `json:"tx_packets"`
}

type ContainerStatus struct {
	Read      time.Time `json:"read"`
	PidsStats struct {
		Current int `json:"current"`
	} `json:"pids_stats"`
	Networks    map[string]Network `json:"networks"`
	MemoryStats struct {
		Stats struct {
			TotalPgmajfault         int `json:"total_pgmajfault"`
			Cache                   int `json:"cache"`
			MappedFile              int `json:"mapped_file"`
			TotalInactiveFile       int `json:"total_inactive_file"`
			Pgpgout                 int `json:"pgpgout"`
			Rss                     int `json:"rss"`
			TotalMappedFile         int `json:"total_mapped_file"`
			Writeback               int `json:"writeback"`
			Unevictable             int `json:"unevictable"`
			Pgpgin                  int `json:"pgpgin"`
			TotalUnevictable        int `json:"total_unevictable"`
			Pgmajfault              int `json:"pgmajfault"`
			TotalRss                int `json:"total_rss"`
			TotalRssHuge            int `json:"total_rss_huge"`
			TotalWriteback          int `json:"total_writeback"`
			TotalInactiveAnon       int `json:"total_inactive_anon"`
			RssHuge                 int `json:"rss_huge"`
			HierarchicalMemoryLimit int `json:"hierarchical_memory_limit"`
			TotalPgfault            int `json:"total_pgfault"`
			TotalActiveFile         int `json:"total_active_file"`
			ActiveAnon              int `json:"active_anon"`
			TotalActiveAnon         int `json:"total_active_anon"`
			TotalPgpgout            int `json:"total_pgpgout"`
			TotalCache              int `json:"total_cache"`
			InactiveAnon            int `json:"inactive_anon"`
			ActiveFile              int `json:"active_file"`
			Pgfault                 int `json:"pgfault"`
			InactiveFile            int `json:"inactive_file"`
			TotalPgpgin             int `json:"total_pgpgin"`
		} `json:"stats"`
		MaxUsage int `json:"max_usage"`
		Usage    int `json:"usage"`
		Failcnt  int `json:"failcnt"`
		Limit    int `json:"limit"`
	} `json:"memory_stats"`
	BlkioStats struct {
	} `json:"blkio_stats"`
	CPUStats struct {
		CPUUsage struct {
			PercpuUsage       []int `json:"percpu_usage"`
			UsageInUsermode   int   `json:"usage_in_usermode"`
			TotalUsage        int   `json:"total_usage"`
			UsageInKernelmode int   `json:"usage_in_kernelmode"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
		OnlineCpus     int   `json:"online_cpus"`
		ThrottlingData struct {
			Periods          int `json:"periods"`
			ThrottledPeriods int `json:"throttled_periods"`
			ThrottledTime    int `json:"throttled_time"`
		} `json:"throttling_data"`
	} `json:"cpu_stats"`
	PrecpuStats struct {
		CPUUsage struct {
			PercpuUsage       []int `json:"percpu_usage"`
			UsageInUsermode   int   `json:"usage_in_usermode"`
			TotalUsage        int   `json:"total_usage"`
			UsageInKernelmode int   `json:"usage_in_kernelmode"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
		OnlineCpus     int   `json:"online_cpus"`
		ThrottlingData struct {
			Periods          int `json:"periods"`
			ThrottledPeriods int `json:"throttled_periods"`
			ThrottledTime    int `json:"throttled_time"`
		} `json:"throttling_data"`
	} `json:"precpu_stats"`
}
