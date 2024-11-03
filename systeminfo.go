package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

func GetCPUStats() (cpu.TimesStat, error) {
	stats, err := cpu.Times(false)
	if err != nil {
		return cpu.TimesStat{}, err
	}
	if len(stats) == 0 {
		return cpu.TimesStat{}, nil
	}

	currStats := stats[0]

	total := currStats.User + currStats.System + currStats.Idle + currStats.Nice +
		currStats.Iowait + currStats.Irq + currStats.Softirq + currStats.Steal +
		currStats.Guest

	if total == 0 {
		return cpu.TimesStat{}, nil
	}

	// Overwrite TimesStat fields with percentage values
	currStats.User = (currStats.User / total) * 100
	currStats.System = (currStats.System / total) * 100
	currStats.Idle = (currStats.Idle / total) * 100
	currStats.Nice = (currStats.Nice / total) * 100
	currStats.Iowait = (currStats.Iowait / total) * 100
	currStats.Irq = (currStats.Irq / total) * 100
	currStats.Softirq = (currStats.Softirq / total) * 100
	currStats.Steal = (currStats.Steal / total) * 100
	currStats.Guest = (currStats.Guest / total) * 100

	return currStats, nil
}

func GetMEMStats() (mem.VirtualMemoryStat, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return mem.VirtualMemoryStat{}, err
	}

	return mem.VirtualMemoryStat{
		Total:       v.Total,
		Used:        v.Used,
		Free:        v.Free,
		UsedPercent: v.UsedPercent,
		Available:   v.Available,
	}, nil
}

type ProcessInfo struct {
	PID         int32
	Name        string
	Username    string
	Memory      uint64
	CPUPercent  float64 // CPU usage percentage
	RunningTime string
}

func GetProcesses(n int) ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var processInfos []ProcessInfo
	for _, p := range procs {
		pid := p.Pid
		name, err := p.Name()
		if err != nil {
			name = "Unknown"
		}

		createTime, err := p.CreateTime()
		if err != nil {
			createTime = 0
		}

		startTime := time.Unix(createTime/1000, 0)
		runningTime := time.Since(startTime).Truncate(time.Second)

		username, err := p.Username()
		if err != nil {
			name = "Unknown"
		}

		memoryInfo, err := p.MemoryInfo()
		if err != nil {
			processInfos = append(processInfos, ProcessInfo{
				PID:         pid,
				Name:        name,
				RunningTime: runningTime.String(),
				Username:    username,
				Memory:      0,
				CPUPercent:  0,
			})
			continue
		}

		memory := memoryInfo.RSS

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			cpuPercent = 0
		}

		processInfos = append(processInfos, ProcessInfo{
			PID:         pid,
			Name:        name,
			RunningTime: runningTime.String(),
			Username:    username,
			Memory:      memory,
			CPUPercent:  cpuPercent,
		})
	}

	sort.Slice(processInfos, func(i, j int) bool {
		return processInfos[i].CPUPercent > processInfos[j].CPUPercent
	})

	if len(processInfos) > n {
		processInfos = processInfos[:n]
	}

	return processInfos, nil
}

func convertBytes(bytes uint64) (string, string) {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f", float64(bytes)/float64(GB)), "GB"
	case bytes >= MB:
		return fmt.Sprintf("%.2f", float64(bytes)/float64(MB)), "MB"
	case bytes >= KB:
		return fmt.Sprintf("%.2f", float64(bytes)/float64(KB)), "KB"
	default:
		return fmt.Sprintf("%d", bytes), "B"
	}
}
