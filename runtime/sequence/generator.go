package sequence

import "sync/atomic"

type SequenceGenerator struct {
	seq uint64
}

func (sg *SequenceGenerator) Next() uint64 {
	return atomic.AddUint64(&sg.seq, 1)
}
