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
	for name, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		nodes, _, err = c.client.Nodes.GetNodes()
		if err == nil {
			break
		} else {
			banClient(name, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if nodes == nil {
		return nil, fmt.Errorf("request to get node list was not successful. It's possible all clients are banned")
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
	for clientName, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		node, _, err = c.client.Nodes.GetNodeStatus(name)
		if err == nil {
			break
		} else {
			banClient(clientName, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, fmt.Errorf("request to get node status was not successful. It's possible all clients are banned")
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
	for clientName, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		vms, _, err = c.client.Nodes.GetNodeQemu(name)
		if err == nil {
			break
		} else {
			banClient(clientName, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if vms == nil {
		return nil, fmt.Errorf("request to get node VMs was not successful. It's possible all clients are banned")
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
	for clientName, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		lxcs, _, err = c.client.Nodes.GetNodeLxc(name)
		if err == nil {
			break
		} else {
			banClient(clientName, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if lxcs == nil {
		return nil, fmt.Errorf("request to get node LXCs was not successful. It's possible all clients are banned")
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
	for clientName, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		disks, _, err = c.client.Nodes.GetNodeDisksList(name)
		if err == nil {
			break
		} else {
			banClient(clientName, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if disks == nil {
		return nil, fmt.Errorf("request to get node disks was not successful. It's possible all clients are banned")
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
	for clientName, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		certs, _, err = c.client.Nodes.GetNodeCertificatesInfo(name)
		if err == nil {
			break
		} else {
			banClient(clientName, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if certs == nil {
		return nil, fmt.Errorf("request to get node certificates was not successful. It's possible all clients are banned")
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
	for clientName, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		store, _, err = c.client.Nodes.GetNodeStorage(name)
		if err == nil {
			break
		} else {
			banClient(clientName, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if store == nil {
		return nil, fmt.Errorf("request to get node storage was not successful. It's possible all clients are banned")
	}

	// Update per-node cache since we have it
	cash.Set(fmt.Sprintf("GetNodeStorage_%s", name), store, cache.DefaultExpiration)

	return store, nil
}

// GetQemuSnapshots returns the snapshots for a VM
func GetQemuSnapshots(nodeName string, vmID int) (*proxmox.GetQemuSnapshotsResponse, error) {
	// Only using VM ID for the cache key because a VM/LXC can be migrated between cluster nodes in some storage configurations (like Ceph)
	cacheKey := fmt.Sprintf("GetQemuSnapshots_%d", vmID)

	// Chech cache
	var out *proxmox.GetQemuSnapshotsResponse
	if x, found := cash.Get(cacheKey); found {
		var ok bool
		out, ok = x.(*proxmox.GetQemuSnapshotsResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetQemuSnapshots", "node", nodeName, "vm_id", vmID)
			return out, nil
		}
	}

	// Make request if not found in cache
	var err error
	for clientName, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		out, _, err = c.client.Nodes.GetQemuSnapshots(nodeName, vmID)
		if err == nil {
			break
		} else {
			banClient(clientName, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, fmt.Errorf("request to get qemu snapshots was not successful. It's possible all clients are banned")
	}

	// Update per-node cache since we have it
	cash.Set(cacheKey, out, cache.DefaultExpiration)

	return out, nil
}

// GetLxcSnapshots returns the snapshots for a LXC
func GetLxcSnapshots(nodeName string, vmID int) (*proxmox.GetLxcSnapshotsResponse, error) {
	// Only using VM ID for the cache key because a VM/LXC can be migrated between cluster nodes in some storage configurations (like Ceph)
	cacheKey := fmt.Sprintf("GetLxcSnapshots_%d", vmID)

	// Chech cache
	var out *proxmox.GetLxcSnapshotsResponse
	if x, found := cash.Get(cacheKey); found {
		var ok bool
		out, ok = x.(*proxmox.GetLxcSnapshotsResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for GetLxcSnapshots", "node", nodeName, "lxc_id", vmID)
			return out, nil
		}
	}

	// Make request if not found in cache
	var err error
	for clientName, c := range clients {
		// Check if client was banned, skip if is
		if c.banned {
			continue
		}

		out, _, err = c.client.Nodes.GetLxcSnapshots(nodeName, vmID)
		if err == nil {
			break
		} else {
			banClient(clientName, c)
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, fmt.Errorf("request to get LXC snapshots was not successful. It's possible all clients are banned")
	}

	// Update per-node cache since we have it
	cash.Set(cacheKey, out, cache.DefaultExpiration)

	return out, nil
}
