package proxmox

// GetBannedClientCount returns the number of banned clients
func GetBannedClientCount() int {
	var count int
	for _, c := range clients {
		if c.banned {
			count++
		}
	}
	return count
}

// GetUnbannedClientCount returns the number of unbanned clients
func GetUnbannedClientCount() int {
	var count int
	for _, c := range clients {
		if !c.banned {
			count++
		}
	}
	return count
}
