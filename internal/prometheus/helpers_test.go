package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
)

func init() {
	logger.Init("error")
	cfg = Config{EnableSnapshotMetrics: false}
}

func testCollector() *Collector {
	return NewCollector()
}

func drainMetrics(ch chan prometheus.Metric) []prometheus.Metric {
	var metrics []prometheus.Metric
	for {
		select {
		case m := <-ch:
			metrics = append(metrics, m)
		default:
			return metrics
		}
	}
}

func getMetricValue(m prometheus.Metric) float64 {
	var d dto.Metric
	_ = m.Write(&d)
	return d.GetGauge().GetValue()
}

func getMetricLabels(m prometheus.Metric) map[string]string {
	var d dto.Metric
	_ = m.Write(&d)
	labels := make(map[string]string)
	for _, lp := range d.GetLabel() {
		labels[lp.GetName()] = lp.GetValue()
	}
	return labels
}
