package prometheus

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	proxmox "github.com/starttoaster/go-proxmox"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
)

// collectNodeResponse is a struct wrapper for all metrics that need to be passed back for control flow,
// usually for cluster-level metrics
type collectNodeResponse struct {
	clusterCPUs      int
	clusterCPUsAlloc int
	clusterMem       int
	clusterMemAlloc  int
}

func (c *Collector) collectNode(ch chan<- prometheus.Metric, clusterResources *proxmox.GetClusterResourcesResponse, node proxmox.GetNodesData, resultChan chan<- collectNodeResponse, wg *sync.WaitGroup) {
	defer wg.Done()
	defer logger.Logger.Debug("finished requests for node data", "node", node.Node)
	var resp collectNodeResponse

	// Collect metrics that just need node data
	c.collectNodeUpMetric(ch, node)

	// Exit here if node status is not "online", no more metrics to collect from this PVE host
	if !strings.EqualFold(node.Status, "online") {
		return
	}

	// Get VM metrics on this node
	var vmMetrics *collectVirtualMachineMetricsResponse
	vms, err := wrappedProxmox.GetNodeQemu(node.Node)
	if err != nil {
		logger.Logger.Error("failed making request to get node VMs", "node", node.Node, "error", err.Error())
	} else {
		vmMetrics = c.collectVirtualMachineMetrics(ch, clusterResources, node, vms)
	}

	// Get lxc data on this node
	var lxcMetrics *collectLxcMetricsResponse
	lxcs, err := wrappedProxmox.GetNodeLxc(node.Node)
	if err != nil {
		logger.Logger.Error("failed making request to get node LXCs", "node", node.Node, "error", err.Error())
	} else {
		lxcMetrics = c.collectLxcMetrics(ch, node, lxcs)
	}

	// Collect VM + LXC aggregate metrics
	if vmMetrics != nil && lxcMetrics != nil {
		resp.clusterCPUsAlloc = vmMetrics.cpusAllocated + lxcMetrics.cpusAllocated
		resp.clusterMemAlloc = vmMetrics.memAllocated + lxcMetrics.memAllocated
		ch <- prometheus.MustNewConstMetric(c.nodeCPUsAlloc, prometheus.GaugeValue, float64(vmMetrics.cpusAllocated+lxcMetrics.cpusAllocated), node.Node)
		ch <- prometheus.MustNewConstMetric(c.nodeMemAlloc, prometheus.GaugeValue, float64(vmMetrics.memAllocated+lxcMetrics.memAllocated), node.Node)
	}

	// Get storage data on this node
	stores, err := wrappedProxmox.GetNodeStorage(node.Node)
	if err != nil {
		logger.Logger.Error("failed making request to get node storage", "node", node.Node, "error", err.Error())
	} else {
		c.collectStorageMetrics(ch, node, stores)
	}

	// Get disk data on this node
	disks, err := wrappedProxmox.GetNodeDisksList(node.Node)
	if err != nil {
		logger.Logger.Error("failed making request to get node disks", "node", node.Node, "error", err.Error())
	} else {
		c.collectDiskMetrics(ch, node, disks)
	}

	// Get certificate data on this node
	certs, err := wrappedProxmox.GetNodeCertificatesInfo(node.Node)
	if err != nil {
		logger.Logger.Error("failed making request to get node certificates", "node", node.Node, "error", err.Error())
	} else {
		c.collectCertificateMetrics(ch, node, certs)
	}

	// Get status on this node
	nodeStatus, err := wrappedProxmox.GetNodeStatus(node.Node)
	if err != nil {
		logger.Logger.Error("failed making request to get node status", "node", node.Node, "error", err.Error())
	} else {
		resp.clusterCPUs = nodeStatus.Data.CPUInfo.CPUs
		resp.clusterMem = nodeStatus.Data.Memory.Total
		c.collectNodeVersionMetric(ch, node, nodeStatus.Data)
		ch <- prometheus.MustNewConstMetric(c.nodeCPUsTotal, prometheus.GaugeValue, float64(nodeStatus.Data.CPUInfo.CPUs), node.Node)
		ch <- prometheus.MustNewConstMetric(c.nodeMemTotal, prometheus.GaugeValue, float64(nodeStatus.Data.Memory.Total), node.Node)
	}

	// Send the result back to the main function through the channel
	resultChan <- resp
}

func (c *Collector) collectNodeVersionMetric(ch chan<- prometheus.Metric, node proxmox.GetNodesData, status proxmox.GetNodeStatusData) {
	ch <- prometheus.MustNewConstMetric(c.nodeVersion, prometheus.GaugeValue, float64(1), node.Node, status.PveVersion)
}

func (c *Collector) collectNodeUpMetric(ch chan<- prometheus.Metric, node proxmox.GetNodesData) {
	status := 0.0
	if strings.EqualFold(node.Status, "online") {
		status = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.nodeUp, prometheus.GaugeValue, status, node.Node)
}

func (c *Collector) collectDiskMetrics(ch chan<- prometheus.Metric, node proxmox.GetNodesData, disks *proxmox.GetNodeDisksListResponse) {
	for _, disk := range disks.Data {
		// Add disk health metric
		status := 0.0
		if strings.EqualFold(disk.Health, "PASSED") {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.diskSmartHealth, prometheus.GaugeValue, status, node.Node, disk.DevPath)
	}
}

func (c *Collector) collectCertificateMetrics(ch chan<- prometheus.Metric, node proxmox.GetNodesData, certs *proxmox.GetNodeCertificatesInfoResponse) {
	for _, cert := range certs.Data {
		// Add days until certificate expiration metric
		expDays := daysUntilUnixTime(cert.NotAfter)
		ch <- prometheus.MustNewConstMetric(c.daysUntilCertExpiry, prometheus.GaugeValue, float64(expDays), node.Node, cert.Subject)
	}
}

func (c *Collector) collectStorageMetrics(ch chan<- prometheus.Metric, node proxmox.GetNodesData, storages *proxmox.GetNodeStorageResponse) {
	for _, storage := range storages.Data {
		// Creates a boolean label string for the PVE storage volume that tells whether the volume is shared in a cluster
		shared := "false"
		if storage.Shared == 1 {
			shared = "true"
		}

		ch <- prometheus.MustNewConstMetric(c.storageTotal, prometheus.GaugeValue, float64(storage.Total), node.Node, storage.Storage, storage.Type, shared)
		ch <- prometheus.MustNewConstMetric(c.storageUsed, prometheus.GaugeValue, float64(storage.Used), node.Node, storage.Storage, storage.Type, shared)
	}
}
