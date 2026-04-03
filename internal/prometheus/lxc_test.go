package prometheus

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	proxmox "github.com/starttoaster/go-proxmox"
)

func TestCollectLxcMetrics(t *testing.T) {
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
		lxcs          []proxmox.GetClusterResourcesData
		expectedCount int
		expectedCPUs  map[string]int
		expectedMem   map[string]int
	}{
		{
			name: "single running LXC",
			lxcs: []proxmox.GetClusterResourcesData{
				{
					Node:   "node1",
					Status: "running",
					Name:   strPtr("lxc1"),
					VMID:   vmidPtr("200"),
					MaxCPU: intPtr(2),
					MaxMem: intPtr(1024),
				},
			},
			expectedCount: 1,
			expectedCPUs:  map[string]int{"node1": 2},
			expectedMem:   map[string]int{"node1": 1024},
		},
		{
			name: "single stopped LXC",
			lxcs: []proxmox.GetClusterResourcesData{
				{
					Node:   "node1",
					Status: "stopped",
					Name:   strPtr("lxc2"),
					VMID:   vmidPtr("201"),
					MaxCPU: intPtr(1),
					MaxMem: intPtr(512),
				},
			},
			expectedCount: 1,
			expectedCPUs:  map[string]int{"node1": 1},
			expectedMem:   map[string]int{"node1": 512},
		},
		{
			name: "multiple LXCs on same node",
			lxcs: []proxmox.GetClusterResourcesData{
				{
					Node:   "node1",
					Status: "running",
					Name:   strPtr("lxc1"),
					VMID:   vmidPtr("200"),
					MaxCPU: intPtr(2),
					MaxMem: intPtr(1024),
				},
				{
					Node:   "node1",
					Status: "running",
					Name:   strPtr("lxc2"),
					VMID:   vmidPtr("201"),
					MaxCPU: intPtr(4),
					MaxMem: intPtr(4096),
				},
			},
			expectedCount: 2,
			expectedCPUs:  map[string]int{"node1": 6},
			expectedMem:   map[string]int{"node1": 5120},
		},
		{
			name: "LXCs across multiple nodes",
			lxcs: []proxmox.GetClusterResourcesData{
				{
					Node:   "node1",
					Status: "running",
					Name:   strPtr("lxc1"),
					VMID:   vmidPtr("200"),
					MaxCPU: intPtr(2),
					MaxMem: intPtr(1024),
				},
				{
					Node:   "node2",
					Status: "running",
					Name:   strPtr("lxc2"),
					VMID:   vmidPtr("201"),
					MaxCPU: intPtr(4),
					MaxMem: intPtr(8192),
				},
			},
			expectedCount: 2,
			expectedCPUs:  map[string]int{"node1": 2, "node2": 4},
			expectedMem:   map[string]int{"node1": 1024, "node2": 8192},
		},
		{
			name: "LXC with nil optional fields",
			lxcs: []proxmox.GetClusterResourcesData{
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
			name:          "no LXCs",
			lxcs:          []proxmox.GetClusterResourcesData{},
			expectedCount: 0,
			expectedCPUs:  map[string]int{},
			expectedMem:   map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testCollector()
			ch := make(chan prometheus.Metric, 100)

			res := c.collectLxcMetrics(ch, tt.lxcs)

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

func TestCollectLxcMetrics_GuestUpLabels(t *testing.T) {
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

	lxcs := []proxmox.GetClusterResourcesData{
		{
			Node:   "node1",
			Status: "running",
			Name:   strPtr("dns-server"),
			VMID:   vmidPtr("200"),
			Tags:   strPtr("infra;dns"),
		},
	}

	c.collectLxcMetrics(ch, lxcs)

	metrics := drainMetrics(ch)
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	labels := getMetricLabels(metrics[0])
	if labels["node"] != "node1" {
		t.Errorf("expected node=node1, got %s", labels["node"])
	}
	if labels["type"] != "lxc" {
		t.Errorf("expected type=lxc, got %s", labels["type"])
	}
	if labels["name"] != "dns-server" {
		t.Errorf("expected name=dns-server, got %s", labels["name"])
	}
	if labels["vmid"] != "200" {
		t.Errorf("expected vmid=200, got %s", labels["vmid"])
	}
	if labels["tags"] != "infra;dns" {
		t.Errorf("expected tags=infra;dns, got %s", labels["tags"])
	}

	v := getMetricValue(metrics[0])
	if v != 1.0 {
		t.Errorf("expected guest_up=1 for running LXC, got %f", v)
	}
}

func TestCollectLxcMetrics_StoppedValue(t *testing.T) {
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

	lxcs := []proxmox.GetClusterResourcesData{
		{
			Node:   "node1",
			Status: "stopped",
			Name:   strPtr("lxc-off"),
			VMID:   vmidPtr("202"),
		},
	}

	c.collectLxcMetrics(ch, lxcs)

	metrics := drainMetrics(ch)
	v := getMetricValue(metrics[0])
	if v != 0.0 {
		t.Errorf("expected guest_up=0 for stopped LXC, got %f", v)
	}
}

func TestCollectLxcMetrics_NilFields(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	c := testCollector()
	ch := make(chan prometheus.Metric, 10)

	lxcs := []proxmox.GetClusterResourcesData{
		{
			Node:   "node1",
			Status: "running",
		},
	}

	c.collectLxcMetrics(ch, lxcs)

	metrics := drainMetrics(ch)
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	labels := getMetricLabels(metrics[0])
	if labels["name"] != "" {
		t.Errorf("expected empty name for nil Name, got %s", labels["name"])
	}
	if labels["tags"] != "" {
		t.Errorf("expected empty tags for nil Tags, got %s", labels["tags"])
	}
}

func TestCollectLxcMetrics_ResponseStructure(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	c := testCollector()
	ch := make(chan prometheus.Metric, 100)

	res := c.collectLxcMetrics(ch, []proxmox.GetClusterResourcesData{})

	if res == nil {
		t.Fatal("expected non-nil response")
	}
	if res.cpusPerNode == nil {
		t.Error("expected non-nil cpusPerNode map")
	}
	if res.memPerNode == nil {
		t.Error("expected non-nil memPerNode map")
	}
}
