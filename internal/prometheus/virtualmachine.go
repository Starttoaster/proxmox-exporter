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

// collectGuestMetricsResponse tracks per-node CPU/memory allocations for both VMs and LXCs
type collectGuestMetricsResponse struct {
	cpusPerNode map[string]int
	memPerNode  map[string]int
}

// collectVirtualMachineMetrics processes qemu entries from cluster resources
func (c *Collector) collectVirtualMachineMetrics(ch chan<- prometheus.Metric, vms []proxmox.GetClusterResourcesData) *collectGuestMetricsResponse {
	res := &collectGuestMetricsResponse{
		cpusPerNode: make(map[string]int),
		memPerNode:  make(map[string]int),
	}
	for _, vm := range vms {
		name := ""
		if vm.Name != nil {
			name = *vm.Name
		}
		var vmid proxmox.IntOrString
		if vm.VMID != nil {
			vmid = *vm.VMID
		}
		tags := ""
		if vm.Tags != nil {
			tags = *vm.Tags
		}

		if vm.Template != nil && *vm.Template == 1 {
			logger.Logger.Debug("excluding VM from collecting metrics because it is a template.", "name", name, "ID", vmid)
			continue
		}

		status := 0.0
		if strings.EqualFold(vm.Status, "running") {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, status, vm.Node, "qemu", name, string(vmid), tags)

		if vm.MaxCPU != nil {
			res.cpusPerNode[vm.Node] += *vm.MaxCPU
		}
		if vm.MaxMem != nil {
			res.memPerNode[vm.Node] += *vm.MaxMem
		}

		if cfg.EnableSnapshotMetrics {
			c.collectQemuSnapshotMetrics(ch, vm.Node, name, vmid, tags)
		}
	}
	return res
}

func (c *Collector) collectQemuSnapshotMetrics(ch chan<- prometheus.Metric, nodeName, name string, vmid proxmox.IntOrString, tags string) {
	vmID, err := strconv.Atoi(string(vmid))
	if err != nil {
		logger.Logger.Error("failed converting VM ID for qemu snapshots", "node", nodeName, "vm_id", vmid, "error", err.Error())
		return
	}

	snapshots, err := wrappedProxmox.GetQemuSnapshots(nodeName, vmID)
	if err != nil {
		logger.Logger.Error("failed making request to get qemu snapshots", "node", nodeName, "vm_id", vmid, "error", err.Error())
		return
	}

	snapshotCount := len(snapshots.Data) - 1 // subtract 1 because proxmox always has 1 "snapshot" that isn't rollback-able
	if snapshotCount < 0 {
		snapshotCount = 0
	}
	ch <- prometheus.MustNewConstMetric(c.guestSnapshotsCount, prometheus.GaugeValue, float64(snapshotCount), nodeName, "qemu", name, string(vmid), tags)

	for _, snapshot := range snapshots.Data {
		if snapshot.SnapTime != nil {
			snapUnixTime := int64(*snapshot.SnapTime)
			snapTime := time.Unix(snapUnixTime, 0)
			secondsAgo := time.Since(snapTime).Seconds()
			ch <- prometheus.MustNewConstMetric(c.guestSnapshotAgeSeconds, prometheus.GaugeValue, float64(secondsAgo), nodeName, "qemu", name, string(vmid), tags, snapshot.Name)
		}
	}
}
