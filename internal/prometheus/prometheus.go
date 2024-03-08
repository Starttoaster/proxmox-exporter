package prometheus

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
	"github.com/starttoaster/proxmox-exporter/pkg/proxmox"
)

// Collector contains all prometheus metric Descs
type Collector struct {
	// Statuses
	nodeUp  *prometheus.Desc
	guestUp *prometheus.Desc

	// CPU
	clusterCPUsTotal *prometheus.Desc
	clusterCPUsAlloc *prometheus.Desc
	nodeCPUsTotal    *prometheus.Desc
	nodeCPUsAlloc    *prometheus.Desc
}

// NewCollector constructor function for Collector
func NewCollector() *Collector {
	return &Collector{
		nodeUp: prometheus.NewDesc(fqAddPrefix("node_up"),
			"Shows whether host nodes in a proxmox cluster are up. (0=down,1=up)",
			[]string{"type", "name"},
			nil,
		),
		guestUp: prometheus.NewDesc(fqAddPrefix("guest_up"),
			"Shows whether VMs and LXCs in a proxmox cluster are up. (0=down,1=up)",
			[]string{"type", "name", "vmid", "host"},
			nil,
		),

		clusterCPUsTotal: prometheus.NewDesc(fqAddPrefix("cluster_cpus_total"),
			"Total number of vCPU (cores/threads) for a cluster.",
			nil,
			nil,
		),
		clusterCPUsAlloc: prometheus.NewDesc(fqAddPrefix("cluster_cpus_allocated"),
			"Total number of vCPU (cores/threads) allocated to guests for a cluster.",
			nil,
			nil,
		),
		nodeCPUsTotal: prometheus.NewDesc(fqAddPrefix("node_cpus_total"),
			"Total number of vCPU (cores/threads) for a node.",
			[]string{"name"},
			nil,
		),
		nodeCPUsAlloc: prometheus.NewDesc(fqAddPrefix("node_cpus_allocated"),
			"Total number of vCPU (cores/threads) allocated to guests for a node.",
			[]string{"name"},
			nil,
		),
	}
}

// Describe contains all the prometheus descriptors for this metric collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.nodeUp
	ch <- c.guestUp
	ch <- c.clusterCPUsTotal
	ch <- c.clusterCPUsAlloc
	ch <- c.nodeCPUsTotal
	ch <- c.nodeCPUsAlloc
}

// Collect instructs the prometheus client how to collect the metrics for each descriptor
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Retrieve node statuses for the cluster
	nodes, err := wrappedProxmox.GetNodes()
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	// Cluster level metric variables (added to in each iteration of the loop below)
	clusterCPUs := 0
	clusterCPUsAlloc := 0

	// Retrieve node info for each node from statuses
	for _, node := range nodes.Data {
		// Get info for specific node
		nodeInfo, err := wrappedProxmox.GetNode(node.Node)
		if err != nil {
			logger.Logger.Error(err.Error())
			return
		}

		// Get VM metrics on this node
		vms, err := wrappedProxmox.GetNodeQemu(node.Node)
		if err != nil {
			logger.Logger.Error(err.Error())
			return
		}
		vmMetrics := c.collectVirtualMachineMetrics(ch, node, vms)

		// Get lxc data on this node
		lxcs, err := wrappedProxmox.GetNodeLxc(node.Node)
		if err != nil {
			logger.Logger.Error(err.Error())
			return
		}
		lxcMetrics := c.collectLxcMetrics(ch, node, lxcs)

		// Collect metrics for this node
		c.collectNodeUpMetric(ch, node)
		ch <- prometheus.MustNewConstMetric(c.nodeCPUsTotal, prometheus.GaugeValue, float64(nodeInfo.Data.CPUInfo.Cpus), node.Node)
		ch <- prometheus.MustNewConstMetric(c.nodeCPUsAlloc, prometheus.GaugeValue, float64(vmMetrics.cpusAllocated+lxcMetrics.cpusAllocated), node.Node)

		// Iterate on cluster metrics
		clusterCPUs += nodeInfo.Data.CPUInfo.Cpus
		clusterCPUsAlloc += vmMetrics.cpusAllocated + lxcMetrics.cpusAllocated
	}

	// Collect cluster metrics
	ch <- prometheus.MustNewConstMetric(c.clusterCPUsTotal, prometheus.GaugeValue, float64(clusterCPUs))
	ch <- prometheus.MustNewConstMetric(c.clusterCPUsAlloc, prometheus.GaugeValue, float64(clusterCPUsAlloc))
}

func (c *Collector) collectNodeUpMetric(ch chan<- prometheus.Metric, node proxmox.GetNodesData) {
	status := 0.0
	if strings.EqualFold(node.Status, "online") {
		status = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.nodeUp, prometheus.GaugeValue, status, node.Type, node.Node)
}

type collectVirtualMachineMetricsResponse struct {
	cpusAllocated int
}

func (c *Collector) collectVirtualMachineMetrics(ch chan<- prometheus.Metric, node proxmox.GetNodesData, vms *proxmox.GetNodeQemuResponse) collectVirtualMachineMetricsResponse {
	var res collectVirtualMachineMetricsResponse
	for _, vm := range vms.Data {
		// Add vm up metric
		status := 0.0
		if strings.EqualFold(vm.Status, "running") {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, status, "qemu", vm.Name, strconv.Itoa(vm.VMID), node.Node)

		// Add to CPU allocated to VMs on this node metric
		res.cpusAllocated += vm.Cpus
	}
	return res
}

type collectLxcMetricsResponse struct {
	cpusAllocated int
}

func (c *Collector) collectLxcMetrics(ch chan<- prometheus.Metric, node proxmox.GetNodesData, lxcs *proxmox.GetNodeLxcResponse) collectLxcMetricsResponse {
	var res collectLxcMetricsResponse
	for _, lxc := range lxcs.Data {
		// Add lxc up metric
		status := 0.0
		if strings.EqualFold(lxc.Status, "running") {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, status, lxc.Type, lxc.Name, lxc.VMID, node.Node)

		// Add to CPU allocated to LXCs on this node metric
		res.cpusAllocated += lxc.Cpus
	}
	return res
}
