package proxmox

import "net/http"

// ClusterService is the service that encapsulates node API methods
type ClusterService struct {
	client *Client
}

// GetClusterStatusResponse contains the response for the /cluster/status endpoint
type GetClusterStatusResponse struct {
	Data []GetClusterStatusData `json:"data"`
}

// GetClusterStatusData contains data of a cluster's status from GetClusterStatus
type GetClusterStatusData struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	IP      *string `json:"ip"`
	Level   *string `json:"level"`
	Local   *int    `json:"local"`
	NodeID  *int    `json:"nodeid"`
	Online  *int    `json:"online"`
	Quorate *int    `json:"quorate"`
	Version *int    `json:"version"`
}

// GetClusterStatus makes a GET request to the /cluster/status endpoint
// https://pve.proxmox.com/pve-docs/api-viewer/index.html#/cluster/status
func (s *ClusterService) GetClusterStatus() (*GetClusterStatusResponse, *http.Response, error) {
	u := "cluster/status"
	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	d := new(GetClusterStatusResponse)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}
