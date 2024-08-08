package prometheus

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	proxmox "github.com/starttoaster/go-proxmox"
)

// collectLxcMetricsResponse is a struct wrapper for all LXC metrics that need to be passed back for control flow,
// usually for node-level or cluster-level metrics
type collectLxcMetricsResponse struct {
	cpusAllocated int
	memAllocated  int
}

// collectLxcMetrics adds metrics to the registry that are per-LXC and returns LXC aggregate data for higher level metrics
func (c *Collector) collectLxcMetrics(ch chan<- prometheus.Metric, node proxmox.GetNodesData, lxcs *proxmox.GetNodeLxcResponse) *collectLxcMetricsResponse {
	var res collectLxcMetricsResponse
	for _, lxc := range lxcs.Data {
		// Add lxc up metric
		status := 0.0
		if strings.EqualFold(lxc.Status, "running") {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, status, node.Node, lxc.Type, lxc.Name, string(lxc.VMID))

		// Add to LXC aggregate metrics
		res.cpusAllocated += lxc.CPUs
		res.memAllocated += lxc.MaxMem
	}
	return &res
}
