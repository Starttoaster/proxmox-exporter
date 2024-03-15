package prometheus

import (
	"fmt"
	"time"
)

// fqPrefix is the prefix to all metrics exported by this tool
const fqPrefix = "proxmox"

// fqAddPrefix takes a metric name and adds a standard prefix in front of it
func fqAddPrefix(name string) string {
	return fmt.Sprintf("%s_%s", fqPrefix, name)
}

// daysUntilUnixTime takes a unix timestamp in an int and returns the integer number of days until the given date
func daysUntilUnixTime(notAfter int) int {
	currentTime := time.Now().Unix()
	differenceSeconds := int64(notAfter) - currentTime
	differenceDays := differenceSeconds / (60 * 60 * 24)
	return int(differenceDays)
}
