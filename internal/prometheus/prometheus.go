package prometheus

import (
	"github.com/luthermonson/go-proxmox"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
)

// Collector contains all prometheus metric Descs
type Collector struct {
	up *prometheus.Desc
}

// NewCollector constructor function for Collector
func NewCollector() *Collector {
	return &Collector{
		up: prometheus.NewDesc(fqAddPrefix("up"),
			"Shows whether nodes and vms in a proxmox cluster are up. (0=down,1=up)",
			[]string{"type", "name"},
			nil,
		),
	}
}

// Describe contains all the prometheus descriptors for this metric collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
}

// Collect instructs the prometheus client how to collect the metrics for each descriptor
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Retrieve node statuses for the cluster
	nodeStatuses, err := wrappedProxmox.Nodes()
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	// Retrieve node info for each node from statuses
	var nodes []*proxmox.Node
	for _, nodeStatus := range nodeStatuses {
		// Add node up metric
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, float64(nodeStatus.Online), "node", nodeStatus.Name)

		// Get node from node status's name
		node, err := wrappedProxmox.Node(nodeStatus.Name)
		if err != nil {
			logger.Logger.Error(err.Error())
			return
		}
		nodes = append(nodes, node)
	}

	for _, node := range nodes {
		// Get VMs for node
		vms, err := wrappedProxmox.VirtualMachinesOnNode(node)
		if err != nil {
			logger.Logger.Error(err.Error())
			return
		}

		for _, vm := range vms {
			// Add vm up metric
			var vmUp float64 = 0.0
			if vm.IsRunning() {
				vmUp = 1.0
			}
			ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, vmUp, "qemu", vm.Name)
		}
	}
}
