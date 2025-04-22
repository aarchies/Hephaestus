//go:build linux
// +build linux

package runtime

import (
	"context"

	"runtime"

	"github.com/aarchies/hephaestus/runtime/actor"
	"github.com/aarchies/hephaestus/runtime/sequence"
	cpu_test "github.com/aarchies/hephaestus/system/cpu"
	"github.com/aarchies/hephaestus/system/sem"

	"time"
	"unsafe"
)

type muintptr *M
type M struct {
	id       int32                     // ID
	actor    *actor.Actuators          // Actuator is a pointer to actuator.Actuator
	pc       int                       // Count is a counter
	p        puintptr                  // 执行 go 代码时持有的 p (如果没有执行则为nil)
	alllink  unsafe.Pointer            // pool
	spinning bool                      // m 当前没有运行 work 且正处于寻找 work 的活跃状态
	futex    *sem.Semaphore            // futex
	sm       *sequence.SequenceManager // seq manager
	metrics  *Metrics                  // current m metrics
	ctx      context.Context
	cancel   context.CancelFunc
}

func mcommoninit() muintptr {

	ctx, cancel := context.WithCancel(context.Background())

	base := &M{
		id:       0,
		spinning: true,
		p:        newp(0),
		actor:    nil,
		metrics:  &Metrics{PID: 0},
		sm:       sched.sm,
		ctx:      ctx,
		cancel:   cancel,
	}
	futex, err := sem.New(0)
	if err != nil {
		panic(err)
	}
	base.futex = futex
	base.p.m = base
	base.bindCPU()
	base.metrics.Pending = base.p.runq.len()
	base.metrics.Update()
	go base.metricsing(3)
	go base.healthCheck()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if base.spinning {
					runtime.Gosched()
					base.futex.Wait()
				}

				// 获取本地队列
				g := (*base.p).runq.next()
				if g != nil {
					println("master m is new task!", base.pc)
					base.pc++
					result := releaseG(g)

					result.isComplted = true
					result = nil
				} else {
					g := (*base.p).runq.fetallp() // 获取全局队列
					if g == nil {
						g = (*base.p).runq.plagiarize() // 剽窃
					}
					if g != nil {
						println("master m is new task!", base.pc)
						base.pc++
						result := releaseG(g)

						result.isComplted = true
						result = nil
					} else if !base.spinning {
						println("master m is spinning!", base.pc)
						base.spinning = true
						base.pc = 0
					}
				}
			}
		}
	}()

	return base
}

func newm(id int32) muintptr {

	ctx, cancel := context.WithCancel(context.Background())

	m := &M{
		id:       id,
		spinning: true,
		pc:       0,
		actor:    actor.NewWorker(id),
		p:        newp(id),
		sm:       sched.sm,
		ctx:      ctx,
		cancel:   cancel,
		metrics:  &Metrics{},
	}
	m.p.m = m
	futex, err := sem.New(uint(id))
	if err != nil {
		panic(err)
	}
	m.futex = futex

	if err := m.actor.Run(); err != nil {
		panic(err)
	}

	m.bindCPU()
	m.metrics.Pending = m.p.runq.len()
	m.metrics.Update()
	go m.metricsing(3)
	go m.reducep()
	go m.runnable()
	go m.healthCheck()

	return m
}

func (m *M) wakeup() {
	if m.spinning {
		m.spinning = false
		m.futex.Post()
	}
}

func (m *M) healthCheck() {
	t := time.NewTicker(time.Duration(mHealthCheckInterval) * time.Second)
	defer t.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-t.C:

			if time.Since(m.metrics.LastUpdated) > 2*time.Duration(mHealthCheckInterval)*time.Second {
				go m.metrics.Reset()
			}

			if m.actor != nil && m.actor.Status() == actor.Exited {
				println("M marked as faulty", m.id)
				go m.recover()
			}

			if m.metrics.CPUUsage > 95 || m.metrics.MemPercent > 95 {
				println("M overloading, evacuating tasks", m.id)
				m.evacuate()
			}
		}
	}
}

func (m *M) recover() {
	if m.id != 0 {
		actor := actor.NewWorker(m.id)
		if err := m.actor.Run(); err != nil {
			println(err)
		}
		m.bindCPU()
		m.actor = actor
	}
}

func (m *M) evacuate() {}

func (m *M) bindCPU() {
	if m.id != 0 {
		if m.actor.Pid() != 0 {
			m.metrics.PID = m.actor.Pid()
			cpu_test.BindCPU(m.actor.Pid(), IdleCpus[0])
			cpu_test.BindCPU(m.actor.Pid(), IdleCpus[1])
			IdleCpus = IdleCpus[2:]
		}
	} else {
		for _, v := range ParentCpus {
			cpu_test.BindCPU(0, v)
		}
	}
}

func (m *M) runnable() {

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			if m.spinning {
				runtime.Gosched()
				m.futex.Wait()
			}
			m.fetch()
		}
	}
}

func (m *M) reducep() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case r := <-m.actor.Output():
			m.sm.Submit(r.SeqId, r.Buffer)
		}
	}
}

func (m *M) invoke(g guintptr) {
	println("m is new task!", m.id)
	data := releaseG(g)
	m.pc++
	m.actor.Input(data.id, data.buffer)
	data.isComplted = true
}

func (m *M) fetch() {
	// 获取本地队列
	g := (*m.p).runq.next()
	if g != nil {
		println("m is new task!", m.pc, m.id)
		m.invoke(g)
	} else {
		g := (*m.p).runq.fetallp() // 获取全局队列
		if g == nil {
			g = (*m.p).runq.plagiarize() // 剽窃其他队列
		}

		if g != nil {
			println("m is new task!", m.pc, m.id)
			m.invoke(g)
		} else if !m.spinning {
			println("m is spinning!", m.pc, m.id)
			m.spinning = true
			m.pc = 0
		}
	}
}

func (m *M) Cease() {
	m.cancel()
	m.actor.Exit()
	(*m.p).cleanup()
}

func (m *M) metricsing(per int) {
	t := time.NewTicker(time.Second * time.Duration(per))
	defer t.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-t.C:
			m.metrics.Pending = m.p.runq.len()
			m.metrics.Update()
		}
	}
}
