//go:build !linux
// +build !linux

package ipc

import (
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	// 内存映射相关API
	procCreateFileMappingW = kernel32.NewProc("CreateFileMappingW")
	procOpenFileMappingW   = kernel32.NewProc("OpenFileMappingW")
	procMapViewOfFile      = kernel32.NewProc("MapViewOfFile")
	procUnmapViewOfFile    = kernel32.NewProc("UnmapViewOfFile")

	// 事件对象相关API
	procCreateEventW        = kernel32.NewProc("CreateEventW")
	procOpenEventW          = kernel32.NewProc("OpenEventW")
	procSetEvent            = kernel32.NewProc("SetEvent")
	procResetEvent          = kernel32.NewProc("ResetEvent")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")

	// 其他工具API
	procCloseHandle = kernel32.NewProc("CloseHandle")
)

const (
	FILE_MAP_ALL_ACCESS = 0xF001F
	PAGE_READWRITE      = 0x04
	EVENT_ALL_ACCESS    = 0x1F0003
	INFINITE            = 0xFFFFFFFF
)

// 创建共享内存对象
func CreateFileMapping(name string, size int) (syscall.Handle, error) {
	namePtr, _ := syscall.UTF16PtrFromString(name)
	ret, _, err := procCreateFileMappingW.Call(
		uintptr(syscall.InvalidHandle),
		uintptr(0),
		uintptr(PAGE_READWRITE),
		uintptr(size>>32),        // 高32位大小
		uintptr(size&0xFFFFFFFF), // 低32位大小
		uintptr(unsafe.Pointer(namePtr)),
	)
	if err != nil {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

// 打开已存在的共享内存
func OpenFileMapping(name string) (syscall.Handle, error) {
	namePtr, _ := syscall.UTF16PtrFromString(name)
	ret, _, err := procOpenFileMappingW.Call(
		uintptr(FILE_MAP_ALL_ACCESS),
		uintptr(0), // 不继承句柄
		uintptr(unsafe.Pointer(namePtr)),
	)
	if err != nil {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

// 映射内存视图
func MapViewOfFile(handle syscall.Handle, size int) uintptr {
	addr, _, _ := procMapViewOfFile.Call(
		uintptr(handle),
		uintptr(FILE_MAP_ALL_ACCESS),
		uintptr(0),
		uintptr(0),
		uintptr(size),
	)
	return addr
}

// 取消内存映射
func UnmapViewOfFile(addr uintptr) {
	procUnmapViewOfFile.Call(addr)
}

// 创建事件对象
func CreateEvent(name string) syscall.Handle {
	namePtr, _ := syscall.UTF16PtrFromString(name)
	ret, _, _ := procCreateEventW.Call(
		uintptr(0), // 默认安全属性
		uintptr(0), // 自动重置模式
		uintptr(0), // 初始非信号状态
		uintptr(unsafe.Pointer(namePtr)),
	)
	return syscall.Handle(ret)
}

// 打开已存在的事件对象
func OpenEvent(name string) syscall.Handle {
	namePtr, _ := syscall.UTF16PtrFromString(name)
	ret, _, _ := procOpenEventW.Call(
		uintptr(EVENT_ALL_ACCESS),
		uintptr(0), // 不继承句柄
		uintptr(unsafe.Pointer(namePtr)),
	)
	return syscall.Handle(ret)
}

// 设置事件为信号状态
func SetEvent(handle syscall.Handle) {
	procSetEvent.Call(uintptr(handle))
}

// 重置事件状态
func ResetEvent(handle syscall.Handle) {
	procResetEvent.Call(uintptr(handle))
}

// 等待对象信号
func Wait(handle syscall.Handle, timeout uint32) uint32 {
	ret, _, _ := procWaitForSingleObject.Call(
		uintptr(handle),
		uintptr(timeout),
	)
	return uint32(ret)
}

// 关闭内核对象句柄
func CloseHandle(handle syscall.Handle) {
	procCloseHandle.Call(uintptr(handle))
}

// 获取错误信息
func GetLastError() error {
	return syscall.GetLastError()
}

// 转换LPCWSTR字符串
func StringToUTF16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}
