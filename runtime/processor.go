//go:build linux
// +build linux

package runtime

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	maxSize     = 0
	ctx, cancel = context.WithCancel(context.Background())
)

const (
	cacheLineSize = 64
	maxQueueSize  = 1 << 20 // 1048576 个元素（调整为2的幂次）
)

type puintptr *P
type (
	P struct {
		id      int32
		status  pstatus  // p 的状态 pidle/prunning/...
		m       muintptr // p 所属的 machine
		runq    *gQueue  // g队列
		max     int      // 最大数量
		preempt bool     // 抢占信号
	}
	gQueue struct {
		queue   *ringQueue
		runnext guintptr
		epoch   uint64
	}
	ringQueue struct {
		head   uint64                  // 头指针（独占缓存行）
		_      [cacheLineSize - 8]byte // 补齐
		tail   uint64                  // 尾指针（独占缓存行）
		_      [cacheLineSize - 8]byte // 补齐
		buffer [maxQueueSize]guintptr  // 缓冲区
		mask   uint64                  // buffer len-1
	}
)

func procresize(newSize int) {
	if newSize > 0 {
		maxSize = newSize
	} else {
		maxSize = maxQueueSize
	}

	// // 使用原子操作更新所有P的max值
	// allp := sched.allp
	// for _, p := range allp {
	// 	atomic.StoreInt32((*int32)(unsafe.Pointer(&p.max)), int32(newSize))
	// }
}

func newGQueue() *gQueue {
	q := &gQueue{
		queue: &ringQueue{
			mask: uint64(maxQueueSize - 1),
		},
	}
	atomic.StoreUint64(&q.queue.head, 0)
	atomic.StoreUint64(&q.queue.tail, 0)
	atomic.StoreUint64(&q.epoch, 0)
	return q
}

func newp(id int32) puintptr {
	p := &P{
		id:      id,
		status:  _Pgcstop,
		runq:    newGQueue(),
		max:     maxSize,
		preempt: false,
	}

	go p.listener()
	return p
}

func (p *gQueue) join(g guintptr) bool {
	if atomic.CompareAndSwapPointer((*unsafe.Pointer)(&p.runnext), nil, unsafe.Pointer(g)) {
		return true
	}

	q := p.queue
	for {
		tail := atomic.LoadUint64(&q.tail)
		head := atomic.LoadUint64(&q.head)

		if tail-head >= uint64(maxQueueSize) {
			return false
		}

		if atomic.CompareAndSwapUint64(&q.tail, tail, tail+1) {
			// 使用CAS写入数据
			slot := &q.buffer[tail&q.mask] // 位运算 等价%size
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(slot), nil, unsafe.Pointer(g)) {
				return true
			}
			// 写入失败回滚tail
			atomic.AddUint64(&q.tail, ^uint64(0)) //取反 等价tail--
		}
	}
}

func (p *gQueue) next() guintptr {
	if g := atomic.SwapPointer((*unsafe.Pointer)(&p.runnext), nil); g != nil {
		return guintptr(g)
	}

	q := p.queue
	for {
		head := atomic.LoadUint64(&q.head)
		tail := atomic.LoadUint64(&q.tail)

		if head >= tail {
			atomic.StoreUint64(&p.epoch, 1)
			return nil
		}

		slot := &q.buffer[head&q.mask]
		g := atomic.LoadPointer((*unsafe.Pointer)(slot))
		if g == nil {
			return nil
		}

		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(slot), g, nil) {
			atomic.StoreUint64(&q.head, head+1)
			p.prefetch(uintptr(unsafe.Pointer(&q.buffer[(head+1)&q.mask])))
			return guintptr(g)
		}
	}
}

func (p *gQueue) prefetch(addr uintptr) {
	_ = (*[cacheLineSize]byte)(unsafe.Pointer(addr))[0]
}

func (p *gQueue) fetallp() guintptr {
	if sched.runqsize == 0 {
		return nil
	}

	if g := sched.runq.next(); g != nil {
		sched.runqsize--
		return g
	}
	return nil
}

func (p *gQueue) plagiarize() guintptr {

	if sched.npidle > 0 {
		for _, v := range sched.pidle {
			if v.runq.len() > 0 {
				return v.runq.next()
			}
		}
	}
	return nil
}

func (p *P) listener() {
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()

	reset := time.NewTicker(time.Second * 30)
	defer reset.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if p.runq.len() < p.max && p.status != _Pidle {
				p.status = _Pidle
				sched.change(p, _Pidle)
			} else if p.runq.len() >= p.max*8/10 && p.status != _Pruning {
				p.status = _Pruning
				sched.change(p, _Pruning)
			}
		case <-reset.C:
			epoch := atomic.LoadUint64(&p.runq.epoch)
			if epoch == 1 && p.m.spinning {
				atomic.StoreUint64(&p.runq.queue.head, 0)
				atomic.StoreUint64(&p.runq.queue.tail, 0)
				atomic.StoreUint64(&p.runq.epoch, 0)
				runtime.GC()
			}
		}
	}
}

func (q *gQueue) len() int {
	qPtr := q.queue
	tail := atomic.LoadUint64(&qPtr.tail)
	head := atomic.LoadUint64(&qPtr.head)
	return int(tail - head)
}

func (q *gQueue) isBottleneck() bool {
	qPtr := q.queue
	tail := atomic.LoadUint64(&qPtr.tail)
	head := atomic.LoadUint64(&qPtr.head)
	return tail-head >= uint64(maxQueueSize)
}

func (p *P) cleanup() {
	cancel()
}
