package prometheus

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
)

// Collector contains all prometheus metric Descs
type Collector struct {
	hostUp  *prometheus.Desc
	guestUp *prometheus.Desc
}

// NewCollector constructor function for Collector
func NewCollector() *Collector {
	return &Collector{
		hostUp: prometheus.NewDesc(fqAddPrefix("host_up"),
			"Shows whether host nodes in a proxmox cluster are up. (0=down,1=up)",
			[]string{"type", "name"},
			nil,
		),
		guestUp: prometheus.NewDesc(fqAddPrefix("guest_up"),
			"Shows whether VMs and LXCs in a proxmox cluster are up. (0=down,1=up)",
			[]string{"type", "name", "host"},
			nil,
		),
	}
}

// Describe contains all the prometheus descriptors for this metric collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.hostUp
	ch <- c.guestUp
}

// Collect instructs the prometheus client how to collect the metrics for each descriptor
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Retrieve node statuses for the cluster
	nodes, err := wrappedProxmox.GetNodes()
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	// Retrieve node info for each node from statuses
	for _, node := range nodes.Data {
		// Get Node status online float metric from string
		var nodeStatusOnline float64
		if strings.EqualFold(node.Status, "online") {
			nodeStatusOnline = 1.0
		}

		// Add node up metric
		ch <- prometheus.MustNewConstMetric(c.hostUp, prometheus.GaugeValue, nodeStatusOnline, node.Type, node.Node)

		// Get VM statuses from each node
		vms, err := wrappedProxmox.GetNodeQemu(node.Node)
		if err != nil {
			logger.Logger.Error(err.Error())
			return
		}

		// Retrieve info for each VM
		for _, vm := range vms.Data {
			// Get Node status online float metric from string
			var vmStatusOnline float64
			if strings.EqualFold(vm.Status, "running") {
				vmStatusOnline = 1.0
			}

			// Add node up metric
			ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, vmStatusOnline, "qemu", vm.Name, node.Node)
		}

		// Get LXC statuses from each node
		lxcs, err := wrappedProxmox.GetNodeLxc(node.Node)
		if err != nil {
			logger.Logger.Error(err.Error())
			return
		}

		// Retrieve info for each LXC
		for _, lxc := range lxcs.Data {
			// Get LXC status online float metric from string
			var lxcStatusOnline float64
			if strings.EqualFold(lxc.Status, "running") {
				lxcStatusOnline = 1.0
			}

			// Add node up metric
			ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, lxcStatusOnline, lxc.Type, lxc.Name, node.Node)
		}
	}
}
