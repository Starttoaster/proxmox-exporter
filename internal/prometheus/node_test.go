package prometheus

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	proxmox "github.com/starttoaster/go-proxmox"
)

func TestCollectNodeUpMetric(t *testing.T) {
	tests := []struct {
		name          string
		node          proxmox.GetClusterResourcesData
		expectedValue float64
		expectedNode  string
	}{
		{
			name:          "online node",
			node:          proxmox.GetClusterResourcesData{Node: "node1", Status: "online"},
			expectedValue: 1.0,
			expectedNode:  "node1",
		},
		{
			name:          "offline node",
			node:          proxmox.GetClusterResourcesData{Node: "node2", Status: "offline"},
			expectedValue: 0.0,
			expectedNode:  "node2",
		},
		{
			name:          "unknown status",
			node:          proxmox.GetClusterResourcesData{Node: "node3", Status: "unknown"},
			expectedValue: 0.0,
			expectedNode:  "node3",
		},
		{
			name:          "online case insensitive",
			node:          proxmox.GetClusterResourcesData{Node: "node4", Status: "Online"},
			expectedValue: 1.0,
			expectedNode:  "node4",
		},
		{
			name:          "ONLINE uppercase",
			node:          proxmox.GetClusterResourcesData{Node: "node5", Status: "ONLINE"},
			expectedValue: 1.0,
			expectedNode:  "node5",
		},
		{
			name:          "empty status",
			node:          proxmox.GetClusterResourcesData{Node: "node6", Status: ""},
			expectedValue: 0.0,
			expectedNode:  "node6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testCollector()
			ch := make(chan prometheus.Metric, 10)

			c.collectNodeUpMetric(ch, tt.node)

			metrics := drainMetrics(ch)
			if len(metrics) != 1 {
				t.Fatalf("expected 1 metric, got %d", len(metrics))
			}

			if v := getMetricValue(metrics[0]); v != tt.expectedValue {
				t.Errorf("expected value %f, got %f", tt.expectedValue, v)
			}

			labels := getMetricLabels(metrics[0])
			if labels["node"] != tt.expectedNode {
				t.Errorf("expected node=%s, got %s", tt.expectedNode, labels["node"])
			}
		})
	}
}

func TestCollectDiskMetrics(t *testing.T) {
	tests := []struct {
		name           string
		nodeName       string
		disks          *proxmox.GetNodeDisksListResponse
		expectedCount  int
		expectedValues []float64
	}{
		{
			name:     "PASSED disk",
			nodeName: "node1",
			disks: &proxmox.GetNodeDisksListResponse{
				Data: []proxmox.GetNodeDisksListData{
					{DevPath: "/dev/sda", Health: "PASSED"},
				},
			},
			expectedCount:  1,
			expectedValues: []float64{1.0},
		},
		{
			name:     "OK disk",
			nodeName: "node1",
			disks: &proxmox.GetNodeDisksListResponse{
				Data: []proxmox.GetNodeDisksListData{
					{DevPath: "/dev/sda", Health: "OK"},
				},
			},
			expectedCount:  1,
			expectedValues: []float64{1.0},
		},
		{
			name:     "FAILED disk",
			nodeName: "node1",
			disks: &proxmox.GetNodeDisksListResponse{
				Data: []proxmox.GetNodeDisksListData{
					{DevPath: "/dev/sda", Health: "FAILED"},
				},
			},
			expectedCount:  1,
			expectedValues: []float64{0.0},
		},
		{
			name:     "UNKNOWN disk",
			nodeName: "node1",
			disks: &proxmox.GetNodeDisksListResponse{
				Data: []proxmox.GetNodeDisksListData{
					{DevPath: "/dev/sda", Health: "UNKNOWN"},
				},
			},
			expectedCount:  1,
			expectedValues: []float64{-1.0},
		},
		{
			name:     "case insensitive passed",
			nodeName: "node1",
			disks: &proxmox.GetNodeDisksListResponse{
				Data: []proxmox.GetNodeDisksListData{
					{DevPath: "/dev/sda", Health: "passed"},
				},
			},
			expectedCount:  1,
			expectedValues: []float64{1.0},
		},
		{
			name:     "case insensitive ok",
			nodeName: "node1",
			disks: &proxmox.GetNodeDisksListResponse{
				Data: []proxmox.GetNodeDisksListData{
					{DevPath: "/dev/sda", Health: "ok"},
				},
			},
			expectedCount:  1,
			expectedValues: []float64{1.0},
		},
		{
			name:     "case insensitive unknown",
			nodeName: "node1",
			disks: &proxmox.GetNodeDisksListResponse{
				Data: []proxmox.GetNodeDisksListData{
					{DevPath: "/dev/sda", Health: "unknown"},
				},
			},
			expectedCount:  1,
			expectedValues: []float64{-1.0},
		},
		{
			name:     "empty health treated as FAIL",
			nodeName: "node1",
			disks: &proxmox.GetNodeDisksListResponse{
				Data: []proxmox.GetNodeDisksListData{
					{DevPath: "/dev/sda", Health: ""},
				},
			},
			expectedCount:  1,
			expectedValues: []float64{0.0},
		},
		{
			name:     "multiple disks mixed health",
			nodeName: "node1",
			disks: &proxmox.GetNodeDisksListResponse{
				Data: []proxmox.GetNodeDisksListData{
					{DevPath: "/dev/sda", Health: "PASSED"},
					{DevPath: "/dev/sdb", Health: "FAILED"},
					{DevPath: "/dev/sdc", Health: "UNKNOWN"},
					{DevPath: "/dev/sdd", Health: "OK"},
				},
			},
			expectedCount:  4,
			expectedValues: []float64{1.0, 0.0, -1.0, 1.0},
		},
		{
			name:     "no disks",
			nodeName: "node1",
			disks: &proxmox.GetNodeDisksListResponse{
				Data: []proxmox.GetNodeDisksListData{},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testCollector()
			ch := make(chan prometheus.Metric, 100)

			c.collectDiskMetrics(ch, tt.nodeName, tt.disks)

			metrics := drainMetrics(ch)
			if len(metrics) != tt.expectedCount {
				t.Fatalf("expected %d metrics, got %d", tt.expectedCount, len(metrics))
			}

			for i, m := range metrics {
				if i < len(tt.expectedValues) {
					if v := getMetricValue(m); v != tt.expectedValues[i] {
						t.Errorf("metric %d: expected value %f, got %f", i, tt.expectedValues[i], v)
					}
				}

				labels := getMetricLabels(m)
				if labels["node"] != tt.nodeName {
					t.Errorf("metric %d: expected node=%s, got %s", i, tt.nodeName, labels["node"])
				}
			}
		})
	}
}

func TestCollectDiskMetrics_Labels(t *testing.T) {
	c := testCollector()
	ch := make(chan prometheus.Metric, 10)

	disks := &proxmox.GetNodeDisksListResponse{
		Data: []proxmox.GetNodeDisksListData{
			{DevPath: "/dev/sda", Health: "PASSED"},
		},
	}

	c.collectDiskMetrics(ch, "pve-node1", disks)

	metrics := drainMetrics(ch)
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	labels := getMetricLabels(metrics[0])
	if labels["node"] != "pve-node1" {
		t.Errorf("expected node=pve-node1, got %s", labels["node"])
	}
	if labels["devpath"] != "/dev/sda" {
		t.Errorf("expected devpath=/dev/sda, got %s", labels["devpath"])
	}
}

func TestCollectCertificateMetrics(t *testing.T) {
	futureTime := int(time.Now().Add(30 * 24 * time.Hour).Unix())
	pastTime := int(time.Now().Add(-1 * 24 * time.Hour).Unix())

	tests := []struct {
		name          string
		nodeName      string
		certs         *proxmox.GetNodeCertificatesInfoResponse
		expectedCount int
	}{
		{
			name:     "single cert future expiry",
			nodeName: "node1",
			certs: &proxmox.GetNodeCertificatesInfoResponse{
				Data: []proxmox.GetNodeCertificatesInfoData{
					{Subject: "CN=pveproxy", NotAfter: futureTime},
				},
			},
			expectedCount: 1,
		},
		{
			name:     "single cert past expiry",
			nodeName: "node1",
			certs: &proxmox.GetNodeCertificatesInfoResponse{
				Data: []proxmox.GetNodeCertificatesInfoData{
					{Subject: "CN=expired", NotAfter: pastTime},
				},
			},
			expectedCount: 1,
		},
		{
			name:     "multiple certs",
			nodeName: "node1",
			certs: &proxmox.GetNodeCertificatesInfoResponse{
				Data: []proxmox.GetNodeCertificatesInfoData{
					{Subject: "CN=cert1", NotAfter: futureTime},
					{Subject: "CN=cert2", NotAfter: pastTime},
					{Subject: "CN=cert3", NotAfter: futureTime},
				},
			},
			expectedCount: 3,
		},
		{
			name:     "no certs",
			nodeName: "node1",
			certs: &proxmox.GetNodeCertificatesInfoResponse{
				Data: []proxmox.GetNodeCertificatesInfoData{},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testCollector()
			ch := make(chan prometheus.Metric, 100)

			c.collectCertificateMetrics(ch, tt.nodeName, tt.certs)

			metrics := drainMetrics(ch)
			if len(metrics) != tt.expectedCount {
				t.Fatalf("expected %d metrics, got %d", tt.expectedCount, len(metrics))
			}

			for _, m := range metrics {
				labels := getMetricLabels(m)
				if labels["node"] != tt.nodeName {
					t.Errorf("expected node=%s, got %s", tt.nodeName, labels["node"])
				}
			}
		})
	}
}

func TestCollectCertificateMetrics_Values(t *testing.T) {
	c := testCollector()
	ch := make(chan prometheus.Metric, 10)

	futureTime := int(time.Now().Add(30 * 24 * time.Hour).Unix())
	certs := &proxmox.GetNodeCertificatesInfoResponse{
		Data: []proxmox.GetNodeCertificatesInfoData{
			{Subject: "CN=pveproxy", NotAfter: futureTime},
		},
	}

	c.collectCertificateMetrics(ch, "node1", certs)

	metrics := drainMetrics(ch)
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	v := getMetricValue(metrics[0])
	if v < 29 || v > 30 {
		t.Errorf("expected ~30 days until expiry, got %f", v)
	}

	labels := getMetricLabels(metrics[0])
	if labels["subject"] != "CN=pveproxy" {
		t.Errorf("expected subject=CN=pveproxy, got %s", labels["subject"])
	}
}

func TestCollectCertificateMetrics_ExpiredCert(t *testing.T) {
	c := testCollector()
	ch := make(chan prometheus.Metric, 10)

	pastTime := int(time.Now().Add(-10 * 24 * time.Hour).Unix())
	certs := &proxmox.GetNodeCertificatesInfoResponse{
		Data: []proxmox.GetNodeCertificatesInfoData{
			{Subject: "CN=expired", NotAfter: pastTime},
		},
	}

	c.collectCertificateMetrics(ch, "node1", certs)

	metrics := drainMetrics(ch)
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	v := getMetricValue(metrics[0])
	if v > -9 || v < -11 {
		t.Errorf("expected ~-10 days for expired cert, got %f", v)
	}
}

func TestCollectStorageMetrics(t *testing.T) {
	intPtr := func(v int) *int { return &v }
	strPtr := func(v string) *string { return &v }

	tests := []struct {
		name          string
		resources     []proxmox.GetClusterResourcesData
		expectedCount int
	}{
		{
			name: "single storage fully populated",
			resources: []proxmox.GetClusterResourcesData{
				{
					Node:       "node1",
					Storage:    strPtr("local"),
					PluginType: strPtr("dir"),
					Shared:     intPtr(0),
					MaxDisk:    intPtr(100000),
					Disk:       intPtr(50000),
				},
			},
			expectedCount: 2,
		},
		{
			name: "storage with nil optional fields",
			resources: []proxmox.GetClusterResourcesData{
				{Node: "node1"},
			},
			expectedCount: 2,
		},
		{
			name: "shared storage",
			resources: []proxmox.GetClusterResourcesData{
				{
					Node:       "node1",
					Storage:    strPtr("ceph"),
					PluginType: strPtr("rbd"),
					Shared:     intPtr(1),
					MaxDisk:    intPtr(500000),
					Disk:       intPtr(200000),
				},
			},
			expectedCount: 2,
		},
		{
			name: "multiple storage entries",
			resources: []proxmox.GetClusterResourcesData{
				{
					Node:       "node1",
					Storage:    strPtr("local"),
					PluginType: strPtr("dir"),
					Shared:     intPtr(0),
					MaxDisk:    intPtr(100000),
					Disk:       intPtr(50000),
				},
				{
					Node:       "node1",
					Storage:    strPtr("ceph"),
					PluginType: strPtr("rbd"),
					Shared:     intPtr(1),
					MaxDisk:    intPtr(500000),
					Disk:       intPtr(200000),
				},
				{
					Node:       "node2",
					Storage:    strPtr("local-zfs"),
					PluginType: strPtr("zfspool"),
					Shared:     intPtr(0),
					MaxDisk:    intPtr(200000),
					Disk:       intPtr(80000),
				},
			},
			expectedCount: 6,
		},
		{
			name:          "no storage",
			resources:     []proxmox.GetClusterResourcesData{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testCollector()
			ch := make(chan prometheus.Metric, 100)

			c.collectStorageMetrics(ch, tt.resources)

			metrics := drainMetrics(ch)
			if len(metrics) != tt.expectedCount {
				t.Fatalf("expected %d metrics, got %d", tt.expectedCount, len(metrics))
			}
		})
	}
}

func TestCollectStorageMetrics_Values(t *testing.T) {
	intPtr := func(v int) *int { return &v }
	strPtr := func(v string) *string { return &v }

	c := testCollector()
	ch := make(chan prometheus.Metric, 10)

	resources := []proxmox.GetClusterResourcesData{
		{
			Node:       "node1",
			Storage:    strPtr("local"),
			PluginType: strPtr("dir"),
			Shared:     intPtr(0),
			MaxDisk:    intPtr(100000),
			Disk:       intPtr(50000),
		},
	}

	c.collectStorageMetrics(ch, resources)

	metrics := drainMetrics(ch)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	totalVal := getMetricValue(metrics[0])
	usedVal := getMetricValue(metrics[1])

	if totalVal != 100000.0 {
		t.Errorf("expected total=100000, got %f", totalVal)
	}
	if usedVal != 50000.0 {
		t.Errorf("expected used=50000, got %f", usedVal)
	}

	labels := getMetricLabels(metrics[0])
	if labels["storage"] != "local" {
		t.Errorf("expected storage=local, got %s", labels["storage"])
	}
	if labels["type"] != "dir" {
		t.Errorf("expected type=dir, got %s", labels["type"])
	}
	if labels["shared"] != "false" {
		t.Errorf("expected shared=false, got %s", labels["shared"])
	}
	if labels["node"] != "node1" {
		t.Errorf("expected node=node1, got %s", labels["node"])
	}
}

func TestCollectStorageMetrics_SharedLabel(t *testing.T) {
	intPtr := func(v int) *int { return &v }
	strPtr := func(v string) *string { return &v }

	c := testCollector()
	ch := make(chan prometheus.Metric, 10)

	resources := []proxmox.GetClusterResourcesData{
		{
			Node:       "node1",
			Storage:    strPtr("ceph"),
			PluginType: strPtr("rbd"),
			Shared:     intPtr(1),
			MaxDisk:    intPtr(500000),
			Disk:       intPtr(200000),
		},
	}

	c.collectStorageMetrics(ch, resources)

	metrics := drainMetrics(ch)
	labels := getMetricLabels(metrics[0])
	if labels["shared"] != "true" {
		t.Errorf("expected shared=true, got %s", labels["shared"])
	}
}

func TestCollectStorageMetrics_NilFields(t *testing.T) {
	c := testCollector()
	ch := make(chan prometheus.Metric, 10)

	resources := []proxmox.GetClusterResourcesData{
		{Node: "node1"},
	}

	c.collectStorageMetrics(ch, resources)

	metrics := drainMetrics(ch)
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	totalVal := getMetricValue(metrics[0])
	usedVal := getMetricValue(metrics[1])

	if totalVal != 0.0 {
		t.Errorf("expected total=0 for nil MaxDisk, got %f", totalVal)
	}
	if usedVal != 0.0 {
		t.Errorf("expected used=0 for nil Disk, got %f", usedVal)
	}

	labels := getMetricLabels(metrics[0])
	if labels["storage"] != "" {
		t.Errorf("expected empty storage for nil Storage, got %s", labels["storage"])
	}
	if labels["type"] != "" {
		t.Errorf("expected empty type for nil PluginType, got %s", labels["type"])
	}
	if labels["shared"] != "false" {
		t.Errorf("expected shared=false for nil Shared, got %s", labels["shared"])
	}
}
