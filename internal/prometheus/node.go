package prometheus

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
	"github.com/starttoaster/proxmox-exporter/pkg/proxmox"
)

// collectNodeResponse is a struct wrapper for all metrics that need to be passed back for control flow,
// usually for cluster-level metrics
type collectNodeResponse struct {
	clusterCPUs      int
	clusterCPUsAlloc int
	clusterMem       int
	clusterMemAlloc  int
}

func (c *Collector) collectNode(ch chan<- prometheus.Metric, node proxmox.GetNodesData, resultChan chan<- collectNodeResponse, wg *sync.WaitGroup) {
	defer wg.Done()

	// Collect metrics that just need node data
	c.collectNodeUpMetric(ch, node)

	// Get VM metrics on this node
	var vmMetrics *collectVirtualMachineMetricsResponse
	vms, err := wrappedProxmox.GetNodeQemu(node.Node)
	if err != nil {
		logger.Logger.Error(err.Error())
	} else {
		vmMetrics = c.collectVirtualMachineMetrics(ch, node, vms)
	}

	// Get lxc data on this node
	var lxcMetrics *collectLxcMetricsResponse
	lxcs, err := wrappedProxmox.GetNodeLxc(node.Node)
	if err != nil {
		logger.Logger.Error(err.Error())
	} else {
		lxcMetrics = c.collectLxcMetrics(ch, node, lxcs)
	}

	// Collect VM + LXC aggregate metrics
	if vmMetrics != nil && lxcMetrics != nil {
		ch <- prometheus.MustNewConstMetric(c.nodeCPUsAlloc, prometheus.GaugeValue, float64(vmMetrics.cpusAllocated+lxcMetrics.cpusAllocated), node.Node)
		ch <- prometheus.MustNewConstMetric(c.nodeMemAlloc, prometheus.GaugeValue, float64(vmMetrics.memAllocated+lxcMetrics.memAllocated), node.Node)
	}

	// Get storage data on this node
	stores, err := wrappedProxmox.GetNodeStorage(node.Node)
	if err != nil {
		logger.Logger.Error(err.Error())
	} else {
		c.collectStorageMetrics(ch, node, stores)
	}

	// Get disk data on this node
	disks, err := wrappedProxmox.GetNodeDisksList(node.Node)
	if err != nil {
		logger.Logger.Error(err.Error())
	} else {
		c.collectDiskMetrics(ch, node, disks)
	}

	// Get certificate data on this node
	certs, err := wrappedProxmox.GetNodeCertificatesInfo(node.Node)
	if err != nil {
		logger.Logger.Error(err.Error())
	} else {
		c.collectCertificateMetrics(ch, node, certs)
	}

	// Get status on this node
	nodeStatus, err := wrappedProxmox.GetNodeStatus(node.Node)
	if err != nil {
		logger.Logger.Error(err.Error())
	} else {
		c.collectNodeVersionMetric(ch, node, nodeStatus.Data)
		ch <- prometheus.MustNewConstMetric(c.nodeCPUsTotal, prometheus.GaugeValue, float64(nodeStatus.Data.CPUInfo.Cpus), node.Node)
		ch <- prometheus.MustNewConstMetric(c.nodeMemTotal, prometheus.GaugeValue, float64(nodeStatus.Data.Memory.Total), node.Node)
	}

	// Send the result back to the main function through the channel
	resultChan <- collectNodeResponse{
		clusterCPUs:      nodeStatus.Data.CPUInfo.Cpus,
		clusterCPUsAlloc: vmMetrics.cpusAllocated + lxcMetrics.cpusAllocated,
		clusterMem:       nodeStatus.Data.Memory.Total,
		clusterMemAlloc:  vmMetrics.memAllocated + lxcMetrics.memAllocated,
	}
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
		expDays, err := daysUntilUnixTime(cert.NotAfter)
		if err != nil {
			// Log error and give 0 days until expiry on metric to report a potential issue
			logger.Logger.Warn(err.Error(), "notafter", cert.NotAfter, "subject", cert.Subject)
			ch <- prometheus.MustNewConstMetric(c.daysUntilCertExpiry, prometheus.GaugeValue, 0.0, node.Node, cert.Subject)
		} else {
			ch <- prometheus.MustNewConstMetric(c.daysUntilCertExpiry, prometheus.GaugeValue, float64(expDays), node.Node, cert.Subject)
		}
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
