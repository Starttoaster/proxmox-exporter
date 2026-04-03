package prometheus

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	proxmox "github.com/starttoaster/go-proxmox"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
)

func (c *Collector) collectNodeUpMetric(ch chan<- prometheus.Metric, node proxmox.GetClusterResourcesData) {
	status := 0.0
	if strings.EqualFold(node.Status, "online") {
		status = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.nodeUp, prometheus.GaugeValue, status, node.Node)
}

// collectNodeSpecificMetrics fetches per-node data that isn't available in cluster resources
func (c *Collector) collectNodeSpecificMetrics(ch chan<- prometheus.Metric, nodeName string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer logger.Logger.Debug("finished requests for node data", "node", nodeName)

	disks, err := wrappedProxmox.GetNodeDisksList(nodeName)
	if err != nil {
		logger.Logger.Error("failed making request to get node disks", "node", nodeName, "error", err.Error())
	} else {
		c.collectDiskMetrics(ch, nodeName, disks)
	}

	certs, err := wrappedProxmox.GetNodeCertificatesInfo(nodeName)
	if err != nil {
		logger.Logger.Error("failed making request to get node certificates", "node", nodeName, "error", err.Error())
	} else {
		c.collectCertificateMetrics(ch, nodeName, certs)
	}

	nodeStatus, err := wrappedProxmox.GetNodeStatus(nodeName)
	if err != nil {
		logger.Logger.Error("failed making request to get node status", "node", nodeName, "error", err.Error())
	} else {
		ch <- prometheus.MustNewConstMetric(c.nodeVersion, prometheus.GaugeValue, float64(1), nodeName, nodeStatus.Data.PveVersion)
	}
}

func (c *Collector) collectDiskMetrics(ch chan<- prometheus.Metric, nodeName string, disks *proxmox.GetNodeDisksListResponse) {
	for _, disk := range disks.Data {
		status := 0.0
		if strings.EqualFold(disk.Health, "PASSED") || strings.EqualFold(disk.Health, "OK") {
			status = 1.0
		}
		if strings.EqualFold(disk.Health, "UNKNOWN") {
			status = -1.0
		}
		ch <- prometheus.MustNewConstMetric(c.diskSmartHealth, prometheus.GaugeValue, status, nodeName, disk.DevPath)
	}
}

func (c *Collector) collectCertificateMetrics(ch chan<- prometheus.Metric, nodeName string, certs *proxmox.GetNodeCertificatesInfoResponse) {
	for _, cert := range certs.Data {
		expDays := daysUntilUnixTime(cert.NotAfter)
		ch <- prometheus.MustNewConstMetric(c.daysUntilCertExpiry, prometheus.GaugeValue, float64(expDays), nodeName, cert.Subject)
	}
}

func (c *Collector) collectStorageMetrics(ch chan<- prometheus.Metric, storageResources []proxmox.GetClusterResourcesData) {
	for _, storage := range storageResources {
		storageName := ""
		if storage.Storage != nil {
			storageName = *storage.Storage
		}
		pluginType := ""
		if storage.PluginType != nil {
			pluginType = *storage.PluginType
		}
		shared := "false"
		if storage.Shared != nil && *storage.Shared == 1 {
			shared = "true"
		}
		var total, used float64
		if storage.MaxDisk != nil {
			total = float64(*storage.MaxDisk)
		}
		if storage.Disk != nil {
			used = float64(*storage.Disk)
		}

		ch <- prometheus.MustNewConstMetric(c.storageTotal, prometheus.GaugeValue, total, storage.Node, storageName, pluginType, shared)
		ch <- prometheus.MustNewConstMetric(c.storageUsed, prometheus.GaugeValue, used, storage.Node, storageName, pluginType, shared)
	}
}
