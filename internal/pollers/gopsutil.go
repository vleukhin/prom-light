package pollers

import (
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type PsPoller struct {
}

func (p PsPoller) Poll() (metrics.Metrics, error) {
	var (
		err         error
		mtrcs       = make(metrics.Metrics, 0, 3)
		memory      *mem.VirtualMemoryStat
		utilization []float64
	)

	memory, err = mem.VirtualMemory()
	if err != nil {
		return mtrcs, err
	}
	utilization, err = cpu.Percent(time.Second, true)
	if err != nil {
		return mtrcs, err
	}

	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("TotalMemory", metrics.Gauge(memory.Total)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("FreeMemory", metrics.Gauge(memory.Free)))
	for cpuNum, percent := range utilization {
		mtrcs = append(mtrcs, metrics.MakeGaugeMetric("CPUutilization"+strconv.Itoa(cpuNum+1), metrics.Gauge(percent)))
	}

	return mtrcs, nil
}
