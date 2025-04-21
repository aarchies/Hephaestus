package system

import (
	"unsafe"
)

const (
	BATCH_NOTIFY = 32      // 批量通知大小
	BUFFER_SIZE  = 1 << 20 // 1mb/缓冲区
	BUF_COUNT    = 1024    // 环形缓冲区数量
	CACHE_LINE   = 64      // 缓存行大小
	MEM_SIZE     = int(unsafe.Sizeof(ShmHeader{})) + BUF_COUNT*int(unsafe.Sizeof(ShmBuffer{}))
)

type ShmHeader struct {
	WriteIdx uint64 `json:"write_idx"`
	_        [CACHE_LINE - 8]byte
	ReadIdx  uint64 `json:"read_idx"`
	_        [CACHE_LINE - 8]byte
	Futex    int32 `json:"futex"`
	_        [CACHE_LINE - 4]byte
}

type ShmBuffer struct {
	ReqId  uint64            `json:"req_id"`
	Length uint32            `json:"length"`
	Data   [BUFFER_SIZE]byte `json:"data"`
}

func GetShmPointers(addr uintptr) (*ShmHeader, *[BUF_COUNT]ShmBuffer) {
	header := (*ShmHeader)(unsafe.Pointer(addr))
	buffers := (*[BUF_COUNT]ShmBuffer)(unsafe.Pointer(
		uintptr(unsafe.Pointer(header)) + unsafe.Sizeof(ShmHeader{}),
	))
	return header, buffers
}
