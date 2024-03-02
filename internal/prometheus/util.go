package prometheus

import "fmt"

// fqPrefix is the prefix to all metrics exported by this tool
const fqPrefix = "proxmox"

// fqAddPrefix takes a metric name and adds a standard prefix in front of it
func fqAddPrefix(name string) string {
	return fmt.Sprintf("%s_%s", fqPrefix, name)
}
