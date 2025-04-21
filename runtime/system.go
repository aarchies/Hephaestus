package runtime

import (
	"runtime"

	"github.com/shirou/gopsutil/mem"
)

const (
	GOOS = runtime.GOOS
)

var (
	ParentCpus  = make([]int, 0) // 父进程cpus
	IdleCpus    = make([]int, 0) // 空闲cpus
	TotalMemory = uint64(0)      // 总内存大小 kb
	NumCPU      = runtime.NumCPU // 总cpus
)

type pstatus uint32

const (
	_Pgcstop = iota
	_Pidle
	_Pruning
	_Psyscall
)

func init() {
	num := runtime.NumCPU()
	for i := 0; i < num; i++ {
		if num > 4 {
			if i < num/2-2 {
				ParentCpus = append(ParentCpus, i)
			} else {
				IdleCpus = append(IdleCpus, i)
			}
		} else {
			ParentCpus = append(ParentCpus, 0, 1)
			IdleCpus = append(IdleCpus, 2)
		}
	}
	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}
	TotalMemory = virtualMem.Total
}
