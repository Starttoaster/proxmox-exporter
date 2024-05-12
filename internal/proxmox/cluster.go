package proxmox

import (
	"fmt"

	"github.com/patrickmn/go-cache"
	proxmox "github.com/starttoaster/go-proxmox"
	log "github.com/starttoaster/proxmox-exporter/internal/logger"
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
	for clientName, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		cluster, _, err = c.client.Cluster.GetClusterStatus()
		if err == nil {
			break
		} else {
			banClient(clientName, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if cluster == nil {
		return nil, fmt.Errorf("request to get cluster status was not successful. It's possible all clients are banned")
	}

	// Update cache
	cash.Set("GetClusterStatus", cluster, cache.NoExpiration)

	return cluster, nil
}

// GetClusterResources returns a proxmox GetClusterResourcesResponse object or an error from the /cluster/resources endpoint
func GetClusterResources() (*proxmox.GetClusterResourcesResponse, error) {
	// Chech cache
	var resources *proxmox.GetClusterResourcesResponse
	if x, found := cash.Get("GetClusterResources"); found {
		var ok bool
		resources, ok = x.(*proxmox.GetClusterResourcesResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetClusterResources")
			return resources, nil
		}
	}

	// Make request if not found in cache
	var err error
	for clientName, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		resources, _, err = c.client.Cluster.GetClusterResources()
		if err == nil {
			break
		} else {
			banClient(clientName, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if resources == nil {
		return nil, fmt.Errorf("request to get cluster resources was not successful. It's possible all clients are banned")
	}

	// Update cache
	cash.Set("GetClusterResources", resources, cache.NoExpiration)

	return resources, nil
}
