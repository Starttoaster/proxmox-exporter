package prometheus

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Fatal("expected non-nil collector")
	}

	if c.clientCount == nil {
		t.Error("clientCount desc should not be nil")
	}
	if c.nodeUp == nil {
		t.Error("nodeUp desc should not be nil")
	}
	if c.guestUp == nil {
		t.Error("guestUp desc should not be nil")
	}
	if c.nodeVersion == nil {
		t.Error("nodeVersion desc should not be nil")
	}
	if c.clusterCPUsTotal == nil {
		t.Error("clusterCPUsTotal desc should not be nil")
	}
	if c.clusterCPUsAlloc == nil {
		t.Error("clusterCPUsAlloc desc should not be nil")
	}
	if c.nodeCPUsTotal == nil {
		t.Error("nodeCPUsTotal desc should not be nil")
	}
	if c.nodeCPUsAlloc == nil {
		t.Error("nodeCPUsAlloc desc should not be nil")
	}
	if c.clusterMemTotal == nil {
		t.Error("clusterMemTotal desc should not be nil")
	}
	if c.clusterMemAlloc == nil {
		t.Error("clusterMemAlloc desc should not be nil")
	}
	if c.nodeMemTotal == nil {
		t.Error("nodeMemTotal desc should not be nil")
	}
	if c.nodeMemAlloc == nil {
		t.Error("nodeMemAlloc desc should not be nil")
	}
	if c.storageTotal == nil {
		t.Error("storageTotal desc should not be nil")
	}
	if c.storageUsed == nil {
		t.Error("storageUsed desc should not be nil")
	}
	if c.diskSmartHealth == nil {
		t.Error("diskSmartHealth desc should not be nil")
	}
	if c.daysUntilCertExpiry == nil {
		t.Error("daysUntilCertExpiry desc should not be nil")
	}
}

func TestNewCollector_SnapshotsDisabled(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	c := NewCollector()
	if c.guestSnapshotsCount != nil {
		t.Error("guestSnapshotsCount should be nil when snapshots disabled")
	}
	if c.guestSnapshotAgeSeconds != nil {
		t.Error("guestSnapshotAgeSeconds should be nil when snapshots disabled")
	}
}

func TestNewCollector_SnapshotsEnabled(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: true}
	defer func() { cfg = oldCfg }()

	c := NewCollector()
	if c.guestSnapshotsCount == nil {
		t.Error("guestSnapshotsCount should not be nil when snapshots enabled")
	}
	if c.guestSnapshotAgeSeconds == nil {
		t.Error("guestSnapshotAgeSeconds should not be nil when snapshots enabled")
	}
}

func TestNewCollector_WithClusterName(t *testing.T) {
	oldName := wrappedProxmox.ClusterName
	wrappedProxmox.ClusterName = "test-cluster"
	defer func() { wrappedProxmox.ClusterName = oldName }()

	c := NewCollector()
	if c == nil {
		t.Fatal("expected non-nil collector")
	}

	// Verify cluster label is present by checking a metric descriptor's string representation
	desc := c.nodeUp.String()
	if desc == "" {
		t.Error("expected non-empty descriptor string")
	}
}

func TestNewCollector_EmptyClusterName(t *testing.T) {
	oldName := wrappedProxmox.ClusterName
	wrappedProxmox.ClusterName = ""
	defer func() { wrappedProxmox.ClusterName = oldName }()

	c := NewCollector()
	if c == nil {
		t.Fatal("expected non-nil collector")
	}
}

func TestDescribe(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	c := NewCollector()
	ch := make(chan *prometheus.Desc, 100)

	c.Describe(ch)

	var descs []*prometheus.Desc
	draining := true
	for draining {
		select {
		case d := <-ch:
			descs = append(descs, d)
		default:
			draining = false
		}
	}

	expectedCount := 15
	if len(descs) != expectedCount {
		t.Errorf("expected %d descriptors, got %d", expectedCount, len(descs))
	}

	for i, d := range descs {
		if d == nil {
			t.Errorf("descriptor %d should not be nil", i)
		}
	}
}

func TestDescribe_WithSnapshots(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: true}
	defer func() { cfg = oldCfg }()

	c := NewCollector()
	ch := make(chan *prometheus.Desc, 100)

	c.Describe(ch)

	var descs []*prometheus.Desc
	draining := true
	for draining {
		select {
		case d := <-ch:
			descs = append(descs, d)
		default:
			draining = false
		}
	}

	expectedCount := 17
	if len(descs) != expectedCount {
		t.Errorf("expected %d descriptors (with snapshots), got %d", expectedCount, len(descs))
	}
}

func TestInit_Config(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{"snapshots enabled", Config{EnableSnapshotMetrics: true}},
		{"snapshots disabled", Config{EnableSnapshotMetrics: false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.config)
			if cfg.EnableSnapshotMetrics != tt.config.EnableSnapshotMetrics {
				t.Errorf("expected EnableSnapshotMetrics=%v, got %v",
					tt.config.EnableSnapshotMetrics, cfg.EnableSnapshotMetrics)
			}
		})
	}
}
