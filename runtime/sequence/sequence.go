package sequence

import (
	"encoding/binary"
	"hash"
	"math/bits"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/cespare/xxhash/v2"
	"github.com/ryszard/goskiplist/skiplist"
)

const (
	BufferSize      = 1 << 18              // 256K elements per shard
	FlushInterval   = 2 * time.Microsecond // Batching interval
	MaxBackPressure = 1 << 20              // 1MB buffer
	CacheLineSize   = 64
)

type (
	entry struct {
		seq  uint64
		data []byte
	}

	SeqShard struct {
		head      atomic.Uint64              // 8 bytes
		_         [CacheLineSize - 8]byte    // Padding
		tail      atomic.Uint64              // 8 bytes
		_         [CacheLineSize - 8]byte    // Padding
		buffer    [BufferSize]unsafe.Pointer // Lock-free ring buffer
		bitmap    [BufferSize / 64]uint64    // Presence bitmap
		entryPool sync.Pool                  // Per-shard object pool
	}

	SequenceManager struct {
		shards       []*SeqShard
		resultChan   chan []byte
		closeCh      chan struct{}
		expected     atomic.Uint64
		currentSeq   atomic.Uint64
		hasherPool   sync.Pool
		sortBuffer   *skiplist.SkipList
		commitMutex  sync.Mutex
		backPressure atomic.Int32
	}
)

func Uint64LessThan(a, b interface{}) bool {
	return a.(uint64) < b.(uint64)
}

func NewSequenceManager() *SequenceManager {
	shardCount := runtime.NumCPU() * 2
	sm := &SequenceManager{
		shards:     make([]*SeqShard, shardCount),
		resultChan: make(chan []byte, MaxBackPressure),
		closeCh:    make(chan struct{}),
		sortBuffer: skiplist.NewCustomMap(Uint64LessThan),
		hasherPool: sync.Pool{
			New: func() interface{} { return xxhash.New() },
		},
	}

	for i := range sm.shards {
		sm.shards[i] = &SeqShard{
			entryPool: sync.Pool{
				New: func() interface{} {
					return &entry{data: make([]byte, 0, 4096)}
				},
			},
		}
	}

	go sm.batchFlusher()
	return sm
}

func (sm *SequenceManager) Submit(seq uint64, data []byte) {
	shard := sm.shards[sm.getShardIndex(seq)]

	// 动态反压控制
	if float64(sm.backPressure.Load()) > float64(MaxBackPressure)*0.8 {
		time.Sleep(time.Duration(sm.backPressure.Load()) * time.Nanosecond)
	}

	for {
		tail := shard.tail.Load()
		head := shard.head.Load()

		if tail-head >= BufferSize {
			runtime.Gosched()
			continue
		}

		pos := tail % BufferSize
		if shard.tail.CompareAndSwap(tail, tail+1) {
			e := shard.entryPool.Get().(*entry)
			e.seq, e.data = seq, append(e.data[:0], data...)
			atomic.StorePointer(&shard.buffer[pos], unsafe.Pointer(e))
			setBit(shard, int(pos/64), 1<<(pos%64))
			return
		}
	}
}

func (sm *SequenceManager) batchFlusher() {
	ticker := time.NewTicker(FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.commitMutex.Lock()
			sm.flushShards()
			sm.commitOrdered()
			sm.commitMutex.Unlock()
		case <-sm.closeCh:
			sm.finalize()
			return
		}
	}
}

func (sm *SequenceManager) flushShards() {
	var wg sync.WaitGroup
	for _, shard := range sm.shards {
		wg.Add(1)
		go func(s *SeqShard) {
			defer wg.Done()
			positions := scanBitmap(s)
			for _, pos := range positions {
				entryPtr := atomic.SwapPointer(&s.buffer[pos], nil)
				if entryPtr != nil {
					e := (*entry)(entryPtr)
					sm.sortBuffer.Set(e.seq, e.data)
					s.entryPool.Put(e)
					sm.backPressure.Add(1) // 关键修复：增加背压计数
				}
				clearBit(s, int(pos/64), 1<<(pos%64))
			}
		}(shard)
	}
	wg.Wait()
}

func (sm *SequenceManager) commitOrdered() {
	current := sm.expected.Load()
	for {
		val, ok := sm.sortBuffer.Get(current)
		if !ok {
			break
		}

		select {
		case sm.resultChan <- val.([]byte):
			sm.sortBuffer.Delete(current)
			current++
			sm.backPressure.Add(-1)
			sm.expected.Store(current) // 确保及时更新expected
		default:
			return
		}
	}
	sm.currentSeq.Store(current - 1)
}

func (sm *SequenceManager) finalize() {
	close(sm.resultChan)
	for _, shard := range sm.shards {
		for i := 0; i < BufferSize; i++ {
			if ptr := atomic.LoadPointer(&shard.buffer[i]); ptr != nil {
				shard.entryPool.Put((*entry)(ptr))
			}
		}
	}
}

func setBit(shard *SeqShard, word int, bit uint64) {
	for {
		old := atomic.LoadUint64(&shard.bitmap[word])
		new := old | bit
		if atomic.CompareAndSwapUint64(&shard.bitmap[word], old, new) {
			return
		}
	}
}

func clearBit(shard *SeqShard, word int, bit uint64) {
	for {
		old := atomic.LoadUint64(&shard.bitmap[word])
		new := old & ^bit
		if atomic.CompareAndSwapUint64(&shard.bitmap[word], old, new) {
			return
		}
	}
}

func scanBitmap(shard *SeqShard) []uint64 {
	var positions []uint64
	for word := 0; word < len(shard.bitmap); word++ {
		bitmap := atomic.LoadUint64(&shard.bitmap[word])
		for bitmap != 0 {
			bit := bitmap & -bitmap
			pos := uint64(word*64 + bits.TrailingZeros64(bit))
			if pos < BufferSize {
				positions = append(positions, pos)
			}
			bitmap ^= bit
		}
	}
	return positions
}

func (sm *SequenceManager) getShardIndex(seq uint64) int {
	h := sm.hasherPool.Get().(hash.Hash64)
	defer sm.hasherPool.Put(h)
	h.Reset()

	binary.Write(h, binary.LittleEndian, seq)
	return int(h.Sum64() % uint64(len(sm.shards)))
}

func (sm *SequenceManager) Close() {
	close(sm.closeCh)
}

func (sm *SequenceManager) OutputChan() chan []byte {
	return sm.resultChan
}

func (sm *SequenceManager) CurrentSequence() uint64 {
	return sm.currentSeq.Load()
}
