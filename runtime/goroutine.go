package runtime

import (
	"time"
	"unsafe"
)

type guintptr unsafe.Pointer
type G struct {
	id         uint64 // Id
	buffer     []byte // 缓冲区
	isComplted bool   // 是否已完成
	time       int64  // 创建时间
}

// newg creates a new goroutine.
func newg(id uint64, buffer []byte) guintptr {
	g := &G{
		id:     id,
		buffer: buffer,
		time:   time.Now().Unix(),
	}

	return guintptr(unsafe.Pointer(g))
}

// 防止指针逃逸的hack
//
//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func releaseG(g guintptr) *G {
	// 使用uintptr转换避免GC问题
	return (*G)(noescape(unsafe.Pointer(g)))
}
