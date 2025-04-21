//go:build !linux
// +build !linux

package cpu_test

// 绑定核心至pid
func BindCPU(pid int, core int) error {

	return nil
}
