package prometheus

import "fmt"

// fqPrefix is the prefix to all metrics exported by this tool
const fqPrefix = "proxmox"

// fqWithPrefix takes a metric name and adds a standard prefix in front of it
func fqWithPrefix(name string) string {
	return fmt.Sprintf("%s_%s", fqPrefix, name)
}
