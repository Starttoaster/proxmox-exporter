package prometheus

import (
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
	"github.com/starttoaster/proxmox-exporter/pkg/proxmox"
)

// Collector contains all prometheus metric Descs
type Collector struct {
	// Statuses
	nodeUp      *prometheus.Desc
	guestUp     *prometheus.Desc
	nodeVersion *prometheus.Desc

	// CPU
	clusterCPUsTotal *prometheus.Desc
	clusterCPUsAlloc *prometheus.Desc
	nodeCPUsTotal    *prometheus.Desc
	nodeCPUsAlloc    *prometheus.Desc

	// Mem
	clusterMemTotal *prometheus.Desc
	clusterMemAlloc *prometheus.Desc
	nodeMemTotal    *prometheus.Desc
	nodeMemAlloc    *prometheus.Desc

	// Storage
	storageTotal *prometheus.Desc
	storageUsed  *prometheus.Desc

	// Disk
	diskSmartHealth *prometheus.Desc

	// Certificates
	daysUntilCertExpiry *prometheus.Desc
}

// NewCollector constructor function for Collector
func NewCollector() *Collector {
	return &Collector{
		// Status metrics
		nodeUp: prometheus.NewDesc(fqAddPrefix("node_up"),
			"Shows whether host nodes in a proxmox cluster are up. (0=down,1=up)",
			[]string{"node"},
			nil,
		),
		guestUp: prometheus.NewDesc(fqAddPrefix("guest_up"),
			"Shows whether VMs and LXCs in a proxmox cluster are up. (0=down,1=up)",
			[]string{"node", "type", "name", "vmid"},
			nil,
		),
		nodeVersion: prometheus.NewDesc(fqAddPrefix("node_version"),
			"Shows PVE manager node version information",
			[]string{"node", "version"},
			nil,
		),

		// CPU metrics
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
			[]string{"node"},
			nil,
		),
		nodeCPUsAlloc: prometheus.NewDesc(fqAddPrefix("node_cpus_allocated"),
			"Total number of vCPU (cores/threads) allocated to guests for a node.",
			[]string{"node"},
			nil,
		),

		// Mem metrics
		clusterMemTotal: prometheus.NewDesc(fqAddPrefix("cluster_memory_total_bytes"),
			"Total amount of memory in bytes for a cluster.",
			nil,
			nil,
		),
		clusterMemAlloc: prometheus.NewDesc(fqAddPrefix("cluster_memory_allocated_bytes"),
			"Total amount of memory allocated in bytes to guests for a cluster.",
			nil,
			nil,
		),
		nodeMemTotal: prometheus.NewDesc(fqAddPrefix("node_memory_total_bytes"),
			"Total amount of memory in bytes for a node.",
			[]string{"node"},
			nil,
		),
		nodeMemAlloc: prometheus.NewDesc(fqAddPrefix("node_memory_allocated_bytes"),
			"Total amount of memory allocated in bytes to guests for a node.",
			[]string{"node"},
			nil,
		),

		// Disk metrics
		storageTotal: prometheus.NewDesc(fqAddPrefix("node_storage_total_bytes"),
			"Total amount of storage available in a volume on a node by storage type.",
			[]string{"node", "storage", "type", "shared"},
			nil,
		),
		storageUsed: prometheus.NewDesc(fqAddPrefix("node_storage_used_bytes"),
			"Total amount of storage used in a volume on a node by storage type.",
			[]string{"node", "storage", "type", "shared"},
			nil,
		),

		// Disk metrics
		diskSmartHealth: prometheus.NewDesc(fqAddPrefix("node_disk_smart_status"),
			"Disk SMART health status. (0=FAIL/Unknown,1=PASSED)",
			[]string{"node", "devpath"},
			nil,
		),

		// Cert metrics
		daysUntilCertExpiry: prometheus.NewDesc(fqAddPrefix("node_days_until_cert_expiration"),
			"Number of days until a certificate in PVE expires. Can report 0 days on metric collection errors, check exporter logs.",
			[]string{"node", "subject"},
			nil,
		),
	}
}

// Describe contains all the prometheus descriptors for this metric collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	// Status metrics
	ch <- c.nodeUp
	ch <- c.guestUp

	// CPU metrics
	ch <- c.clusterCPUsTotal
	ch <- c.clusterCPUsAlloc
	ch <- c.nodeCPUsTotal
	ch <- c.nodeCPUsAlloc

	// Mem metrics
	ch <- c.clusterMemTotal
	ch <- c.clusterMemAlloc
	ch <- c.nodeMemTotal
	ch <- c.nodeMemAlloc

	// Storage metrics
	ch <- c.storageTotal
	ch <- c.storageUsed

	// Disk metrics
	ch <- c.diskSmartHealth

	// Cert metrics
	ch <- c.daysUntilCertExpiry
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
	clusterMem := 0
	clusterMemAlloc := 0

	// Make waitgroup and results channel for cluster level metrics
	var wg sync.WaitGroup
	resultChan := make(chan collectNodeResponse, len(nodes.Data))

	// Collect node metrics from each of the nodes
	for _, node := range nodes.Data {
		wg.Add(1)
		go c.collectNode(ch, node, resultChan, &wg)
	}

	// Close the result channel after all goroutines finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results from the channel
	for result := range resultChan {
		clusterCPUs += result.clusterCPUs
		clusterCPUsAlloc += result.clusterCPUsAlloc
		clusterMem += result.clusterMem
		clusterMemAlloc += result.clusterMemAlloc
	}

	// Collect cluster metrics
	ch <- prometheus.MustNewConstMetric(c.clusterCPUsTotal, prometheus.GaugeValue, float64(clusterCPUs))
	ch <- prometheus.MustNewConstMetric(c.clusterCPUsAlloc, prometheus.GaugeValue, float64(clusterCPUsAlloc))
	ch <- prometheus.MustNewConstMetric(c.clusterMemTotal, prometheus.GaugeValue, float64(clusterMem))
	ch <- prometheus.MustNewConstMetric(c.clusterMemAlloc, prometheus.GaugeValue, float64(clusterMemAlloc))
}

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

	// Get storage data on this node
	stores, err := wrappedProxmox.GetNodeStorage(node.Node)
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}
	c.collectStorageMetrics(ch, node, stores)

	// Get disk data on this node
	disks, err := wrappedProxmox.GetNodeDisksList(node.Node)
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}
	c.collectDiskMetrics(ch, node, disks)

	// Get certificate data on this node
	certs, err := wrappedProxmox.GetNodeCertificatesInfo(node.Node)
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}
	c.collectCertificateMetrics(ch, node, certs)

	// Get status on this node
	nodeStatus, err := wrappedProxmox.GetNodeStatus(node.Node)
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	// Collect metrics for this node
	c.collectNodeUpMetric(ch, node)
	c.collectNodeVersionMetric(ch, node, nodeStatus.Data)
	ch <- prometheus.MustNewConstMetric(c.nodeCPUsTotal, prometheus.GaugeValue, float64(nodeStatus.Data.CPUInfo.Cpus), node.Node)
	ch <- prometheus.MustNewConstMetric(c.nodeCPUsAlloc, prometheus.GaugeValue, float64(vmMetrics.cpusAllocated+lxcMetrics.cpusAllocated), node.Node)
	ch <- prometheus.MustNewConstMetric(c.nodeMemTotal, prometheus.GaugeValue, float64(nodeStatus.Data.Memory.Total), node.Node)
	ch <- prometheus.MustNewConstMetric(c.nodeMemAlloc, prometheus.GaugeValue, float64(vmMetrics.memAllocated+lxcMetrics.memAllocated), node.Node)

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

// collectVirtualMachineMetricsResponse is a struct wrapper for all VM metrics that need to be passed back for control flow,
// usually for node-level or cluster-level metrics
type collectVirtualMachineMetricsResponse struct {
	cpusAllocated int
	memAllocated  int
}

// collectLxcMetrics adds metrics to the registry that are per-VM and returns VM aggregate data for higher level metrics
func (c *Collector) collectVirtualMachineMetrics(ch chan<- prometheus.Metric, node proxmox.GetNodesData, vms *proxmox.GetNodeQemuResponse) collectVirtualMachineMetricsResponse {
	var res collectVirtualMachineMetricsResponse
	for _, vm := range vms.Data {
		// Add vm up metric
		status := 0.0
		if strings.EqualFold(vm.Status, "running") {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, status, node.Node, "qemu", vm.Name, strconv.Itoa(vm.VMID))

		// Add to CPU allocated to VMs on this node metric
		res.cpusAllocated += vm.Cpus
		res.memAllocated += vm.MaxMem
	}
	return res
}

// collectLxcMetricsResponse is a struct wrapper for all LXC metrics that need to be passed back for control flow,
// usually for node-level or cluster-level metrics
type collectLxcMetricsResponse struct {
	cpusAllocated int
	memAllocated  int
}

// collectLxcMetrics adds metrics to the registry that are per-LXC and returns LXC aggregate data for higher level metrics
func (c *Collector) collectLxcMetrics(ch chan<- prometheus.Metric, node proxmox.GetNodesData, lxcs *proxmox.GetNodeLxcResponse) collectLxcMetricsResponse {
	var res collectLxcMetricsResponse
	for _, lxc := range lxcs.Data {
		// Add lxc up metric
		status := 0.0
		if strings.EqualFold(lxc.Status, "running") {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.guestUp, prometheus.GaugeValue, status, node.Node, lxc.Type, lxc.Name, lxc.VMID)

		// Add to CPU allocated to LXCs on this node metric
		res.cpusAllocated += lxc.Cpus
		res.memAllocated += lxc.MaxMem
	}
	return res
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
