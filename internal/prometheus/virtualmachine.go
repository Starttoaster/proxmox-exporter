package prometheus

import (
	"strings"

	"github.com/starttoaster/proxmox-exporter/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
	proxmox "github.com/starttoaster/go-proxmox"
)

// collectVirtualMachineMetricsResponse is a struct wrapper for all VM metrics that need to be passed back for control flow,
// usually for node-level or cluster-level metrics
type collectVirtualMachineMetricsResponse struct {
	cpusAllocated int
	memAllocated  int
}

// collectLxcMetrics adds metrics to the registry that are per-VM and returns VM aggregate data for higher level metrics
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
	}
	return &res
}
