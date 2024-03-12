package proxmox

import (
	"github.com/patrickmn/go-cache"
	log "github.com/starttoaster/proxmox-exporter/internal/logger"
	"github.com/starttoaster/proxmox-exporter/pkg/proxmox"
)

// GetClusterStatus returns a proxmox GetClusterStatusResponse object or an error from the /cluster/status endpoint
func GetClusterStatus() (*proxmox.GetClusterStatusResponse, error) {
	// Chech cache
	var cluster *proxmox.GetClusterStatusResponse
	if x, found := cash.Get("GetClusterStatus"); found {
		var ok bool
		cluster, ok = x.(*proxmox.GetClusterStatusResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetClusterStatus")
			return cluster, nil
		}
	}

	// Make request if not found in cache
	var err error
	for _, c := range clients {
		cluster, _, err = c.Cluster.GetClusterStatus()
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Update cache
	cash.Set("GetClusterStatus", cluster, cache.NoExpiration)

	return cluster, nil
}
