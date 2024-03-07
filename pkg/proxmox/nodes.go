package proxmox

import (
	"net/http"
)

// NodeService is the service that encapsulates node API methods
type NodeService struct {
	client *Client
}

// Nodes is a list of Node types
type Nodes []Node

// Node contains data attributes for a node in the Proxmox nodes API
type Node struct {
	CPU            float64 `json:"cpu"`
	Disk           int64   `json:"disk"`
	ID             string  `json:"id"`
	Level          string  `json:"level"`
	Maxcpu         int     `json:"maxcpu"`
	Maxdisk        int64   `json:"maxdisk"`
	Maxmem         int64   `json:"maxmem"`
	Mem            int64   `json:"mem"`
	Node           string  `json:"node"`
	SslFingerprint string  `json:"ssl_fingerprint"`
	Status         string  `json:"status"`
	Type           string  `json:"type"`
	Uptime         int     `json:"uptime"`
}

// Get makes a GET request to the /nodes endpoint
func (s *NodeService) Get() (*Nodes, *http.Response, error) {
	u := "nodes"
	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	d := new(Nodes)
	resp, err := s.client.Do(req, &d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}
