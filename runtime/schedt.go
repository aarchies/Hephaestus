//go:build linux
// +build linux

package runtime

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/aarchies/hephaestus/runtime/sequence"

	"github.com/sirupsen/logrus"
)

var (
	sched = &Sched{
		mu:         sync.Mutex{},
		runq:       newGQueue(),
		seq:        &sequence.SequenceGenerator{},
		sm:         sequence.NewSequenceManager(),
		dispatcher: &Dispatcher{},
		allm:       make(map[int]muintptr, 0),
		allp:       make([]puintptr, 0),
		allmt:      make([]*Metrics, 0),
		pidle:      make([]puintptr, 0),
	}
	mHealthCheckInterval = 10
)

const (
	scaleUpCPULimit   = 70.0 // CPU使用率扩容阈值
	scaleDownMemLimit = 25.0 // 内存使用率缩容阈值
	maxPendingAlert   = 1000 // 任务积压告警阈值
)

type Sched struct {
	seq        *sequence.SequenceGenerator // seq index
	sm         *sequence.SequenceManager   // seq manager
	index      int                         // roundRobin index
	strategy   string                      // p pool balance strategy
	m          muintptr                    // global M
	next       muintptr                    // next m
	mnext      int32                       // next m id
	allm       map[int]muintptr            // m pool,key is pid
	allp       []puintptr                  // p pool
	allmt      []*Metrics                  // metrics pool
	runq       *gQueue                     // global runq g queue
	runqsize   int32                       // global runq g count
	pidle      []puintptr                  // global idle p array
	npidle     uint32                      // global idle p count
	maxM       int                         // max m count
	maxP       int                         // max p count
	dispatcher *Dispatcher                 // global dispatcher
	ctx        context.Context             // sched global ctx
	mu         sync.Mutex                  // futex
}

func Init(ctx context.Context) {

	m := mcommoninit()

	sched.mu.Lock()
	sched.maxM = len(IdleCpus) / 2
	sched.maxP = len(IdleCpus) / 2
	sched.ctx = ctx
	sched.m = m
	sched.mnext++
	sched.allm[m.metrics.PID] = m
	sched.allp = append(sched.allp, m.p)
	sched.allmt = append(sched.allmt, m.metrics)
	sched.m.alllink = unsafe.Pointer(&sched.allm)
	sched.dispatcher.workers.Store(sched.allmt)
	sched.mu.Unlock()

	//sched.scaleUp()
	procresize(0)

	go sched.g_runb()
}

func (s *Sched) autoScale() {
	if s.shouldScaleUp() {
		s.scaleUp()
	}
	if s.shouldScaleDown() {
		s.scaleDown()
	}
}

func (s *Sched) shouldScaleUp() bool {
	return s.m.metrics.CPUUsage > scaleUpCPULimit ||
		s.m.metrics.Pending > maxPendingAlert ||
		s.m.metrics.MemPercent > 80.0
}

func (s *Sched) shouldScaleDown() bool {
	return s.m.metrics.CPUUsage < 30.0 &&
		s.m.metrics.Pending < 50 &&
		s.m.metrics.MemPercent < scaleDownMemLimit
}

func (s *Sched) scaleUp() {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Println("scaleUp")
	// cpu占用高且内存占用低
	// 1.增加p队列数量

	// 内存占用高且队列占用高
	// 1.增加m数量

	// 队列占用高且cpu占用高
	//
	// 1. 优先尝试增加P数量
	// for _, m := range s.allm {
	// 	if m.p == nil || len(m.p.runq.b) > m.p.max/2 {
	// 		newP := c.createP(m)
	// 		m.bindP(newP)
	// 		return
	// 	}
	// }

	// 2. 没有可扩展P则创建新M
	if len(s.allm) < s.maxM {
		m := newm(s.mnext)
		s.mnext++
		sched.allm[m.metrics.PID] = m
		sched.allp = append(sched.allp, m.p)
		sched.allmt = append(sched.allmt, m.metrics)
		return
	}

	// 3. 达到最大M数则调整P配置
	for _, m := range s.allm {
		m.p.max = min(m.p.max+10000, m.p.max)
	}
}

func (s *Sched) scaleDown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 1. 优先减少P数量
	// for i := len(c.activeM) - 1; i >= 0; i-- {
	// 	m := c.activeM[i]
	// 	if m.p != nil && m.p.runq.Len() < m.p.max/3 {
	// 		m.releaseP()
	// 		return
	// 	}
	// }

	// 2. 移除空闲M
	if len(s.allm) > 1 {
		fmt.Println("scaleDown")
		// 获取msater m以外其他空闲m
		for i, v := range s.pidle {
			if v.id != 0 {
				for _, d := range s.allm {
					if d.id == v.id {
						(*d).Cease()
						s.mnext--
						delete(s.allm, d.metrics.PID)
						s.pidle = append(s.pidle[:i], s.pidle[i+1:]...)
						s.npidle--
						s.allp = append(s.allp[:i], s.allp[i+1:]...)
						s.allmt = append(s.allmt[:i], s.allmt[i+1:]...)
					}
				}
			}
		}
	}
}

func WeightPut(bytes []byte) error {
	g := newg(sched.seq.Next(), bytes)

	if sched.next == nil {
		worker := sched.dispatcher.SelectWorker()
		m := sched.allm[worker.PID]
		sched.next = m
	}

	if p := (*sched.next.p); p.runq.join(g) {
		(*p.m).wakeup()
	} else {
		sched.allot(g)
	}

	return nil
}

func LoadBalancePut(bytes []byte) error {

	g := newg(sched.seq.Next(), bytes)

	p := sched.balanced()
	if p != nil {
		(*p).runq.join(g)
	} else {
		sched.allot(g)
	}

	return nil
}

func (s *Sched) allot(g guintptr) {

	// 获取当前所有空闲队列
	for _, p := range s.pidle {
		if !p.runq.isBottleneck() {
			if (*p).runq.join(g) {
				(*p.m).wakeup()
			}
		} else {
			s.change(p, _Pruning)
		}
	}

	// 存入全局队列
	if sched.runq.join(g) {
		sched.runqsize++
	}
}

func (s *Sched) g_runb() {
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	interval := time.NewTicker(time.Duration(mHealthCheckInterval) * time.Second)
	defer interval.Stop()

	for {
		select {
		case <-s.ctx.Done():
			for _, v := range s.allm {
				(*v).Cease()
			}
			s.sm.Close()
			logrus.Infoln("stoping core runtime")
			os.Exit(0)
			return
		case <-t.C:
			newWeights := calculateWeights(s.allmt) // 计算新权重
			s.dispatcher.UpdateWeights(newWeights)  // 平滑更新

			// 缓存权重结果
			worker := sched.dispatcher.SelectWorker()
			s.mu.Lock()
			m := s.allm[worker.PID]
			s.next = m
			s.mu.Unlock()
		case <-interval.C:
			// s.autoScale()
			// mHealthCheckInterval *= 2
			// interval.Reset(time.Duration(mHealthCheckInterval))
		}
	}
}

func (s *Sched) balanced() puintptr {

	switch s.strategy {
	case "least_conn":
		return s.least()
	case "random":
		return s.random()
	default:
		return s.roundRobin()
	}
}

func (s *Sched) least() puintptr {

	if len(s.pidle) == 0 {
		return nil
	}

	var count = s.pidle[0].runq.len()
	var p puintptr
	for _, v := range s.pidle {
		c := v.runq.len()
		if c <= count {
			count = c
			p = v
		}
	}
	return p
}

func (s *Sched) random() puintptr {

	if len(s.pidle) == 0 {
		return nil
	}

	return s.pidle[rand.Intn(len(s.pidle))]
}

func (s *Sched) roundRobin() puintptr {

	if len(s.pidle) == 0 {
		return nil
	}
	index := 0

	if s.index != -1 {
		index = (s.index + 1) % len(s.pidle)
	}

	s.index = index
	return s.pidle[index]
}

func (s *Sched) change(p puintptr, status pstatus) {

	if s == nil {
		return
	}

	switch status {
	case _Pidle:
		if !s.isContains(p) {
			s.pidle = append(s.pidle, p)
			s.npidle++
		}
	case _Pruning:
		for i, v := range s.pidle {
			if v == p {
				s.pidle = append(s.pidle[:i], s.pidle[i+1:]...)
				s.npidle--
				break
			}
		}
	case _Psyscall:
	case _Pgcstop:
	}
}

func (s *Sched) isContains(p puintptr) bool {
	for _, v := range s.pidle {
		if v == p {
			return true
		}
	}
	return false
}

func (s *Sched) SetStrategy(strategy string) {
	s.strategy = strategy
}

func G_output() chan []byte {
	return sched.sm.OutputChan()
}
