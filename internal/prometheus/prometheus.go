package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Collector contains all prometheus metric Descs
type Collector struct {
}

// NewCollector constructor function for Collector
func NewCollector() *Collector {
	return &Collector{}
}

// Describe contains all the prometheus descriptors for this metric collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {

}

// Collect instructs the prometheus client how to collect the metrics for each descriptor
func (c *Collector) Collect(ch chan<- prometheus.Metric) {

}
