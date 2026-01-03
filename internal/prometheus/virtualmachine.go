package prometheus

import (
	"strconv"
	"strings"

	"github.com/starttoaster/proxmox-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
	proxmox "github.com/starttoaster/go-proxmox"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
)

// collectVirtualMachineMetricsResponse is a struct wrapper for all VM metrics that need to be passed back for control flow,
// usually for node-level or cluster-level metrics
type collectVirtualMachineMetricsResponse struct {
	cpusAllocated int
	memAllocated  int
}

// collectVirtualMachineMetrics adds metrics to the registry that are per-VM and returns VM aggregate data for higher level metrics
func (c *Collector) collectVirtualMachineMetrics(ch chan<- prometheus.Metric, clusterResources *proxmox.GetClusterResourcesResponse, node proxmox.GetNodesData, vms *proxmox.GetNodeQemuResponse) *collectVirtualMachineMetricsResponse {
	var res collectVirtualMachineMetricsResponse
	for _, vm := range vms.Data {
		// Checks if cluster resources were provided. If they were, this will check if a VM is a template.
		var vmIsTemplate bool
		if clusterResources != nil {
			for _, res := range clusterResources.Data {
				var id proxmox.IntOrString
				if res.VMID != nil {
					id = *res.VMID
				}
				var template int
				if res.Template != nil {
					template = *res.Template
				}
				if vm.VMID == id && template == 1 {
					vmIsTemplate = true
				}
			}
		}

		// Don't collect VM metrics on templates
		if vmIsTemplate {
			logger.Logger.Debug("excluding VM from collecting metrics because it is a template.", "name", vm.Name, "ID", vm.VMID)
			continue
		}

		// Add vm up metric
		status := 0.0
		if strings.EqualFold(vm.Status, "running") {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, status, node.Node, "qemu", vm.Name, string(vm.VMID), vm.Tags)

		// Add to VM aggregate metrics
		res.cpusAllocated += vm.CPUs
		res.memAllocated += vm.MaxMem

		// Snapshot metrics (if enabled)
		if cfg.EnableSnapshotMetrics {
			c.collectQemuSnapshotMetrics(ch, node, vm)
		}
	}
	return &res
}

func (c *Collector) collectQemuSnapshotMetrics(ch chan<- prometheus.Metric, node proxmox.GetNodesData, vm proxmox.GetNodeQemuData) {
	// Some versions of proxmox's API returned the VM ID as a string, so we have to convert it to integer here
	vmID, err := strconv.Atoi(string(vm.VMID))
	if err != nil {
		logger.Logger.Error("failed making request to get qemu snapshots", "node", node.Node, "vm_id", vm.VMID, "error", err.Error())
		return
	}

	snapshots, err := wrappedProxmox.GetQemuSnapshots(node.Node, vmID)
	if err != nil {
		logger.Logger.Error("failed making request to get qemu snapshots", "node", node.Node, "vm_id", vm.VMID, "error", err.Error())
		return
	}

	// Get snap count metric
	snapshotCount := len(snapshots.Data) - 1 // subtract 1 because proxmox always has 1 "snapshot" that isn't rollback-able
	if snapshotCount < 0 {
		snapshotCount = 0 // This should never be the case that the snapshot count is a negative number, but just in case
	}
	ch <- prometheus.MustNewConstMetric(c.guestSnapshotsCount, prometheus.GaugeValue, float64(snapshotCount), node.Node, "qemu", vm.Name, string(vm.VMID), vm.Tags)
}
