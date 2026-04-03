package prometheus

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	proxmox "github.com/starttoaster/go-proxmox"
)

func TestCollectVirtualMachineMetrics(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	intPtr := func(v int) *int { return &v }
	strPtr := func(v string) *string { return &v }
	vmidPtr := func(v string) *proxmox.IntOrString {
		ios := proxmox.IntOrString(v)
		return &ios
	}

	tests := []struct {
		name          string
		vms           []proxmox.GetClusterResourcesData
		expectedCount int
		expectedCPUs  map[string]int
		expectedMem   map[string]int
	}{
		{
			name: "single running VM",
			vms: []proxmox.GetClusterResourcesData{
				{
					Node:   "node1",
					Status: "running",
					Name:   strPtr("vm1"),
					VMID:   vmidPtr("100"),
					MaxCPU: intPtr(4),
					MaxMem: intPtr(4096),
				},
			},
			expectedCount: 1,
			expectedCPUs:  map[string]int{"node1": 4},
			expectedMem:   map[string]int{"node1": 4096},
		},
		{
			name: "single stopped VM",
			vms: []proxmox.GetClusterResourcesData{
				{
					Node:   "node1",
					Status: "stopped",
					Name:   strPtr("vm2"),
					VMID:   vmidPtr("101"),
					MaxCPU: intPtr(2),
					MaxMem: intPtr(2048),
				},
			},
			expectedCount: 1,
			expectedCPUs:  map[string]int{"node1": 2},
			expectedMem:   map[string]int{"node1": 2048},
		},
		{
			name: "template VM excluded",
			vms: []proxmox.GetClusterResourcesData{
				{
					Node:     "node1",
					Status:   "stopped",
					Name:     strPtr("template-vm"),
					VMID:     vmidPtr("900"),
					Template: intPtr(1),
					MaxCPU:   intPtr(2),
					MaxMem:   intPtr(2048),
				},
			},
			expectedCount: 0,
			expectedCPUs:  map[string]int{},
			expectedMem:   map[string]int{},
		},
		{
			name: "template=0 is not excluded",
			vms: []proxmox.GetClusterResourcesData{
				{
					Node:     "node1",
					Status:   "running",
					Name:     strPtr("normal-vm"),
					VMID:     vmidPtr("102"),
					Template: intPtr(0),
					MaxCPU:   intPtr(2),
					MaxMem:   intPtr(2048),
				},
			},
			expectedCount: 1,
			expectedCPUs:  map[string]int{"node1": 2},
			expectedMem:   map[string]int{"node1": 2048},
		},
		{
			name: "nil template is not excluded",
			vms: []proxmox.GetClusterResourcesData{
				{
					Node:   "node1",
					Status: "running",
					Name:   strPtr("vm3"),
					VMID:   vmidPtr("103"),
					MaxCPU: intPtr(4),
					MaxMem: intPtr(8192),
				},
			},
			expectedCount: 1,
			expectedCPUs:  map[string]int{"node1": 4},
			expectedMem:   map[string]int{"node1": 8192},
		},
		{
			name: "multiple VMs on same node",
			vms: []proxmox.GetClusterResourcesData{
				{
					Node:   "node1",
					Status: "running",
					Name:   strPtr("vm1"),
					VMID:   vmidPtr("100"),
					MaxCPU: intPtr(4),
					MaxMem: intPtr(4096),
				},
				{
					Node:   "node1",
					Status: "running",
					Name:   strPtr("vm2"),
					VMID:   vmidPtr("101"),
					MaxCPU: intPtr(2),
					MaxMem: intPtr(2048),
				},
			},
			expectedCount: 2,
			expectedCPUs:  map[string]int{"node1": 6},
			expectedMem:   map[string]int{"node1": 6144},
		},
		{
			name: "VMs across multiple nodes",
			vms: []proxmox.GetClusterResourcesData{
				{
					Node:   "node1",
					Status: "running",
					Name:   strPtr("vm1"),
					VMID:   vmidPtr("100"),
					MaxCPU: intPtr(4),
					MaxMem: intPtr(4096),
				},
				{
					Node:   "node2",
					Status: "running",
					Name:   strPtr("vm2"),
					VMID:   vmidPtr("101"),
					MaxCPU: intPtr(8),
					MaxMem: intPtr(16384),
				},
			},
			expectedCount: 2,
			expectedCPUs:  map[string]int{"node1": 4, "node2": 8},
			expectedMem:   map[string]int{"node1": 4096, "node2": 16384},
		},
		{
			name: "VM with nil optional fields",
			vms: []proxmox.GetClusterResourcesData{
				{
					Node:   "node1",
					Status: "stopped",
				},
			},
			expectedCount: 1,
			expectedCPUs:  map[string]int{},
			expectedMem:   map[string]int{},
		},
		{
			name:          "no VMs",
			vms:           []proxmox.GetClusterResourcesData{},
			expectedCount: 0,
			expectedCPUs:  map[string]int{},
			expectedMem:   map[string]int{},
		},
		{
			name: "mix of templates and real VMs",
			vms: []proxmox.GetClusterResourcesData{
				{
					Node:     "node1",
					Status:   "running",
					Name:     strPtr("vm1"),
					VMID:     vmidPtr("100"),
					MaxCPU:   intPtr(4),
					MaxMem:   intPtr(4096),
					Template: intPtr(0),
				},
				{
					Node:     "node1",
					Status:   "stopped",
					Name:     strPtr("template"),
					VMID:     vmidPtr("900"),
					MaxCPU:   intPtr(2),
					MaxMem:   intPtr(2048),
					Template: intPtr(1),
				},
				{
					Node:   "node1",
					Status: "running",
					Name:   strPtr("vm2"),
					VMID:   vmidPtr("101"),
					MaxCPU: intPtr(2),
					MaxMem: intPtr(2048),
				},
			},
			expectedCount: 2,
			expectedCPUs:  map[string]int{"node1": 6},
			expectedMem:   map[string]int{"node1": 6144},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testCollector()
			ch := make(chan prometheus.Metric, 100)

			res := c.collectVirtualMachineMetrics(ch, tt.vms)

			metrics := drainMetrics(ch)
			if len(metrics) != tt.expectedCount {
				t.Fatalf("expected %d metrics, got %d", tt.expectedCount, len(metrics))
			}

			for node, expectedCPU := range tt.expectedCPUs {
				if res.cpusPerNode[node] != expectedCPU {
					t.Errorf("node %s: expected CPU alloc %d, got %d", node, expectedCPU, res.cpusPerNode[node])
				}
			}
			for node, expectedMem := range tt.expectedMem {
				if res.memPerNode[node] != expectedMem {
					t.Errorf("node %s: expected mem alloc %d, got %d", node, expectedMem, res.memPerNode[node])
				}
			}
		})
	}
}

func TestCollectVirtualMachineMetrics_GuestUpLabels(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	strPtr := func(v string) *string { return &v }
	vmidPtr := func(v string) *proxmox.IntOrString {
		ios := proxmox.IntOrString(v)
		return &ios
	}

	c := testCollector()
	ch := make(chan prometheus.Metric, 10)

	vms := []proxmox.GetClusterResourcesData{
		{
			Node:   "node1",
			Status: "running",
			Name:   strPtr("web-server"),
			VMID:   vmidPtr("100"),
			Tags:   strPtr("prod;web"),
		},
	}

	c.collectVirtualMachineMetrics(ch, vms)

	metrics := drainMetrics(ch)
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	labels := getMetricLabels(metrics[0])
	if labels["node"] != "node1" {
		t.Errorf("expected node=node1, got %s", labels["node"])
	}
	if labels["type"] != "qemu" {
		t.Errorf("expected type=qemu, got %s", labels["type"])
	}
	if labels["name"] != "web-server" {
		t.Errorf("expected name=web-server, got %s", labels["name"])
	}
	if labels["vmid"] != "100" {
		t.Errorf("expected vmid=100, got %s", labels["vmid"])
	}
	if labels["tags"] != "prod;web" {
		t.Errorf("expected tags=prod;web, got %s", labels["tags"])
	}

	v := getMetricValue(metrics[0])
	if v != 1.0 {
		t.Errorf("expected guest_up=1 for running VM, got %f", v)
	}
}

func TestCollectVirtualMachineMetrics_StoppedValue(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	strPtr := func(v string) *string { return &v }
	vmidPtr := func(v string) *proxmox.IntOrString {
		ios := proxmox.IntOrString(v)
		return &ios
	}

	c := testCollector()
	ch := make(chan prometheus.Metric, 10)

	vms := []proxmox.GetClusterResourcesData{
		{
			Node:   "node1",
			Status: "stopped",
			Name:   strPtr("vm-off"),
			VMID:   vmidPtr("102"),
		},
	}

	c.collectVirtualMachineMetrics(ch, vms)

	metrics := drainMetrics(ch)
	v := getMetricValue(metrics[0])
	if v != 0.0 {
		t.Errorf("expected guest_up=0 for stopped VM, got %f", v)
	}
}

func TestCollectVirtualMachineMetrics_NilName(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	c := testCollector()
	ch := make(chan prometheus.Metric, 10)

	vms := []proxmox.GetClusterResourcesData{
		{
			Node:   "node1",
			Status: "running",
		},
	}

	c.collectVirtualMachineMetrics(ch, vms)

	metrics := drainMetrics(ch)
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	labels := getMetricLabels(metrics[0])
	if labels["name"] != "" {
		t.Errorf("expected empty name for nil Name, got %s", labels["name"])
	}
}
