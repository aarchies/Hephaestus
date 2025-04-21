//go:build linux
// +build linux

package ipc

import (
	"os"
	"syscall"

	"github.com/aarchies/hephaestus/system/shm"

	"golang.org/x/sys/unix"
)

// 打开共享内存
func Shm_Open(name string, flag int, perm os.FileMode) (int, error) {
	fd, err := shm.Open(name, flag, perm)
	if err != nil {
		return 0, err
	}
	return fd, nil
}

// 取消内存共享
func Shm_Unlink(name string) error {
	if err := shm.Unlink(name); err != nil {
		return err
	}
	return nil
}

// 设置内存大小
func Ftruncate(fd int, size int64) error {
	if err := unix.Ftruncate(fd, size); err != nil {
		return err
	}
	return nil
}

// 映射内存片
func Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	addr, err := unix.Mmap(fd, offset, length, prot, flags)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// 取消映射
func Munmap(addr []byte) error {
	if err := unix.Munmap(addr); err != nil {
		return err
	}
	return nil
}

// 内存预分配
func Madvise(b []byte, advice int) error {
	if err := unix.Madvise(b, advice); err != nil {
		return err
	}
	return nil
}

// 让出执行权
func SchedYield() error {
	_, _, errno := syscall.Syscall(syscall.SYS_SCHED_YIELD, 0, 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
