package pollers

import (
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/vleukhin/prom-light/internal/metrics"
)

type PsPoller struct {
	ticker *time.Ticker
}

func NewPsPoller(interval time.Duration) PsPoller {
	return PsPoller{
		time.NewTicker(interval),
	}
}

func (p PsPoller) Poll(ch chan<- metrics.Metrics) {
	var (
		err         error
		mtrcs       = make(metrics.Metrics, 0, 3)
		memory      *mem.VirtualMemoryStat
		utilization []float64
	)

	for {
		<-p.ticker.C

		memory, err = mem.VirtualMemory()
		if err != nil {
			log.Error().Err(err)
			continue
		}
		utilization, err = cpu.Percent(time.Second, true)
		if err != nil {
			log.Error().Err(err)
			continue
		}

		mtrcs = append(mtrcs, metrics.MakeGaugeMetric("TotalMemory", metrics.Gauge(memory.Total)))
		mtrcs = append(mtrcs, metrics.MakeGaugeMetric("FreeMemory", metrics.Gauge(memory.Free)))
		for cpuNum, percent := range utilization {
			mtrcs = append(mtrcs, metrics.MakeGaugeMetric("CPUutilization"+strconv.Itoa(cpuNum+1), metrics.Gauge(percent)))
		}

		ch <- mtrcs
	}
}
