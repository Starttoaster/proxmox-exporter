package prometheus

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	proxmox "github.com/starttoaster/go-proxmox"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
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
		ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, status, node.Node, lxc.Type, lxc.Name, string(lxc.VMID), lxc.Tags)

		// Add to LXC aggregate metrics
		res.cpusAllocated += lxc.CPUs
		res.memAllocated += lxc.MaxMem

		// Snapshot metrics (if enabled)
		if cfg.EnableSnapshotMetrics {
			c.collectLxcSnapshotMetrics(ch, node, lxc)
		}
	}
	return &res
}

func (c *Collector) collectLxcSnapshotMetrics(ch chan<- prometheus.Metric, node proxmox.GetNodesData, lxc proxmox.GetNodeLxcData) {
	// Some versions of proxmox's API returned the VM ID as a string, so we have to convert it to integer here
	vmID, err := strconv.Atoi(string(lxc.VMID))
	if err != nil {
		logger.Logger.Error("failed making request to get lxc snapshots", "node", node.Node, "vm_id", lxc.VMID, "error", err.Error())
		return
	}

	snapshots, err := wrappedProxmox.GetLxcSnapshots(node.Node, vmID)
	if err != nil {
		logger.Logger.Error("failed making request to get lxc snapshots", "node", node.Node, "vm_id", lxc.VMID, "error", err.Error())
		return
	}

	// Get snap count metric
	snapshotCount := len(snapshots.Data) - 1 // subtract 1 because proxmox always has 1 "snapshot" that isn't rollback-able
	if snapshotCount < 0 {
		snapshotCount = 0 // This should never be the case that the snapshot count is a negative number, but just in case
	}
	ch <- prometheus.MustNewConstMetric(c.guestSnapshotsCount, prometheus.GaugeValue, float64(snapshotCount), node.Node, "lxc", lxc.Name, string(lxc.VMID), lxc.Tags)
}
