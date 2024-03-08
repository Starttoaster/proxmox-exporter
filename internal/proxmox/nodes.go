package proxmox

import (
	"fmt"

	"github.com/patrickmn/go-cache"
	log "github.com/starttoaster/proxmox-exporter/internal/logger"
	"github.com/starttoaster/proxmox-exporter/pkg/proxmox"
)

// GetNodes returns a proxmox NodeStatuses object or an error from the /nodes endpoint
func GetNodes() (*proxmox.GetNodesResponse, error) {
	// Chech cache
	var nodes *proxmox.GetNodesResponse
	if x, found := cash.Get("GetNodes"); found {
		var ok bool
		nodes, ok = x.(*proxmox.GetNodesResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for Nodes")
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

// GetNode returns a proxmox Node object or an error from the /nodes/%s/status endpoint
func GetNode(name string) (*proxmox.GetNodeResponse, error) {
	// Chech cache
	var node *proxmox.GetNodeResponse
	if x, found := cash.Get(fmt.Sprintf("GetNode_%s", name)); found {
		var ok bool
		node, ok = x.(*proxmox.GetNodeResponse)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for Node", "node", name)
			return node, nil
		}
	}

	// Make request if not found in cache
	var err error
	for _, c := range clients {
		node, _, err = c.Nodes.GetNode(name)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Update cache
	cash.Set(fmt.Sprintf("GetNode_%s", name), node, cache.DefaultExpiration)

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
