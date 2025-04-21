package procs

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

func CPUPercent(pid int32) (float64, error) {
	if pid == 0 {
		cpuPercent, err := cpu.Percent(time.Second, false)
		if err != nil {
			return 0, err
		}
		return cpuPercent[0], nil
	} else {

		proc, err := process.NewProcess(pid)
		if err != nil {
			return 0, err
		}
		cpuPercent, err := proc.CPUPercent()
		if err != nil {
			return 0, err
		}
		return cpuPercent, nil
	}
}

func MemoryInfo(pid int32) (uint64, float64, error) {
	if pid == 0 {
		virtualMem, err := mem.VirtualMemory()
		if err != nil {
			return 0, 0, err
		}
		return virtualMem.Used, virtualMem.UsedPercent, nil
	} else {
		proc, err := process.NewProcess(pid)
		if err != nil {
			return 0, 0, err
		}

		virtualMem, err := mem.VirtualMemory()
		if err != nil {
			return 0, 0, err
		}
		totalMemory := virtualMem.Total

		memInfo, err := proc.MemoryInfo()
		if err != nil {
			return 0, 0, err
		}
		memPercent := (float64(memInfo.RSS) / float64(totalMemory)) * 100
		return memInfo.RSS, memPercent, nil
	}
}
