package proxmox

import (
	"fmt"

	"github.com/patrickmn/go-cache"
	proxmox "github.com/starttoaster/go-proxmox"
	log "github.com/starttoaster/proxmox-exporter/internal/logger"
)

// GetNodes returns a proxmox NodeStatuses object or an error from the /nodes endpoint
func GetNodes() (*proxmox.GetNodesResponse, error) {
	// Chech cache
	var nodes *proxmox.GetNodesResponse
	if x, found := cash.Get("GetNodes"); found {
		var ok bool
		nodes, ok = x.(*proxmox.GetNodesResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetNodes")
			return nodes, nil
		}
	}

	// Make request if not found in cache
	var err error
	for _, c := range clients {
		nodes, _, err = c.Nodes.GetNodes()
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Update cache
	cash.Set("GetNodes", nodes, cache.DefaultExpiration)

	return nodes, nil
}

// GetNodeStatus returns a proxmox Node object or an error from the /nodes/%s/status endpoint
func GetNodeStatus(name string) (*proxmox.GetNodeStatusResponse, error) {
	// Chech cache
	var node *proxmox.GetNodeStatusResponse
	if x, found := cash.Get(fmt.Sprintf("GetNodeStatus_%s", name)); found {
		var ok bool
		node, ok = x.(*proxmox.GetNodeStatusResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetNodeStatus", "node", name)
			return node, nil
		}
	}

	// Make request if not found in cache
	var err error
	for _, c := range clients {
		node, _, err = c.Nodes.GetNodeStatus(name)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Update cache
	cash.Set(fmt.Sprintf("GetNodeStatus_%s", name), node, cache.DefaultExpiration)

	return node, nil
}

// GetNodeQemu returns the virtual machines for a node
func GetNodeQemu(name string) (*proxmox.GetNodeQemuResponse, error) {
	// Chech cache
	var vms *proxmox.GetNodeQemuResponse
	if x, found := cash.Get(fmt.Sprintf("GetNodeQemu_%s", name)); found {
		var ok bool
		vms, ok = x.(*proxmox.GetNodeQemuResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetNodeQemu", "node", name)
			return vms, nil
		}
	}

	// Make request if not found in cache
	var err error
	for _, c := range clients {
		vms, _, err = c.Nodes.GetNodeQemu(name)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Update per-node cache since we have it
	cash.Set(fmt.Sprintf("GetNodeQemu_%s", name), vms, cache.DefaultExpiration)

	return vms, nil
}

// GetNodeLxc returns the LXCs for a node
func GetNodeLxc(name string) (*proxmox.GetNodeLxcResponse, error) {
	// Chech cache
	var lxcs *proxmox.GetNodeLxcResponse
	if x, found := cash.Get(fmt.Sprintf("GetNodeLxc_%s", name)); found {
		var ok bool
		lxcs, ok = x.(*proxmox.GetNodeLxcResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetNodeLxc", "node", name)
			return lxcs, nil
		}
	}

	// Make request if not found in cache
	var err error
	for _, c := range clients {
		lxcs, _, err = c.Nodes.GetNodeLxc(name)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Update per-node cache since we have it
	cash.Set(fmt.Sprintf("GetNodeLxc_%s", name), lxcs, cache.DefaultExpiration)

	return lxcs, nil
}

// GetNodeDisksList returns the disks for a node
func GetNodeDisksList(name string) (*proxmox.GetNodeDisksListResponse, error) {
	// Chech cache
	var disks *proxmox.GetNodeDisksListResponse
	if x, found := cash.Get(fmt.Sprintf("GetNodeDisksList_%s", name)); found {
		var ok bool
		disks, ok = x.(*proxmox.GetNodeDisksListResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetNodeDisksList", "node", name)
			return disks, nil
		}
	}

	// Make request if not found in cache
	var err error
	for _, c := range clients {
		disks, _, err = c.Nodes.GetNodeDisksList(name)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Update per-node cache since we have it
	cash.Set(fmt.Sprintf("GetNodeDisksList_%s", name), disks, cache.DefaultExpiration)

	return disks, nil
}

// GetNodeCertificatesInfo returns the certificates for a node
func GetNodeCertificatesInfo(name string) (*proxmox.GetNodeCertificatesInfoResponse, error) {
	// Chech cache
	var certs *proxmox.GetNodeCertificatesInfoResponse
	if x, found := cash.Get(fmt.Sprintf("GetNodeCertificatesInfo_%s", name)); found {
		var ok bool
		certs, ok = x.(*proxmox.GetNodeCertificatesInfoResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetNodeCertificatesInfo", "node", name)
			return certs, nil
		}
	}

	// Make request if not found in cache
	var err error
	for _, c := range clients {
		certs, _, err = c.Nodes.GetNodeCertificatesInfo(name)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Update per-node cache since we have it
	cash.Set(fmt.Sprintf("GetNodeCertificatesInfo_%s", name), certs, cache.DefaultExpiration)

	return certs, nil
}

// GetNodeStorage returns the storage for a node
func GetNodeStorage(name string) (*proxmox.GetNodeStorageResponse, error) {
	// Chech cache
	var store *proxmox.GetNodeStorageResponse
	if x, found := cash.Get(fmt.Sprintf("GetNodeStorage_%s", name)); found {
		var ok bool
		store, ok = x.(*proxmox.GetNodeStorageResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetNodeStorage", "node", name)
			return store, nil
		}
	}

	// Make request if not found in cache
	var err error
	for _, c := range clients {
		store, _, err = c.Nodes.GetNodeStorage(name)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Update per-node cache since we have it
	cash.Set(fmt.Sprintf("GetNodeStorage_%s", name), store, cache.DefaultExpiration)

	return store, nil
}
