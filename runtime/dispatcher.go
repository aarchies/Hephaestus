package runtime

import (
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/aarchies/hephaestus/system/procs"
)

const (
	dataExpiration  = 5 * time.Second // 指标数据有效期
	minParentWeight = 0.2             // 父进程最低权重
	emaAlpha        = 0.7             // 平滑系数
)

type (
	Metrics struct {
		PID         int
		CPUUsage    float64   // 最近1秒平均CPU使用率[0-100]
		MemUsage    uint64    // RSS内存占用(bytes)
		MemPercent  float64   // 内存使用百分比[0-100]
		Pending     int       // 待处理任务数
		LastUpdated time.Time // 最后更新时间
	}

	Dispatcher struct {
		workers     atomic.Value // []*Metrics
		weightTable atomic.Value // map[int]float64
		rrCounter   uint64
	}
)

func (m *Metrics) Reset() {
	m.Update()
}

// 更新指标数据
func (m *Metrics) Update() {
	if cpu, err := procs.CPUPercent(int32(m.PID)); err == nil {
		m.CPUUsage = math.Min(math.Max(cpu, 0), 100) // 限制在0-100范围
	}

	if mem, percent, err := procs.MemoryInfo(int32(m.PID)); err == nil {
		m.MemUsage = mem
		m.MemPercent = math.Min(math.Max(percent, 0), 100)
	}

	m.LastUpdated = time.Now()
}

// 智能权重计算
func calculateWeights(metrics []*Metrics) map[int]float64 {
	now := time.Now()
	validMetrics := make([]*Metrics, 0, len(metrics))

	// 1. 数据有效性过滤
	for _, m := range metrics {
		if now.Sub(m.LastUpdated) < dataExpiration {
			validMetrics = append(validMetrics, m)
		}
	}
	if len(validMetrics) == 0 {
		return nil
	}

	// 2. 计算动态参数
	maxPending := 1                      // 防止除零
	for _, m := range validMetrics[1:] { // 排除父进程
		if m.Pending > maxPending {
			maxPending = m.Pending
		}
	}
	if maxPending < 100 { // 设置最小基准值
		maxPending = 100
	}

	weights := make(map[int]float64)
	totalChildWeight := 0.0

	// 3. 父进程权重计算（PID=0）
	parent := validMetrics[0]
	parentWeight := 0.4*(1-parent.CPUUsage/100) +
		0.3*(1-float64(parent.Pending)/1000) +
		0.3*(1-parent.MemPercent/100)
	parentWeight = math.Max(minParentWeight, parentWeight)

	// 4. 子进程权重计算
	for _, m := range validMetrics[1:] {
		// 负载因子计算
		cpuFactor := 1 - m.CPUUsage/100
		pendingFactor := 1 - float64(m.Pending)/float64(maxPending)
		memFactor := 1 - m.MemPercent/100

		// 综合加权（可配置权重系数）
		load := 0.5*cpuFactor + 0.2*pendingFactor + 0.3*memFactor
		load = math.Max(load, 0) // 确保非负

		weights[m.PID] = load
		totalChildWeight += load
	}

	// 5. 权重归一化
	remaining := 1.0 - parentWeight
	if totalChildWeight > 0 {
		scale := remaining / totalChildWeight
		for pid := range weights {
			weights[pid] *= scale
		}
	} else if len(validMetrics) > 1 { // 存在子进程但总权重为0
		avg := remaining / float64(len(validMetrics)-1)
		for _, m := range validMetrics[1:] {
			weights[m.PID] = avg
		}
	}
	weights[parent.PID] = parentWeight

	return weights
}

// 权重平滑迁移
func (d *Dispatcher) UpdateWeights(newWeights map[int]float64) {
	oldWeights, _ := d.weightTable.Load().(map[int]float64)
	blended := make(map[int]float64)

	for pid, newW := range newWeights {
		oldW := 0.0
		if oldWeights != nil {
			oldW = oldWeights[pid] // 不存在时返回0
		}
		blended[pid] = (1-emaAlpha)*oldW + emaAlpha*newW
	}
	d.weightTable.Store(blended)
}

// 增强型选择算法
func (d *Dispatcher) SelectWorker() *Metrics {
	workers, _ := d.workers.Load().([]*Metrics)
	if len(workers) == 0 {
		return nil
	}

	weights, _ := d.weightTable.Load().(map[int]float64)
	if weights == nil {
		// 无权重时降级为轮询
		atomic.AddUint64(&d.rrCounter, 1)
		return workers[int(d.rrCounter)%len(workers)]
	}

	// 构建有效权重列表
	var (
		total      float64
		candidates []*Metrics
		cumWeights []float64
	)
	for _, w := range workers {
		if weight, ok := weights[w.PID]; ok {
			total += weight
			cumWeights = append(cumWeights, total)
			candidates = append(candidates, w)
		}
	}

	if total <= 0 || len(candidates) == 0 {
		// 权重异常时降级为轮询
		atomic.AddUint64(&d.rrCounter, 1)
		return workers[int(d.rrCounter)%len(workers)]
	}

	// 带权重选择
	randVal := rand.Float64() * total
	for i, cum := range cumWeights {
		if randVal <= cum {
			return candidates[i]
		}
	}
	return candidates[len(candidates)-1] // 兜底返回最后一个
}
