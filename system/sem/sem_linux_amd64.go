// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs sem_linux.go

package sem

import (
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/unix"
)

const MAX_RETRY = 3

type Semaphore [32]byte

type newSem struct {
	Value    uint32
	Private  int32
	NWaiters uint64
}

func New(value uint) (*Semaphore, error) {
	sem := new(Semaphore)

	if err := sem.initialize(value); err != nil {
		return nil, err
	}

	return sem, nil
}

func atomicDecrementIfPositive(mem *uint32) uint32 {
	for {
		if old := atomic.LoadUint32(mem); old == 0 || atomic.CompareAndSwapUint32(mem, old, old-1) {
			return old
		}
	}
}

func (sem *Semaphore) Wait() error {
	isem := (*newSem)(unsafe.Pointer(sem))

	if atomicDecrementIfPositive(&isem.Value) > 0 {
		return nil
	}

	atomic.AddUint64(&isem.NWaiters, 1)
 
	for {

		if _, _, err := unix.Syscall6(unix.SYS_FUTEX, uintptr(unsafe.Pointer(&isem.Value)), uintptr(0x0), 0, 0, 0, 0); err != 0 && err != unix.EWOULDBLOCK {
			atomic.AddUint64(&isem.NWaiters, ^uint64(0))
			return err
		}

		if atomicDecrementIfPositive(&isem.Value) > 0 {
			atomic.AddUint64(&isem.NWaiters, ^uint64(0))
			return nil
		}
	}
}

func (sem *Semaphore) WaitwithTime(time int64) error {
	isem := (*newSem)(unsafe.Pointer(sem))

	if atomicDecrementIfPositive(&isem.Value) > 0 {
		return nil
	}

	atomic.AddUint64(&isem.NWaiters, 1)

	// 1s 
	timeout := unix.Timespec{
        Sec:  time,
        Nsec: 0,
    }

	for {

		if _, _, err := unix.Syscall6(unix.SYS_FUTEX, uintptr(unsafe.Pointer(&isem.Value)), uintptr(0x0), 0, uintptr(unsafe.Pointer(&timeout)), 0, 0); err != 0 && err != unix.EWOULDBLOCK {
			atomic.AddUint64(&isem.NWaiters, ^uint64(0))
			return err
		}

		if atomicDecrementIfPositive(&isem.Value) > 0 {
			atomic.AddUint64(&isem.NWaiters, ^uint64(0))
			return nil
		}
	}
}

// todo error
func (sem *Semaphore) TryWait() error {
	isem := (*newSem)(unsafe.Pointer(sem))

	if atomicDecrementIfPositive(&isem.Value) > 0 {
		return nil
	}

	return unix.EAGAIN
}


func (s *Semaphore) PostWithRetry(retry int) {
    for i := 0; i < retry; i++ {
        if err := s.Post(); err == nil {
            return
        }
    }
}

func (sem *Semaphore) Post() error {
	isem := (*newSem)(unsafe.Pointer(sem))

	for {
		cur := atomic.LoadUint32(&isem.Value)

		if cur == 0x7fffffff {
			return unix.EOVERFLOW
		}

		if atomic.CompareAndSwapUint32(&isem.Value, cur, cur+1) {
			break
		}
	}

	if atomic.LoadUint64(&isem.NWaiters) <= 0 {
		return nil
	}

	if _, _, err := unix.Syscall6(unix.SYS_FUTEX, uintptr(unsafe.Pointer(&isem.Value)), uintptr(0x1), 1, 0, 0, 0); err != 0 {
		return err
	}

	return nil
}

func (sem *Semaphore) initialize(value uint) error {
	if value > 0x7fffffff {
		return unix.EINVAL
	}

	isem := (*newSem)(unsafe.Pointer(sem))
	isem.Value = uint32(value)
	isem.Private = 0
	isem.NWaiters = 0

	return nil
}

func (sem *Semaphore) Destroy() error {
	isem := (*newSem)(unsafe.Pointer(sem))

	// 唤醒当前睡眠中线程
	if _, _, err := unix.Syscall6(unix.SYS_FUTEX, uintptr(unsafe.Pointer(&isem.Value)), uintptr(0x1), 1, 0, 0, 0); err != 0 {
		return err
	}
	sem = nil
	return nil
}
