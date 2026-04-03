package prometheus

import (
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	proxmox "github.com/starttoaster/go-proxmox"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
)

// collectLxcMetrics processes lxc entries from cluster resources
func (c *Collector) collectLxcMetrics(ch chan<- prometheus.Metric, lxcs []proxmox.GetClusterResourcesData) *collectGuestMetricsResponse {
	res := &collectGuestMetricsResponse{
		cpusPerNode: make(map[string]int),
		memPerNode:  make(map[string]int),
	}
	for _, lxc := range lxcs {
		name := ""
		if lxc.Name != nil {
			name = *lxc.Name
		}
		var vmid proxmox.IntOrString
		if lxc.VMID != nil {
			vmid = *lxc.VMID
		}
		tags := ""
		if lxc.Tags != nil {
			tags = *lxc.Tags
		}

		status := 0.0
		if strings.EqualFold(lxc.Status, "running") {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, status, lxc.Node, "lxc", name, string(vmid), tags)

		if lxc.MaxCPU != nil {
			res.cpusPerNode[lxc.Node] += *lxc.MaxCPU
		}
		if lxc.MaxMem != nil {
			res.memPerNode[lxc.Node] += *lxc.MaxMem
		}

		if cfg.EnableSnapshotMetrics {
			c.collectLxcSnapshotMetrics(ch, lxc.Node, name, vmid, tags)
		}
	}
	return res
}

func (c *Collector) collectLxcSnapshotMetrics(ch chan<- prometheus.Metric, nodeName, name string, vmid proxmox.IntOrString, tags string) {
	vmID, err := strconv.Atoi(string(vmid))
	if err != nil {
		logger.Logger.Error("failed converting LXC ID for lxc snapshots", "node", nodeName, "vm_id", vmid, "error", err.Error())
		return
	}

	snapshots, err := wrappedProxmox.GetLxcSnapshots(nodeName, vmID)
	if err != nil {
		logger.Logger.Error("failed making request to get lxc snapshots", "node", nodeName, "vm_id", vmid, "error", err.Error())
		return
	}

	snapshotCount := len(snapshots.Data) - 1 // subtract 1 because proxmox always has 1 "snapshot" that isn't rollback-able
	if snapshotCount < 0 {
		snapshotCount = 0
	}
	ch <- prometheus.MustNewConstMetric(c.guestSnapshotsCount, prometheus.GaugeValue, float64(snapshotCount), nodeName, "lxc", name, string(vmid), tags)

	for _, snapshot := range snapshots.Data {
		if snapshot.SnapTime != nil {
			snapUnixTime := int64(*snapshot.SnapTime)
			snapTime := time.Unix(snapUnixTime, 0)
			secondsAgo := time.Since(snapTime).Seconds()
			ch <- prometheus.MustNewConstMetric(c.guestSnapshotAgeSeconds, prometheus.GaugeValue, float64(secondsAgo), nodeName, "lxc", name, string(vmid), tags, snapshot.Name)
		}
	}
}
