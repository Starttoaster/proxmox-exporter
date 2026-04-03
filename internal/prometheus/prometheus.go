package prometheus

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/starttoaster/go-proxmox"
	"github.com/starttoaster/proxmox-exporter/internal/logger"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
)

var cfg Config

// Config is the configuration to pass to the init function
type Config struct {
	EnableSnapshotMetrics bool
}

// Init is a helper to configure the metrics collector
func Init(c Config) {
	cfg = c
}

// Collector contains all prometheus metric Descs
type Collector struct {
	// Exporter
	clientCount *prometheus.Desc

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

	// Snapshots
	guestSnapshotsCount     *prometheus.Desc
	guestSnapshotAgeSeconds *prometheus.Desc

	// Disk
	diskSmartHealth *prometheus.Desc

	// Certificates
	daysUntilCertExpiry *prometheus.Desc
}

// NewCollector constructor function for Collector
func NewCollector() *Collector {
	// Initialize constant labels for timeseries this exporter makes
	var constLabels = make(prometheus.Labels)

	// Add cluster label if the API package found a cluster name
	if wrappedProxmox.ClusterName != "" {
		constLabels["cluster"] = wrappedProxmox.ClusterName
	}

	collector := Collector{
		// Exporter metrics
		clientCount: prometheus.NewDesc(fqAddPrefix("exporter_client_count"),
			"Counts number of Proxmox clients (banned and unbanned)",
			[]string{"status"},
			constLabels,
		),

		// Status metrics
		nodeUp: prometheus.NewDesc(fqAddPrefix("node_up"),
			"Shows whether host nodes in a proxmox cluster are up. (0=down,1=up)",
			[]string{"node"},
			constLabels,
		),
		guestUp: prometheus.NewDesc(fqAddPrefix("guest_up"),
			"Shows whether VMs and LXCs in a proxmox cluster are up. (0=down,1=up)",
			[]string{"node", "type", "name", "vmid", "tags"},
			constLabels,
		),
		nodeVersion: prometheus.NewDesc(fqAddPrefix("node_version"),
			"Shows PVE manager node version information",
			[]string{"node", "version"},
			constLabels,
		),

		// CPU metrics
		clusterCPUsTotal: prometheus.NewDesc(fqAddPrefix("cluster_cpus_total"),
			"Total number of vCPU (cores/threads) for a cluster.",
			nil,
			constLabels,
		),
		clusterCPUsAlloc: prometheus.NewDesc(fqAddPrefix("cluster_cpus_allocated"),
			"Total number of vCPU (cores/threads) allocated to guests for a cluster.",
			nil,
			constLabels,
		),
		nodeCPUsTotal: prometheus.NewDesc(fqAddPrefix("node_cpus_total"),
			"Total number of vCPU (cores/threads) for a node.",
			[]string{"node"},
			constLabels,
		),
		nodeCPUsAlloc: prometheus.NewDesc(fqAddPrefix("node_cpus_allocated"),
			"Total number of vCPU (cores/threads) allocated to guests for a node.",
			[]string{"node"},
			constLabels,
		),

		// Mem metrics
		clusterMemTotal: prometheus.NewDesc(fqAddPrefix("cluster_memory_total_bytes"),
			"Total amount of memory in bytes for a cluster.",
			nil,
			constLabels,
		),
		clusterMemAlloc: prometheus.NewDesc(fqAddPrefix("cluster_memory_allocated_bytes"),
			"Total amount of memory allocated in bytes to guests for a cluster.",
			nil,
			constLabels,
		),
		nodeMemTotal: prometheus.NewDesc(fqAddPrefix("node_memory_total_bytes"),
			"Total amount of memory in bytes for a node.",
			[]string{"node"},
			constLabels,
		),
		nodeMemAlloc: prometheus.NewDesc(fqAddPrefix("node_memory_allocated_bytes"),
			"Total amount of memory allocated in bytes to guests for a node.",
			[]string{"node"},
			constLabels,
		),

		// Disk metrics
		storageTotal: prometheus.NewDesc(fqAddPrefix("node_storage_total_bytes"),
			"Total amount of storage available in a volume on a node by storage type.",
			[]string{"node", "storage", "type", "shared"},
			constLabels,
		),
		storageUsed: prometheus.NewDesc(fqAddPrefix("node_storage_used_bytes"),
			"Total amount of storage used in a volume on a node by storage type.",
			[]string{"node", "storage", "type", "shared"},
			constLabels,
		),

		// Disk metrics
		diskSmartHealth: prometheus.NewDesc(fqAddPrefix("node_disk_smart_status"),
			"Disk SMART health status. (0=FAIL/Unknown,1=PASSED/OK)",
			[]string{"node", "devpath"},
			constLabels,
		),

		// Cert metrics
		daysUntilCertExpiry: prometheus.NewDesc(fqAddPrefix("node_days_until_cert_expiration"),
			"Number of days until a certificate in PVE expires. Can report 0 days on metric collection errors, check exporter logs.",
			[]string{"node", "subject"},
			constLabels,
		),
	}

	// Enable snapshot metrics
	if cfg.EnableSnapshotMetrics {
		collector.guestSnapshotsCount = prometheus.NewDesc(fqAddPrefix("guest_snapshots"),
			"Count of snapshots taken for a given guest.",
			[]string{"node", "type", "name", "vmid", "tags"},
			constLabels,
		)
		collector.guestSnapshotAgeSeconds = prometheus.NewDesc(fqAddPrefix("guest_snapshot_age_seconds"),
			"Number of seconds since a snapshot was taken for a given guest.",
			[]string{"node", "type", "name", "vmid", "tags", "snapshot"},
			constLabels,
		)
	}

	return &collector
}

// Describe contains all the prometheus descriptors for this metric collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	// Exporter metrics
	ch <- c.clientCount

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

	// Snapshot metrics
	if cfg.EnableSnapshotMetrics {
		ch <- c.guestSnapshotsCount
		ch <- c.guestSnapshotAgeSeconds
	}

	// Disk metrics
	ch <- c.diskSmartHealth

	// Cert metrics
	ch <- c.daysUntilCertExpiry
}

// Collect instructs the prometheus client how to collect the metrics for each descriptor
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.clientCount, prometheus.GaugeValue, float64(wrappedProxmox.GetBannedClientCount()), "banned")
	ch <- prometheus.MustNewConstMetric(c.clientCount, prometheus.GaugeValue, float64(wrappedProxmox.GetUnbannedClientCount()), "unbanned")

	// Single API call replaces GetNodes + per-node GetNodeQemu/GetNodeLxc/GetNodeStorage
	clusterResources, err := wrappedProxmox.GetClusterResources()
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	// Categorize resources by type in a single pass
	var nodeResources, qemuResources, lxcResources, storageResources []proxmox.GetClusterResourcesData
	for _, r := range clusterResources.Data {
		switch r.Type {
		case "node":
			nodeResources = append(nodeResources, r)
		case "qemu":
			qemuResources = append(qemuResources, r)
		case "lxc":
			lxcResources = append(lxcResources, r)
		case "storage":
			storageResources = append(storageResources, r)
		}
	}

	// Process node metrics from cluster resources
	clusterCPUs := 0
	clusterMem := 0
	var onlineNodes []string
	for _, node := range nodeResources {
		c.collectNodeUpMetric(ch, node)
		if strings.EqualFold(node.Status, "online") {
			onlineNodes = append(onlineNodes, node.Node)
		}
		if node.MaxCPU != nil {
			ch <- prometheus.MustNewConstMetric(c.nodeCPUsTotal, prometheus.GaugeValue, float64(*node.MaxCPU), node.Node)
			clusterCPUs += *node.MaxCPU
		}
		if node.MaxMem != nil {
			ch <- prometheus.MustNewConstMetric(c.nodeMemTotal, prometheus.GaugeValue, float64(*node.MaxMem), node.Node)
			clusterMem += *node.MaxMem
		}
	}

	// Process guest metrics from cluster resources
	vmMetrics := c.collectVirtualMachineMetrics(ch, qemuResources)
	lxcMetrics := c.collectLxcMetrics(ch, lxcResources)

	// Combine VM + LXC allocations per node
	clusterCPUsAlloc := 0
	clusterMemAlloc := 0
	allocNodes := make(map[string]bool)
	for node := range vmMetrics.cpusPerNode {
		allocNodes[node] = true
	}
	for node := range lxcMetrics.cpusPerNode {
		allocNodes[node] = true
	}
	for node := range allocNodes {
		cpus := vmMetrics.cpusPerNode[node] + lxcMetrics.cpusPerNode[node]
		mem := vmMetrics.memPerNode[node] + lxcMetrics.memPerNode[node]
		ch <- prometheus.MustNewConstMetric(c.nodeCPUsAlloc, prometheus.GaugeValue, float64(cpus), node)
		ch <- prometheus.MustNewConstMetric(c.nodeMemAlloc, prometheus.GaugeValue, float64(mem), node)
		clusterCPUsAlloc += cpus
		clusterMemAlloc += mem
	}

	// Process storage metrics from cluster resources
	c.collectStorageMetrics(ch, storageResources)

	// Emit cluster-level metrics
	ch <- prometheus.MustNewConstMetric(c.clusterCPUsTotal, prometheus.GaugeValue, float64(clusterCPUs))
	ch <- prometheus.MustNewConstMetric(c.clusterCPUsAlloc, prometheus.GaugeValue, float64(clusterCPUsAlloc))
	ch <- prometheus.MustNewConstMetric(c.clusterMemTotal, prometheus.GaugeValue, float64(clusterMem))
	ch <- prometheus.MustNewConstMetric(c.clusterMemAlloc, prometheus.GaugeValue, float64(clusterMemAlloc))

	// Per-node API calls for data not available in cluster resources (disk SMART, certs, PVE version)
	var wg sync.WaitGroup
	for _, nodeName := range onlineNodes {
		wg.Add(1)
		go c.collectNodeSpecificMetrics(ch, nodeName, &wg)
	}
	wg.Wait()
}
