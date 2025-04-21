//go:build linux
// +build linux

package actor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/aarchies/hephaestus/system"
	"github.com/aarchies/hephaestus/system/ipc"
	"github.com/aarchies/hephaestus/system/sem"
	"github.com/aarchies/hephaestus/system/shm"

	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

type (
	seqCh struct {
		SeqId  uint64
		Buffer []byte
	}
	muintptr struct {
		addr []byte
		shm  string
	}
	Actuators struct {
		id       int32
		cmd      *exec.Cmd
		status   Status
		sigCh    chan os.Signal
		inCh     chan seqCh
		outCh    chan seqCh
		reqMmap  muintptr
		respMmap muintptr
		ctx      context.Context
		cancel   context.CancelFunc
	}
)

func NewWorker(id int32) *Actuators {
	ctx, cancel := context.WithCancel(context.Background())
	return &Actuators{
		id:     id,
		sigCh:  make(chan os.Signal, 1),
		inCh:   make(chan seqCh),
		outCh:  make(chan seqCh),
		status: Padding,
		cmd:    &exec.Cmd{},
		ctx:    ctx,
		cancel: cancel,
	}
}

func initShm(name string) ([]byte, error) {
	// 创建/打开共享内存
	fd, err := shm.Open(name, unix.O_CREAT|unix.O_RDWR|unix.O_EXCL, 0666)
	if err != nil {
		return nil, err
	}

	// 设置共享内存大小
	if err := unix.Ftruncate(fd, int64(system.MEM_SIZE)); err != nil {
		return nil, err
	}

	// 内存映射
	addr, err := unix.Mmap(fd, 0, system.MEM_SIZE, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED|unix.MAP_LOCKED)
	if err != nil {
		return nil, err
	}

	// 内存预分配对齐,防止数据量徒增引起缺页中断，和数据延时徒增
	if err := unix.Madvise(addr, unix.MADV_HUGEPAGE); err != nil {
		return nil, err
	}
	return addr, nil
}

func (c *Actuators) Run() (err error) {

	// args
	c.reqMmap.shm = fmt.Sprintf("%s_%d", "/req_perf_shm", c.id)
	c.respMmap.shm = fmt.Sprintf("%s_%d", "/resp_perf_shm", c.id)

	// init req shm
	c.reqMmap.addr, err = initShm(c.reqMmap.shm)
	if err != nil {
		return err
	}

	// init resp shm
	c.respMmap.addr, err = initShm(c.respMmap.shm)
	if err != nil {
		return err
	}

	// child process
	c.cmd = exec.Command("go", "run", "./cmd/worker", "-mem", c.reqMmap.shm, "-res_mem", c.respMmap.shm)
	c.cmd.Dir = "./"
	c.cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH"))

	c.cmd.Stdout = os.Stdout
	c.cmd.Stderr = os.Stderr

	if err = c.cmd.Start(); err != nil {
		shm.Unlink(c.reqMmap.shm)
		unix.Munmap(c.reqMmap.addr)
		shm.Unlink(c.respMmap.shm)
		unix.Munmap(c.reqMmap.addr)

		c.status = Error
		return err
	}

	go c.callBack()
	go c.listener()
	go func() {
		err := c.cmd.Wait()
		if err != nil {
			fmt.Printf("Process exited with error: %v\n", err)
		} else {
			fmt.Println("Process exited successfully")
		}
		c.status = Exited
	}()
	time.Sleep(time.Second * 2)
	c.status = Ready
	return nil
}

func (c *Actuators) listener() {
	c.status = Start
	header, buffers := system.GetShmPointers(uintptr(unsafe.Pointer(&c.reqMmap.addr[0])))
	header.WriteIdx = 0
	header.ReadIdx = 0
	futex := (*sem.Semaphore)(unsafe.Pointer(&header.Futex))

	for {
		select {
		case <-c.sigCh:
			shm.Unlink(c.reqMmap.shm)
			unix.Munmap(c.reqMmap.addr)
			return
		case data := <-c.inCh:
			var w uint64
			for {

				r := atomic.LoadUint64(&header.ReadIdx)
				w = atomic.LoadUint64(&header.WriteIdx)

				if w-r >= system.BUF_COUNT {
					ipc.SchedYield()
					continue
				}
				break
			}

			idx := header.WriteIdx % system.BUF_COUNT

			buffers[idx].ReqId = data.SeqId
			buffers[idx].Length = uint32(len(data.Buffer))
			copy(buffers[idx].Data[:], data.Buffer)

			atomic.CompareAndSwapUint64(&header.WriteIdx, w, w+1)
			futex.PostWithRetry(3)
		}
	}
}

func (c *Actuators) callBack() {
	header, buffers := system.GetShmPointers(uintptr(unsafe.Pointer(&c.respMmap.addr[0])))
	header.WriteIdx = 0
	header.ReadIdx = 0
	futex := (*sem.Semaphore)(unsafe.Pointer(&header.Futex))

	for {
		select {
		case <-c.ctx.Done():
			shm.Unlink(c.respMmap.shm)
			unix.Munmap(c.respMmap.addr)
			return
		default:
			futex.WaitwithTime(1)

			r := atomic.LoadUint64(&header.ReadIdx)
			w := atomic.LoadUint64(&header.WriteIdx)

			if r >= w {
				ipc.SchedYield()
				continue
			}

			idx := r % system.BUF_COUNT

			len := buffers[idx].Length

			c.outCh <- seqCh{
				SeqId:  buffers[idx].ReqId,
				Buffer: buffers[idx].Data[:len],
			}
			atomic.CompareAndSwapUint64(&header.ReadIdx, r, r+1)
		}
	}
}

func (c *Actuators) Wait() {
	c.cmd.Wait()
}

func (c *Actuators) Exit() {

	if c == nil || c.cmd == nil {
		return
	}

	c.signal()

	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		fmt.Println(err)
	}

	c.cmd = nil
	c.status = Exited
}

func (c *Actuators) signal() {
	c.sigCh <- os.Interrupt
	c.cancel()
}

func (c *Actuators) Input(seq uint64, buffer []byte) {
	c.inCh <- seqCh{
		SeqId:  seq,
		Buffer: buffer,
	}
}

func (c *Actuators) Output() chan seqCh {
	return c.outCh
}

func (c *Actuators) Pid() int {
	if c.cmd != nil {
		return c.cmd.Process.Pid
	}
	return 0
}

func (c *Actuators) Status() Status {
	return c.status
}
