package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Collector contains all prometheus metric Descs
type Collector struct {
	nodeUp *prometheus.Desc
}

// NewCollector constructor function for Collector
func NewCollector() *Collector {
	return &Collector{
		nodeUp: prometheus.NewDesc(fqAddPrefix("node_up"),
			"Shows whether nodes in a proxmox cluster are up.",
			[]string{}, nil,
		),
	}
}

// Describe contains all the prometheus descriptors for this metric collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.nodeUp
}

// Collect instructs the prometheus client how to collect the metrics for each descriptor
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.nodeUp, prometheus.GaugeValue, 1.0)
}
