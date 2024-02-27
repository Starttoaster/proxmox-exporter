package proxmox

import (
	"context"
	"fmt"

	"github.com/luthermonson/go-proxmox"
	"github.com/patrickmn/go-cache"
	log "github.com/starttoaster/proxmox-exporter/internal/logger"
)

// VirtualMachinesAllNodes loops through all found nodes by the /nodes endpoint, gets the node details for each from /node,
// and returns the virtual machines for all nodes
func VirtualMachinesAllNodes() (proxmox.VirtualMachines, error) {
	// Chech cache
	var vms proxmox.VirtualMachines
	if x, found := cash.Get("VirtualMachinesAllNodes"); found {
		var ok bool
		vms, ok = x.(proxmox.VirtualMachines)
		if ok {
			log.Logger.Debug("proxmox request was found in cache for VirtualMachinesAllNodes")
			return vms, nil
		}
	}

	// Get all nodes' statuses, needed because the request to /nodes/%s/qemu is per-node
	// And to do that we'll need the names of all of our nodes
	nodeStatuses, err := Nodes()
	if err != nil {
		return nil, err
	}

	// Get node object now that we know the names and statuses of each node
	for _, nodeStatus := range nodeStatuses {
		// Get node object from status
		node, err := Node(nodeStatus.Name)
		if err != nil {
			return nil, err
		}

		// Get VMs on the node for this iteration
		nodeVms, err := node.VirtualMachines(context.Background())
		if err != nil {
			return nil, fmt.Errorf("encountered error making request to /nodes/%s/qemu: \n%v", node.Name, err)
		}

		// Update per-node cache since we have it
		cash.Set(fmt.Sprintf("VirtualMachinesOnNode_%s", nodeStatus.Name), nodeVms, cache.DefaultExpiration)

		// Append to overall vms list
		vms = append(vms, nodeVms...)
	}

	// Update cache
	cash.Set("VirtualMachinesAllNodes", vms, cache.DefaultExpiration)

	return vms, nil
}
