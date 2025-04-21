//go:build linux
// +build linux

package cpu_test

import (
	"golang.org/x/sys/unix"
)

// 绑定核心至pid
func BindCPU(pid int, core int) error {
	var cpuSet unix.CPUSet
	cpuSet.Set(core)
	return unix.SchedSetaffinity(pid, &cpuSet)
}
