package ipc

import "unsafe"

const (
	BATCH_NOTIFY = 32      // 批量通知大小
	BUFFER_SIZE  = 1 << 10 // 1kb/缓冲区
	BUF_COUNT    = 1024    // 环形缓冲区数量
	CACHE_LINE   = 64      // 缓存行大小
	Timeout      = 500     // 毫秒
	MEM_SIZE     = int(unsafe.Sizeof(ShmHeader{})) + BUF_COUNT*int(unsafe.Sizeof(ShmBuffer{}))
)

const (
	PROT_READ     = 0x1
	PROT_WRITE    = 0x2
	MAP_SHARED    = 0x1
	MAP_LOCKED    = 0x2000
	MADV_HUGEPAGE = 0xe
)

// 请求内存结构
type ShmHeader struct {
	WriteIdx uint64 `json:"write_idx"`
	_        [CACHE_LINE - 8]byte
	ReadIdx  uint64 `json:"read_idx"`
	_        [CACHE_LINE - 8]byte
	Futex    [32]byte `json:"futex"`
	_        [CACHE_LINE - 32]byte
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
