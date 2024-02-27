package proxmox

import (
	"context"
	"fmt"

	"github.com/luthermonson/go-proxmox"
	"github.com/patrickmn/go-cache"
	log "github.com/starttoaster/proxmox-exporter/internal/logger"
)

// Nodes returns a proxmox NodeStatuses object or an error from the /nodes endpoint
func Nodes() (proxmox.NodeStatuses, error) {
	// Chech cache
	var nodes proxmox.NodeStatuses
	if x, found := cash.Get("Nodes"); found {
		var ok bool
		nodes, ok = x.(proxmox.NodeStatuses)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for Nodes")
			return nodes, nil
		}
	}

	// Make request if not found in cache
	nodes, err := anyClient().Nodes(context.Background())
	if err != nil {
		return nil, fmt.Errorf("encountered error making request to /nodes: \n%v", err)
	}

	// Update cache
	cash.Set("Nodes", nodes, cache.DefaultExpiration)

	return nodes, nil
}

// Node returns a proxmox Node object or an error from the /nodes/%s/status endpoint
func Node(name string) (*proxmox.Node, error) {
	// Chech cache
	var node *proxmox.Node
	if x, found := cash.Get(fmt.Sprintf("Node_%s", name)); found {
		var ok bool
		node, ok = x.(*proxmox.Node)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for Node", "node", name)
			return node, nil
		}
	}

	// Make request if not found in cache
	node, err := anyClient().Node(context.Background(), name)
	if err != nil {
		return nil, fmt.Errorf("encountered error making request to /nodes/%s/status: \n%v", name, err)
	}

	// Update cache
	cash.Set(fmt.Sprintf("Node_%s", name), node, cache.DefaultExpiration)

	return node, nil
}
