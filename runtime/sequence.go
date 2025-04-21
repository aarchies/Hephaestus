package runtime

import (
	"encoding/binary"
	"hash/fnv"

	"sync"
	"sync/atomic"
	"time"
)

const (
	SeqShardCount = 12 * 4 // 分片数，降低锁竞争
	SeqWindowSize = 4096   // 最大允许的序列号窗口
)

type SeqShard struct {
	mu     sync.Mutex
	buffer map[uint64][]byte // 当前分片内的序列缓存
	maxSeq uint64            // 当前分片最大接收序列号
}

type SequenceManager struct {
	shards      [SeqShardCount]*SeqShard // 分片数组
	currentSeq  atomic.Uint64            // 全局当前已提交序列号
	resultChan  chan []byte              // 有序结果通道
	flushTicker *time.Ticker             // 定时刷新
	closeCh     chan struct{}
}

func NewSequenceManager() *SequenceManager {
	sm := &SequenceManager{
		resultChan:  make(chan []byte, 10240), // 大缓冲减少阻塞
		flushTicker: time.NewTicker(10 * time.Millisecond),
		closeCh:     make(chan struct{}),
	}

	// 初始化分片
	for i := range sm.shards {
		sm.shards[i] = &SeqShard{
			buffer: make(map[uint64][]byte, SeqWindowSize/SeqShardCount),
		}
	}

	go sm.backgroundFlush()
	return sm
}

// 按序列号分片提交
func (sm *SequenceManager) Submit(seq uint64, data []byte) {

	// 检查反压
	if sm.backpressure() {
		time.Sleep(10 * time.Microsecond)
	}

	shard := sm.shards[getShardIndex(seq)]

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// 淘汰过期序列（防内存泄漏）
	if seq < sm.currentSeq.Load() {
		return
	}

	shard.buffer[seq] = data
	if seq > shard.maxSeq {
		shard.maxSeq = seq
	}
}

// 当结果通道超过90%容量时触发反压
func (sm *SequenceManager) backpressure() bool {
	return len(sm.resultChan) > cap(sm.resultChan)*9/10
}

func (sm *SequenceManager) backgroundFlush() {
	for {
		select {
		case <-sm.flushTicker.C:
			sm.tryFlush()
		case <-sm.closeCh:
			return
		}
	}
}

func (sm *SequenceManager) tryFlush() {
	expected := sm.currentSeq.Load() + 1

	// 遍历所有分片寻找连续序列
	for {
		found := false

		// 并行检查分片
		for _, shard := range sm.shards {
			shard.mu.Lock()

			if data, exists := shard.buffer[expected]; exists {
				select {
				case sm.resultChan <- data:
					delete(shard.buffer, expected)
					expected++
					found = true
				default: // 通道满则停止本轮提交
					shard.mu.Unlock()
					return
				}
			}

			shard.mu.Unlock()
		}

		if !found {
			break
		}
	}

	sm.currentSeq.Store(expected - 1)
}

func (sm *SequenceManager) Close() {
	sm.closeCh <- struct{}{}
}

// 哈希分片，减少伪共享
func getShardIndex(seq uint64) int {
	h := fnv.New32a()
	h.Write(binary.LittleEndian.AppendUint64(nil, seq))
	return int(h.Sum32() % SeqShardCount)
}

// func getShardIndex1(seq uint64) int {
// 	h := xxhash.New()
// 	h.Write(binary.LittleEndian.AppendUint64(nil, seq))
// 	return int(h.Sum64() % SeqShardCount)
// }
