package proxmox

import (
	"fmt"
	"net/http"
)

// NodeService is the service that encapsulates node API methods
type NodeService struct {
	client *Client
}

// GetNodesResponse contains the response for the /nodes endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes
type GetNodesResponse struct {
	Data []GetNodesData `json:"data"`
}

// GetNodesData contains data of one node from a GetNodes response
type GetNodesData struct {
	CPU            float64 `json:"cpu"`
	Disk           int64   `json:"disk"`
	ID             string  `json:"id"`
	Level          string  `json:"level"`
	MaxCPU         int     `json:"maxcpu"`
	MaxDisk        int64   `json:"maxdisk"`
	MaxMem         int64   `json:"maxmem"`
	Mem            int64   `json:"mem"`
	Node           string  `json:"node"`
	SslFingerprint string  `json:"ssl_fingerprint"`
	Status         string  `json:"status"`
	Type           string  `json:"type"`
	Uptime         int     `json:"uptime"`
}

// GetNodes makes a GET request to the /nodes endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes
func (s *NodeService) GetNodes() (*GetNodesResponse, *http.Response, error) {
	u := "nodes"
	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	d := new(GetNodesResponse)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// GetNodeResponse contains the response for the /nodes/{node}/status endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/status
type GetNodeResponse struct {
	Data GetNodeData `json:"data"`
}

// GetNodeResponse contains data of one node from a GetNode response
type GetNodeData struct {
	BootInfo      BootInfo      `json:"boot-info"`
	CPU           float64       `json:"cpu"`
	CPUInfo       CPUInfo       `json:"cpuinfo"`
	CurrentKernel CurrentKernel `json:"current-kernel"`
	Idle          int           `json:"idle"`
	Ksm           Ksm           `json:"ksm"`
	Kversion      string        `json:"kversion"`
	LoadAvg       []string      `json:"loadavg"`
	Memory        Memory        `json:"memory"`
	PveVersion    string        `json:"pveversion"`
	RootFs        RootFs        `json:"rootfs"`
	Swap          Swap          `json:"swap"`
	Uptime        int           `json:"uptime"`
	Wait          float64       `json:"wait"`
}

// GetNode makes a GET request to the /nodes/{node}/status endpoint
// This returns more information about a node than the /nodes endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/status
func (s *NodeService) GetNode(name string) (*GetNodeResponse, *http.Response, error) {
	u := fmt.Sprintf("nodes/%s/status", name)
	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	d := new(GetNodeResponse)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// GetNodeVersionResponse contains the response for the /nodes/{node}/version endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/version
type GetNodeVersionResponse struct {
	Release string `json:"release"`
	RepoID  string `json:"repoid"`
	Version string `json:"version"`
}

// GetNodeVersion makes a GET request to the /nodes/{node}/version endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/version
func (s *NodeService) GetNodeVersion(name string) (*GetNodeVersionResponse, *http.Response, error) {
	u := fmt.Sprintf("nodes/%s/version", name)
	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	d := new(GetNodeVersionResponse)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// GetNodeQemuResponse contains the response for the /nodes/{node}/qemu endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes
type GetNodeQemuResponse struct {
	Data []GetNodeQemuData `json:"data"`
}

// GetNodeQemuData contains data of one VM from a GetNodeQemu response
type GetNodeQemuData struct {
	CPU       float64 `json:"cpu"`
	Cpus      int     `json:"cpus"`
	Disk      int     `json:"disk"`
	DiskRead  int     `json:"diskread"`
	DiskWrite int     `json:"diskwrite"`
	MaxDisk   int64   `json:"maxdisk"`
	MaxMem    int64   `json:"maxmem"`
	Mem       int64   `json:"mem"`
	Name      string  `json:"name"`
	NetIn     int64   `json:"netin"`
	NetOut    int64   `json:"netout"`
	Pid       int     `json:"pid"`
	Status    string  `json:"status"`
	Uptime    int     `json:"uptime"`
	VMID      int     `json:"vmid"`
}

// GetNodeQemu makes a GET request to the /nodes/{node}/qemu endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/qemu
func (s *NodeService) GetNodeQemu(name string) (*GetNodeQemuResponse, *http.Response, error) {
	u := fmt.Sprintf("nodes/%s/qemu", name)
	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	d := new(GetNodeQemuResponse)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// GetNodeLxcResponse contains the response for the /nodes/{node}/lxc endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/lxc
type GetNodeLxcResponse struct {
	Data []GetNodeLxcData `json:"data"`
}

// GetNodeLxcData contains data of one VM from a GetNodeLxc response
type GetNodeLxcData struct {
	CPU       float64 `json:"cpu"`
	Cpus      int     `json:"cpus"`
	Disk      int     `json:"disk"`
	DiskRead  int     `json:"diskread"`
	DiskWrite int     `json:"diskwrite"`
	MaxDisk   int64   `json:"maxdisk"`
	MaxMem    int64   `json:"maxmem"`
	MaxSwap   int64   `json:"maxswap"`
	Mem       int64   `json:"mem"`
	Name      string  `json:"name"`
	NetIn     int64   `json:"netin"`
	NetOut    int64   `json:"netout"`
	Status    string  `json:"status"`
	Type      string  `json:"type"`
	Uptime    int     `json:"uptime"`
	VMID      string  `json:"vmid"`
}

// GetNodeLxc makes a GET request to the /nodes/{node}/lxc endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/lxc
func (s *NodeService) GetNodeLxc(name string) (*GetNodeLxcResponse, *http.Response, error) {
	u := fmt.Sprintf("nodes/%s/lxc", name)
	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	d := new(GetNodeLxcResponse)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}
