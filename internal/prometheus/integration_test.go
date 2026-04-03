package prometheus

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	wrappedProxmox "github.com/starttoaster/proxmox-exporter/internal/proxmox"
)

const intClusterStatusJSON = `{
	"data": [
		{"type": "cluster", "id": "cluster", "name": "integration-cluster", "version": 5, "quorate": 1, "nodes": 2},
		{"type": "node", "id": "node/node1", "name": "node1", "online": 1, "nodeid": 1, "ip": "10.0.0.1"},
		{"type": "node", "id": "node/node2", "name": "node2", "online": 1, "nodeid": 2, "ip": "10.0.0.2"}
	]
}`

const intClusterResourcesJSON = `{
	"data": [
		{"id": "node/node1", "node": "node1", "type": "node", "status": "online", "maxcpu": 8, "maxmem": 16384000000},
		{"id": "node/node2", "node": "node2", "type": "node", "status": "online", "maxcpu": 16, "maxmem": 32768000000},
		{"id": "qemu/100", "node": "node1", "type": "qemu", "status": "running", "name": "web-server", "vmid": 100, "maxcpu": 4, "maxmem": 8589934592, "template": 0, "tags": "prod;web"},
		{"id": "qemu/101", "node": "node2", "type": "qemu", "status": "stopped", "name": "db-server", "vmid": 101, "maxcpu": 8, "maxmem": 17179869184, "template": 0},
		{"id": "qemu/900", "node": "node1", "type": "qemu", "status": "stopped", "name": "template-vm", "vmid": 900, "maxcpu": 2, "maxmem": 2147483648, "template": 1},
		{"id": "lxc/200", "node": "node1", "type": "lxc", "status": "running", "name": "dns-server", "vmid": 200, "maxcpu": 1, "maxmem": 536870912, "tags": "infra"},
		{"id": "storage/node1/local", "node": "node1", "type": "storage", "status": "available", "storage": "local", "plugintype": "dir", "shared": 0, "maxdisk": 100000000000, "disk": 30000000000},
		{"id": "storage/node1/ceph", "node": "node1", "type": "storage", "status": "available", "storage": "ceph", "plugintype": "rbd", "shared": 1, "maxdisk": 500000000000, "disk": 200000000000},
		{"id": "storage/node2/local", "node": "node2", "type": "storage", "status": "available", "storage": "local", "plugintype": "dir", "shared": 0, "maxdisk": 200000000000, "disk": 60000000000}
	]
}`

const intNodeStatusJSON = `{
	"data": {
		"cpu": 0.05, "uptime": 100000,
		"pveversion": "pve-manager/8.1.3/bbf3993334bfa916",
		"kversion": "Linux 6.5.11-8-pve",
		"memory": {"total": 16384000000, "used": 8192000000, "free": 8192000000},
		"swap": {"total": 8589930496, "used": 0, "free": 8589930496},
		"rootfs": {"total": 100000000000, "used": 20000000000, "free": 80000000000, "avail": 75000000000},
		"cpuinfo": {"cpus": 8, "cores": 4, "sockets": 1, "model": "Test CPU", "mhz": "3600", "hvm": "1", "flags": "test", "user_hz": 100},
		"boot-info": {"mode": "efi", "secureboot": 0},
		"current-kernel": {"sysname": "Linux", "release": "6.5.11-8-pve", "version": "#1 SMP", "machine": "x86_64"},
		"ksm": {"shared": 0},
		"loadavg": ["0.5", "0.4", "0.3"],
		"idle": 0, "wait": 0.001
	}
}`

const intNodeDisksJSON = `{
	"data": [
		{"devpath": "/dev/nvme0n1", "health": "PASSED", "model": "Samsung SSD", "serial": "S123", "size": 512110190592, "type": "nvme", "vendor": "Samsung", "wwn": "test", "by_id_link": "/dev/disk/by-id/nvme", "rpm": 0, "wearout": 95, "gpt": 1, "used": "ext4"},
		{"devpath": "/dev/sda", "health": "UNKNOWN", "model": "WD HDD", "serial": "W456", "size": 2000398934016, "type": "hdd", "vendor": "WDC", "wwn": "test2", "by_id_link": "/dev/disk/by-id/ata", "rpm": "7200", "wearout": "N/A", "gpt": 0, "used": "LVM"}
	]
}`

const intQemuSnapshotsJSON = `{
	"data": [
		{"name": "snap1", "snaptime": 1700000000, "description": "First snapshot", "vmstate": 1},
		{"name": "snap2", "snaptime": 1700100000, "parent": "snap1", "description": "Second snapshot", "vmstate": 0},
		{"name": "current", "running": 1, "parent": "snap2", "description": "You are here!", "digest": "abc123"}
	]
}`

const intLxcSnapshotsJSON = `{
	"data": [
		{"name": "lxc-snap1", "snaptime": 1700000000, "description": "LXC snapshot"},
		{"name": "current", "running": 0, "parent": "lxc-snap1", "description": "You are here!", "digest": "def456"}
	]
}`

func intCertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		thirtyDays := time.Now().Add(30 * 24 * time.Hour).Unix()
		oneYear := time.Now().Add(365 * 24 * time.Hour).Unix()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"data": [
				{"filename": "pveproxy-ssl.pem", "fingerprint": "AA:BB", "issuer": "CN=PVE", "notafter": %d, "notbefore": 1600000000, "pem": "test", "public-key-bits": 2048, "public-key-type": "rsa", "san": ["test"], "subject": "CN=pveproxy"},
				{"filename": "pve-ssl.pem", "fingerprint": "CC:DD", "issuer": "CN=PVE CA", "notafter": %d, "notbefore": 1600000000, "pem": "test2", "public-key-bits": 4096, "public-key-type": "rsa", "san": ["test"], "subject": "CN=PVE CA"}
			]
		}`, thirtyDays, oneYear)
	}
}

func intJSONHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}
}

func setupIntegrationMux(withSnapshots bool) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api2/json/cluster/status", intJSONHandler(intClusterStatusJSON))
	mux.HandleFunc("/api2/json/cluster/resources", intJSONHandler(intClusterResourcesJSON))
	mux.HandleFunc("/api2/json/nodes", intJSONHandler(`{"data": [{"node": "node1", "status": "online"}, {"node": "node2", "status": "online"}]}`))
	mux.HandleFunc("/api2/json/nodes/{node}/status", intJSONHandler(intNodeStatusJSON))
	mux.HandleFunc("/api2/json/nodes/{node}/disks/list", intJSONHandler(intNodeDisksJSON))
	mux.HandleFunc("/api2/json/nodes/{node}/certificates/info", intCertHandler())

	if withSnapshots {
		mux.HandleFunc("/api2/json/nodes/{node}/qemu/{vmid}/snapshot", intJSONHandler(intQemuSnapshotsJSON))
		mux.HandleFunc("/api2/json/nodes/{node}/lxc/{vmid}/snapshot", intJSONHandler(intLxcSnapshotsJSON))
	}

	return mux
}

func initProxmoxForIntegration(t *testing.T, serverURL string) {
	t.Helper()

	oldClusterName := wrappedProxmox.ClusterName
	t.Cleanup(func() {
		wrappedProxmox.ClusterName = oldClusterName
	})

	err := wrappedProxmox.Init([]string{serverURL}, "test-id", "test-token", true)
	if err != nil {
		t.Fatalf("failed to init proxmox: %v", err)
	}
}

func countByDesc(metrics []prometheus.Metric, desc *prometheus.Desc) int {
	n := 0
	for _, m := range metrics {
		if m.Desc() == desc {
			n++
		}
	}
	return n
}

func findByDesc(metrics []prometheus.Metric, desc *prometheus.Desc) []prometheus.Metric {
	var found []prometheus.Metric
	for _, m := range metrics {
		if m.Desc() == desc {
			found = append(found, m)
		}
	}
	return found
}

func TestCollect_Integration(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	mux := setupIntegrationMux(false)
	server := httptest.NewTLSServer(mux)
	defer server.Close()

	// Override the default http transport to accept the test server's self-signed cert
	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer func() {
		http.DefaultTransport = &http.Transport{}
	}()

	initProxmoxForIntegration(t, server.URL)

	c := NewCollector()
	ch := make(chan prometheus.Metric, 1000)
	c.Collect(ch)
	metrics := drainMetrics(ch)

	// --- Verify metric counts by type ---

	// Client counts: 2 (banned + unbanned)
	if n := countByDesc(metrics, c.clientCount); n != 2 {
		t.Errorf("clientCount: expected 2, got %d", n)
	}

	// Node up: 2 (node1, node2)
	if n := countByDesc(metrics, c.nodeUp); n != 2 {
		t.Errorf("nodeUp: expected 2, got %d", n)
	}

	// Node CPUs total: 2
	if n := countByDesc(metrics, c.nodeCPUsTotal); n != 2 {
		t.Errorf("nodeCPUsTotal: expected 2, got %d", n)
	}

	// Node mem total: 2
	if n := countByDesc(metrics, c.nodeMemTotal); n != 2 {
		t.Errorf("nodeMemTotal: expected 2, got %d", n)
	}

	// Guest up: 3 (2 non-template VMs + 1 LXC; template VM excluded)
	if n := countByDesc(metrics, c.guestUp); n != 3 {
		t.Errorf("guestUp: expected 3, got %d", n)
	}

	// Node CPUs allocated: 2 (node1 and node2 both have guests)
	if n := countByDesc(metrics, c.nodeCPUsAlloc); n != 2 {
		t.Errorf("nodeCPUsAlloc: expected 2, got %d", n)
	}

	// Node mem allocated: 2
	if n := countByDesc(metrics, c.nodeMemAlloc); n != 2 {
		t.Errorf("nodeMemAlloc: expected 2, got %d", n)
	}

	// Storage total: 3 entries
	if n := countByDesc(metrics, c.storageTotal); n != 3 {
		t.Errorf("storageTotal: expected 3, got %d", n)
	}

	// Storage used: 3 entries
	if n := countByDesc(metrics, c.storageUsed); n != 3 {
		t.Errorf("storageUsed: expected 3, got %d", n)
	}

	// Cluster-level: 1 each
	if n := countByDesc(metrics, c.clusterCPUsTotal); n != 1 {
		t.Errorf("clusterCPUsTotal: expected 1, got %d", n)
	}
	if n := countByDesc(metrics, c.clusterCPUsAlloc); n != 1 {
		t.Errorf("clusterCPUsAlloc: expected 1, got %d", n)
	}
	if n := countByDesc(metrics, c.clusterMemTotal); n != 1 {
		t.Errorf("clusterMemTotal: expected 1, got %d", n)
	}
	if n := countByDesc(metrics, c.clusterMemAlloc); n != 1 {
		t.Errorf("clusterMemAlloc: expected 1, got %d", n)
	}

	// Per-node specific metrics (2 online nodes)
	// Node version: 2
	if n := countByDesc(metrics, c.nodeVersion); n != 2 {
		t.Errorf("nodeVersion: expected 2, got %d", n)
	}

	// Disk SMART: 4 (2 disks × 2 nodes)
	if n := countByDesc(metrics, c.diskSmartHealth); n != 4 {
		t.Errorf("diskSmartHealth: expected 4, got %d", n)
	}

	// Cert expiry: 4 (2 certs × 2 nodes)
	if n := countByDesc(metrics, c.daysUntilCertExpiry); n != 4 {
		t.Errorf("daysUntilCertExpiry: expected 4, got %d", n)
	}

	// Snapshot metrics should NOT be present
	if n := countByDesc(metrics, c.guestSnapshotsCount); n != 0 {
		t.Errorf("guestSnapshotsCount: expected 0, got %d", n)
	}

	// --- Spot-check specific values ---

	// Cluster CPUs total = 8 + 16 = 24
	clusterCPUs := findByDesc(metrics, c.clusterCPUsTotal)
	if len(clusterCPUs) == 1 {
		if v := getMetricValue(clusterCPUs[0]); v != 24.0 {
			t.Errorf("cluster CPUs total: expected 24, got %f", v)
		}
	}

	// Cluster CPUs allocated = (4 VM + 1 LXC on node1) + (8 VM on node2) = 13
	clusterCPUsAlloc := findByDesc(metrics, c.clusterCPUsAlloc)
	if len(clusterCPUsAlloc) == 1 {
		if v := getMetricValue(clusterCPUsAlloc[0]); v != 13.0 {
			t.Errorf("cluster CPUs allocated: expected 13, got %f", v)
		}
	}

	// Cluster mem total = 16384000000 + 32768000000 = 49152000000
	clusterMem := findByDesc(metrics, c.clusterMemTotal)
	if len(clusterMem) == 1 {
		if v := getMetricValue(clusterMem[0]); v != 49152000000.0 {
			t.Errorf("cluster mem total: expected 49152000000, got %f", v)
		}
	}

	// Cluster mem allocated = (8589934592 VM + 536870912 LXC) + 17179869184 VM = 26306674688
	clusterMemAlloc := findByDesc(metrics, c.clusterMemAlloc)
	if len(clusterMemAlloc) == 1 {
		if v := getMetricValue(clusterMemAlloc[0]); v != 26306674688.0 {
			t.Errorf("cluster mem allocated: expected 26306674688, got %f", v)
		}
	}

	// Node up values
	nodeUps := findByDesc(metrics, c.nodeUp)
	for _, m := range nodeUps {
		v := getMetricValue(m)
		if v != 1.0 {
			t.Errorf("all nodes should be online (1.0), got %f", v)
		}
	}

	// Guest up values: check for running vs stopped
	guestUps := findByDesc(metrics, c.guestUp)
	runningCount := 0
	stoppedCount := 0
	for _, m := range guestUps {
		v := getMetricValue(m)
		if v == 1.0 {
			runningCount++
		} else {
			stoppedCount++
		}
	}
	if runningCount != 2 {
		t.Errorf("expected 2 running guests, got %d", runningCount)
	}
	if stoppedCount != 1 {
		t.Errorf("expected 1 stopped guest, got %d", stoppedCount)
	}

	// Disk SMART values: each node gets 1 PASSED (1.0) + 1 UNKNOWN (-1.0)
	disks := findByDesc(metrics, c.diskSmartHealth)
	passedCount := 0
	unknownCount := 0
	for _, m := range disks {
		v := getMetricValue(m)
		switch v {
		case 1.0:
			passedCount++
		case -1.0:
			unknownCount++
		}
	}
	if passedCount != 2 {
		t.Errorf("expected 2 PASSED disks, got %d", passedCount)
	}
	if unknownCount != 2 {
		t.Errorf("expected 2 UNKNOWN disks, got %d", unknownCount)
	}

	// Cert expiry: should have values around 29-30 days and 364-365 days
	certs := findByDesc(metrics, c.daysUntilCertExpiry)
	shortExpiry := 0
	longExpiry := 0
	for _, m := range certs {
		v := getMetricValue(m)
		if v >= 29 && v <= 31 {
			shortExpiry++
		} else if v >= 363 && v <= 366 {
			longExpiry++
		}
	}
	if shortExpiry != 2 {
		t.Errorf("expected 2 certs expiring in ~30 days, got %d", shortExpiry)
	}
	if longExpiry != 2 {
		t.Errorf("expected 2 certs expiring in ~365 days, got %d", longExpiry)
	}

	// Node version labels
	versions := findByDesc(metrics, c.nodeVersion)
	for _, m := range versions {
		labels := getMetricLabels(m)
		if labels["version"] != "pve-manager/8.1.3/bbf3993334bfa916" {
			t.Errorf("unexpected PVE version label: %s", labels["version"])
		}
	}

	// Total metric count
	expectedTotal := 35
	if len(metrics) != expectedTotal {
		t.Errorf("total metrics: expected %d, got %d", expectedTotal, len(metrics))
	}
}

func TestCollect_WithSnapshots_Integration(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: true}
	defer func() { cfg = oldCfg }()

	mux := setupIntegrationMux(true)
	server := httptest.NewTLSServer(mux)
	defer server.Close()

	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer func() {
		http.DefaultTransport = &http.Transport{}
	}()

	initProxmoxForIntegration(t, server.URL)

	c := NewCollector()
	ch := make(chan prometheus.Metric, 1000)
	c.Collect(ch)
	metrics := drainMetrics(ch)

	// Snapshot count metrics: 3 (2 non-template VMs + 1 LXC)
	if n := countByDesc(metrics, c.guestSnapshotsCount); n != 3 {
		t.Errorf("guestSnapshotsCount: expected 3, got %d", n)
	}

	// Snapshot age metrics:
	// Each VM: 2 snapshots with snaptime → 2 age metrics each → 4 total
	// LXC: 1 snapshot with snaptime → 1 age metric
	// Total: 5
	if n := countByDesc(metrics, c.guestSnapshotAgeSeconds); n != 5 {
		t.Errorf("guestSnapshotAgeSeconds: expected 5, got %d", n)
	}

	// Verify snapshot counts: each VM has 3 items - 1 current = 2
	snapshotCounts := findByDesc(metrics, c.guestSnapshotsCount)
	for _, m := range snapshotCounts {
		labels := getMetricLabels(m)
		v := getMetricValue(m)
		switch labels["type"] {
		case "qemu":
			if v != 2.0 {
				t.Errorf("VM %s snapshot count: expected 2, got %f", labels["vmid"], v)
			}
		case "lxc":
			if v != 1.0 {
				t.Errorf("LXC %s snapshot count: expected 1, got %f", labels["vmid"], v)
			}
		}
	}

	// Verify snapshot age is positive (snapshots are in the past)
	snapshotAges := findByDesc(metrics, c.guestSnapshotAgeSeconds)
	for _, m := range snapshotAges {
		v := getMetricValue(m)
		if v <= 0 {
			labels := getMetricLabels(m)
			t.Errorf("snapshot %s age should be positive, got %f", labels["snapshot"], v)
		}
	}

	// Total metric count with snapshots: 35 base + 3 snapshot counts + 5 snapshot ages = 43
	expectedTotal := 43
	if len(metrics) != expectedTotal {
		t.Errorf("total metrics with snapshots: expected %d, got %d", expectedTotal, len(metrics))
	}
}

func TestCollect_Integration_ClusterLabel(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	mux := setupIntegrationMux(false)
	server := httptest.NewTLSServer(mux)
	defer server.Close()

	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer func() {
		http.DefaultTransport = &http.Transport{}
	}()

	initProxmoxForIntegration(t, server.URL)

	// After Init, ClusterName should be set from the mock cluster/status response
	if wrappedProxmox.ClusterName != "integration-cluster" {
		t.Fatalf("expected ClusterName='integration-cluster', got %q", wrappedProxmox.ClusterName)
	}

	c := NewCollector()
	ch := make(chan prometheus.Metric, 1000)
	c.Collect(ch)
	metrics := drainMetrics(ch)

	// All metrics should have the "cluster" const label
	for _, m := range metrics {
		labels := getMetricLabels(m)
		if labels["cluster"] != "integration-cluster" {
			t.Errorf("expected cluster=integration-cluster label, got %q", labels["cluster"])
			break
		}
	}
}

func TestCollect_Integration_GuestLabels(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	mux := setupIntegrationMux(false)
	server := httptest.NewTLSServer(mux)
	defer server.Close()

	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer func() {
		http.DefaultTransport = &http.Transport{}
	}()

	initProxmoxForIntegration(t, server.URL)

	c := NewCollector()
	ch := make(chan prometheus.Metric, 1000)
	c.Collect(ch)
	metrics := drainMetrics(ch)

	guestUps := findByDesc(metrics, c.guestUp)

	foundVM := false
	foundLXC := false
	for _, m := range guestUps {
		labels := getMetricLabels(m)
		if labels["type"] == "qemu" && labels["name"] == "web-server" {
			foundVM = true
			if labels["vmid"] != "100" {
				t.Errorf("expected vmid=100, got %s", labels["vmid"])
			}
			if labels["tags"] != "prod;web" {
				t.Errorf("expected tags=prod;web, got %s", labels["tags"])
			}
			if v := getMetricValue(m); v != 1.0 {
				t.Errorf("expected running VM value=1, got %f", v)
			}
		}
		if labels["type"] == "lxc" && labels["name"] == "dns-server" {
			foundLXC = true
			if labels["vmid"] != "200" {
				t.Errorf("expected vmid=200, got %s", labels["vmid"])
			}
			if labels["tags"] != "infra" {
				t.Errorf("expected tags=infra, got %s", labels["tags"])
			}
		}
	}
	if !foundVM {
		t.Error("running VM 'web-server' not found in guest_up metrics")
	}
	if !foundLXC {
		t.Error("running LXC 'dns-server' not found in guest_up metrics")
	}
}

func TestCollect_Integration_StorageLabels(t *testing.T) {
	oldCfg := cfg
	cfg = Config{EnableSnapshotMetrics: false}
	defer func() { cfg = oldCfg }()

	mux := setupIntegrationMux(false)
	server := httptest.NewTLSServer(mux)
	defer server.Close()

	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer func() {
		http.DefaultTransport = &http.Transport{}
	}()

	initProxmoxForIntegration(t, server.URL)

	c := NewCollector()
	ch := make(chan prometheus.Metric, 1000)
	c.Collect(ch)
	metrics := drainMetrics(ch)

	storageTotals := findByDesc(metrics, c.storageTotal)

	foundShared := false
	foundLocal := false
	for _, m := range storageTotals {
		labels := getMetricLabels(m)
		if labels["storage"] == "ceph" {
			foundShared = true
			if labels["type"] != "rbd" {
				t.Errorf("expected ceph type=rbd, got %s", labels["type"])
			}
			if labels["shared"] != "true" {
				t.Errorf("expected shared=true for ceph, got %s", labels["shared"])
			}
			if v := getMetricValue(m); v != 500000000000.0 {
				t.Errorf("expected ceph total=500000000000, got %f", v)
			}
		}
		if labels["storage"] == "local" && labels["node"] == "node1" {
			foundLocal = true
			if labels["type"] != "dir" {
				t.Errorf("expected local type=dir, got %s", labels["type"])
			}
			if labels["shared"] != "false" {
				t.Errorf("expected shared=false for local, got %s", labels["shared"])
			}
			if v := getMetricValue(m); v != 100000000000.0 {
				t.Errorf("expected local total=100000000000, got %f", v)
			}
		}
	}
	if !foundShared {
		t.Error("shared ceph storage not found")
	}
	if !foundLocal {
		t.Error("local storage on node1 not found")
	}
}
