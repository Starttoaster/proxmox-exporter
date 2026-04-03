package proxmox

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	proxmox "github.com/starttoaster/go-proxmox"
)

func jsonHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}
}

func errorHandler(code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}
}

func countingHandler(body string, counter *atomic.Int32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		counter.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}
}

func setupIntegrationTest(t *testing.T, mux *http.ServeMux) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(mux)

	c, err := proxmox.NewClient("test-id", "test-token",
		proxmox.WithBaseURL(server.URL+"/"),
		proxmox.WithHTTPClient(&http.Client{}),
	)
	if err != nil {
		t.Fatal(err)
	}

	clients = map[string]wrappedClient{
		"mock": {client: c},
	}
	cash = cache.New(24*time.Second, 5*time.Second)

	t.Cleanup(func() {
		server.Close()
		clients = nil
		cash = nil
	})

	return server
}

const integrationClusterStatusJSON = `{
	"data": [
		{"type": "cluster", "id": "cluster", "name": "test-cluster", "version": 5, "quorate": 1, "nodes": 3},
		{"type": "node", "id": "node/node1", "name": "node1", "online": 1, "nodeid": 1, "ip": "10.0.0.1"},
		{"type": "node", "id": "node/node2", "name": "node2", "online": 1, "nodeid": 2, "ip": "10.0.0.2"}
	]
}`

const integrationClusterResourcesJSON = `{
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

const integrationNodeStatusJSON = `{
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

const integrationNodeDisksJSON = `{
	"data": [
		{"devpath": "/dev/nvme0n1", "health": "PASSED", "model": "Samsung SSD", "serial": "S123", "size": 512110190592, "type": "nvme", "vendor": "Samsung", "wwn": "test", "by_id_link": "/dev/disk/by-id/nvme-samsung", "rpm": 0, "wearout": 95, "gpt": 1, "used": "ext4"},
		{"devpath": "/dev/sda", "health": "UNKNOWN", "model": "WD HDD", "serial": "W456", "size": 2000398934016, "type": "hdd", "vendor": "WDC", "wwn": "test2", "by_id_link": "/dev/disk/by-id/ata-wdc", "rpm": "7200", "wearout": "N/A", "gpt": 0, "used": "LVM"}
	]
}`

const integrationNodeCertsJSON = `{
	"data": [
		{"filename": "pveproxy-ssl.pem", "fingerprint": "AA:BB", "issuer": "CN=PVE", "notafter": 1900000000, "notbefore": 1600000000, "pem": "test", "public-key-bits": 2048, "public-key-type": "rsa", "san": ["test"], "subject": "CN=pveproxy"},
		{"filename": "pve-ssl.pem", "fingerprint": "CC:DD", "issuer": "CN=PVE CA", "notafter": 2000000000, "notbefore": 1600000000, "pem": "test2", "public-key-bits": 4096, "public-key-type": "rsa", "san": ["test"], "subject": "CN=PVE CA"}
	]
}`

const integrationQemuSnapshotsJSON = `{
	"data": [
		{"name": "snap1", "snaptime": 1700000000, "description": "First snapshot", "vmstate": 1},
		{"name": "snap2", "snaptime": 1700100000, "parent": "snap1", "description": "Second snapshot", "vmstate": 0},
		{"name": "current", "running": 1, "parent": "snap2", "description": "You are here!", "digest": "abc123"}
	]
}`

const integrationLxcSnapshotsJSON = `{
	"data": [
		{"name": "lxc-snap1", "snaptime": 1700000000, "description": "LXC snapshot"},
		{"name": "current", "running": 0, "parent": "lxc-snap1", "description": "You are here!", "digest": "def456"}
	]
}`

func TestGetClusterStatus_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/cluster/status", jsonHandler(integrationClusterStatusJSON))
	setupIntegrationTest(t, mux)

	resp, err := GetClusterStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 items, got %d", len(resp.Data))
	}

	var clusterFound bool
	var nodeCount int
	for _, d := range resp.Data {
		switch d.Type {
		case "cluster":
			clusterFound = true
			if d.Name != "test-cluster" {
				t.Errorf("expected cluster name 'test-cluster', got %s", d.Name)
			}
		case "node":
			nodeCount++
		}
	}
	if !clusterFound {
		t.Error("cluster entry not found in response")
	}
	if nodeCount != 2 {
		t.Errorf("expected 2 node entries, got %d", nodeCount)
	}
}

func TestGetClusterResources_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/cluster/resources", jsonHandler(integrationClusterResourcesJSON))
	setupIntegrationTest(t, mux)

	resp, err := GetClusterResources()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var nodes, qemu, lxc, storage int
	for _, r := range resp.Data {
		switch r.Type {
		case "node":
			nodes++
		case "qemu":
			qemu++
		case "lxc":
			lxc++
		case "storage":
			storage++
		}
	}

	if nodes != 2 {
		t.Errorf("expected 2 nodes, got %d", nodes)
	}
	if qemu != 3 {
		t.Errorf("expected 3 qemu (including template), got %d", qemu)
	}
	if lxc != 1 {
		t.Errorf("expected 1 lxc, got %d", lxc)
	}
	if storage != 3 {
		t.Errorf("expected 3 storage, got %d", storage)
	}
}

func TestGetClusterResources_FieldParsing(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/cluster/resources", jsonHandler(integrationClusterResourcesJSON))
	setupIntegrationTest(t, mux)

	resp, err := GetClusterResources()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, r := range resp.Data {
		if r.Type == "qemu" && r.Node == "node1" && r.Status == "running" {
			if r.Name == nil || *r.Name != "web-server" {
				t.Errorf("expected VM name 'web-server'")
			}
			if r.MaxCPU == nil || *r.MaxCPU != 4 {
				t.Errorf("expected MaxCPU=4")
			}
			if r.MaxMem == nil || *r.MaxMem != 8589934592 {
				t.Errorf("expected MaxMem=8589934592")
			}
			if r.Tags == nil || *r.Tags != "prod;web" {
				t.Errorf("expected tags 'prod;web'")
			}
			if r.Template == nil || *r.Template != 0 {
				t.Errorf("expected template=0")
			}
			return
		}
	}
	t.Error("running VM on node1 not found")
}

func TestGetNodeStatus_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/nodes/{node}/status", jsonHandler(integrationNodeStatusJSON))
	setupIntegrationTest(t, mux)

	resp, err := GetNodeStatus("node1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Data.PveVersion != "pve-manager/8.1.3/bbf3993334bfa916" {
		t.Errorf("unexpected PVE version: %s", resp.Data.PveVersion)
	}
	if resp.Data.Memory.Total != 16384000000 {
		t.Errorf("expected memory total 16384000000, got %d", resp.Data.Memory.Total)
	}
	if resp.Data.Uptime != 100000 {
		t.Errorf("expected uptime 100000, got %d", resp.Data.Uptime)
	}
	if resp.Data.CPUInfo.Cores != 4 {
		t.Errorf("expected 4 cores, got %d", resp.Data.CPUInfo.Cores)
	}
	if resp.Data.CPUInfo.Sockets != 1 {
		t.Errorf("expected 1 socket, got %d", resp.Data.CPUInfo.Sockets)
	}
}

func TestGetNodeDisksList_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/nodes/{node}/disks/list", jsonHandler(integrationNodeDisksJSON))
	setupIntegrationTest(t, mux)

	resp, err := GetNodeDisksList("node1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 disks, got %d", len(resp.Data))
	}

	if resp.Data[0].DevPath != "/dev/nvme0n1" {
		t.Errorf("expected devpath /dev/nvme0n1, got %s", resp.Data[0].DevPath)
	}
	if resp.Data[0].Health != "PASSED" {
		t.Errorf("expected health PASSED, got %s", resp.Data[0].Health)
	}
	if resp.Data[0].Size != 512110190592 {
		t.Errorf("expected size 512110190592, got %d", resp.Data[0].Size)
	}

	if resp.Data[1].DevPath != "/dev/sda" {
		t.Errorf("expected devpath /dev/sda, got %s", resp.Data[1].DevPath)
	}
	if resp.Data[1].Health != "UNKNOWN" {
		t.Errorf("expected health UNKNOWN, got %s", resp.Data[1].Health)
	}
}

func TestGetNodeCertificatesInfo_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/nodes/{node}/certificates/info", jsonHandler(integrationNodeCertsJSON))
	setupIntegrationTest(t, mux)

	resp, err := GetNodeCertificatesInfo("node1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 certs, got %d", len(resp.Data))
	}

	if resp.Data[0].Subject != "CN=pveproxy" {
		t.Errorf("expected subject CN=pveproxy, got %s", resp.Data[0].Subject)
	}
	if resp.Data[0].NotAfter != 1900000000 {
		t.Errorf("expected notafter 1900000000, got %d", resp.Data[0].NotAfter)
	}
	if resp.Data[0].PublicKeyBits != 2048 {
		t.Errorf("expected 2048-bit key, got %d", resp.Data[0].PublicKeyBits)
	}

	if resp.Data[1].Subject != "CN=PVE CA" {
		t.Errorf("expected subject CN=PVE CA, got %s", resp.Data[1].Subject)
	}
	if resp.Data[1].PublicKeyBits != 4096 {
		t.Errorf("expected 4096-bit key, got %d", resp.Data[1].PublicKeyBits)
	}
}

func TestGetQemuSnapshots_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/nodes/{node}/qemu/{vmid}/snapshot", jsonHandler(integrationQemuSnapshotsJSON))
	setupIntegrationTest(t, mux)

	resp, err := GetQemuSnapshots("node1", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 snapshots (2 real + current), got %d", len(resp.Data))
	}

	snapNames := make(map[string]bool)
	for _, s := range resp.Data {
		snapNames[s.Name] = true
	}
	for _, name := range []string{"snap1", "snap2", "current"} {
		if !snapNames[name] {
			t.Errorf("snapshot %q not found", name)
		}
	}

	for _, s := range resp.Data {
		if s.Name == "snap1" && (s.SnapTime == nil || *s.SnapTime != 1700000000) {
			t.Errorf("expected snap1 snaptime=1700000000")
		}
	}
}

func TestGetLxcSnapshots_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/nodes/{node}/lxc/{vmid}/snapshot", jsonHandler(integrationLxcSnapshotsJSON))
	setupIntegrationTest(t, mux)

	resp, err := GetLxcSnapshots("node1", 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 snapshots (1 real + current), got %d", len(resp.Data))
	}

	snapNames := make(map[string]bool)
	for _, s := range resp.Data {
		snapNames[s.Name] = true
	}
	if !snapNames["lxc-snap1"] {
		t.Error("snapshot 'lxc-snap1' not found")
	}
	if !snapNames["current"] {
		t.Error("snapshot 'current' not found")
	}
}

func TestCaching_ClusterStatus(t *testing.T) {
	var counter atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/cluster/status", countingHandler(integrationClusterStatusJSON, &counter))
	setupIntegrationTest(t, mux)

	resp1, err := GetClusterStatus()
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	resp2, err := GetClusterStatus()
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if counter.Load() != 1 {
		t.Errorf("expected 1 HTTP request (second should hit cache), got %d", counter.Load())
	}
	if len(resp1.Data) != len(resp2.Data) {
		t.Error("cached response differs from original")
	}
}

func TestCaching_ClusterResources(t *testing.T) {
	var counter atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/cluster/resources", countingHandler(integrationClusterResourcesJSON, &counter))
	setupIntegrationTest(t, mux)

	_, err := GetClusterResources()
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	_, err = GetClusterResources()
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if counter.Load() != 1 {
		t.Errorf("expected 1 HTTP request (second should hit cache), got %d", counter.Load())
	}
}

func TestCaching_NodeStatus(t *testing.T) {
	var counter atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/nodes/{node}/status", countingHandler(integrationNodeStatusJSON, &counter))
	setupIntegrationTest(t, mux)

	_, err := GetNodeStatus("node1")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	_, err = GetNodeStatus("node1")
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if counter.Load() != 1 {
		t.Errorf("expected 1 HTTP request, got %d", counter.Load())
	}
}

func TestCaching_DifferentNodes(t *testing.T) {
	var counter atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/nodes/{node}/status", countingHandler(integrationNodeStatusJSON, &counter))
	setupIntegrationTest(t, mux)

	_, err := GetNodeStatus("node1")
	if err != nil {
		t.Fatalf("node1 call: %v", err)
	}

	_, err = GetNodeStatus("node2")
	if err != nil {
		t.Fatalf("node2 call: %v", err)
	}

	if counter.Load() != 2 {
		t.Errorf("expected 2 HTTP requests (different cache keys per node), got %d", counter.Load())
	}
}

func TestClientBanning_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/cluster/resources", errorHandler(500))
	setupIntegrationTest(t, mux)

	_, err := GetClusterResources()
	if err == nil {
		t.Fatal("expected error from 500 response")
	}

	if GetBannedClientCount() != 1 {
		t.Errorf("expected 1 banned client, got %d", GetBannedClientCount())
	}
	if GetUnbannedClientCount() != 0 {
		t.Errorf("expected 0 unbanned clients, got %d", GetUnbannedClientCount())
	}
}

func TestAllClientsBanned_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/cluster/status", errorHandler(500))
	setupIntegrationTest(t, mux)

	_, err := GetClusterStatus()
	if err == nil {
		t.Fatal("expected error when all clients fail")
	}

	// After banning, second call should also fail (all clients banned)
	_, err = GetClusterStatus()
	if err == nil {
		t.Fatal("expected error when all clients are banned")
	}
}

func TestClientFailover_Integration(t *testing.T) {
	workingMux := http.NewServeMux()
	workingMux.HandleFunc("/api2/json/cluster/status", jsonHandler(integrationClusterStatusJSON))
	workingServer := httptest.NewServer(workingMux)
	defer workingServer.Close()

	failingMux := http.NewServeMux()
	failingMux.HandleFunc("/api2/json/cluster/status", errorHandler(500))
	failingServer := httptest.NewServer(failingMux)
	defer failingServer.Close()

	workingClient, err := proxmox.NewClient("test-id", "test-token",
		proxmox.WithBaseURL(workingServer.URL+"/"),
		proxmox.WithHTTPClient(&http.Client{}),
	)
	if err != nil {
		t.Fatal(err)
	}

	failingClient, err := proxmox.NewClient("test-id", "test-token",
		proxmox.WithBaseURL(failingServer.URL+"/"),
		proxmox.WithHTTPClient(&http.Client{}),
	)
	if err != nil {
		t.Fatal(err)
	}

	clients = map[string]wrappedClient{
		"working": {client: workingClient},
		"failing": {client: failingClient},
	}
	cash = cache.New(24*time.Second, 5*time.Second)
	t.Cleanup(func() {
		clients = nil
		cash = nil
	})

	resp, err := GetClusterStatus()
	if err != nil {
		t.Fatalf("expected success with failover, got: %v", err)
	}
	if len(resp.Data) != 3 {
		t.Errorf("expected 3 items, got %d", len(resp.Data))
	}

	banned := GetBannedClientCount()
	unbanned := GetUnbannedClientCount()
	if banned+unbanned != 2 {
		t.Errorf("expected 2 total clients, got %d", banned+unbanned)
	}
}

func TestCaching_SnapshotsByVMID(t *testing.T) {
	var counter atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/nodes/{node}/qemu/{vmid}/snapshot", countingHandler(integrationQemuSnapshotsJSON, &counter))
	setupIntegrationTest(t, mux)

	_, err := GetQemuSnapshots("node1", 100)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Same VMID, different node — should use cache (keyed by VMID only)
	_, err = GetQemuSnapshots("node2", 100)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if counter.Load() != 1 {
		t.Errorf("expected 1 HTTP request (cache keyed by VMID), got %d", counter.Load())
	}

	// Different VMID — should miss cache
	_, err = GetQemuSnapshots("node1", 101)
	if err != nil {
		t.Fatalf("third call: %v", err)
	}

	if counter.Load() != 2 {
		t.Errorf("expected 2 HTTP requests (new VMID), got %d", counter.Load())
	}
}

func TestBanDuration_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/node/status", errorHandler(500))
	setupIntegrationTest(t, mux)

	c := clients["mock"]
	banClient("mock", c)

	updated := clients["mock"]
	expectedBanEnd := time.Now().Add(banDuration)

	diff := updated.bannedUntil.Sub(expectedBanEnd)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("bannedUntil is off by %v from expected", diff)
	}
}

func TestRetrieveClusterName_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/cluster/status", jsonHandler(integrationClusterStatusJSON))
	setupIntegrationTest(t, mux)

	oldName := ClusterName
	defer func() { ClusterName = oldName }()

	ClusterName = ""
	retrieveClusterName()

	if ClusterName != "test-cluster" {
		t.Errorf("expected ClusterName='test-cluster', got %q", ClusterName)
	}
}

func TestRetrieveClusterName_EmptyResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/cluster/status", jsonHandler(`{"data": []}`))
	setupIntegrationTest(t, mux)

	oldName := ClusterName
	defer func() { ClusterName = oldName }()

	ClusterName = ""
	retrieveClusterName()

	if ClusterName != "" {
		t.Errorf("expected empty ClusterName, got %q", ClusterName)
	}
}

func TestRetrieveClusterName_NoClusterEntry(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api2/json/cluster/status", jsonHandler(`{
		"data": [
			{"type": "node", "id": "node/node1", "name": "node1", "online": 1}
		]
	}`))
	setupIntegrationTest(t, mux)

	oldName := ClusterName
	defer func() { ClusterName = oldName }()

	ClusterName = ""
	retrieveClusterName()

	if ClusterName != "" {
		t.Errorf("expected empty ClusterName when no cluster entry, got %q", ClusterName)
	}
}
