# Proxmox Exporter

A Prometheus exporter for Proxmox clusters.

This exporter does client request dispersion, and response cacheing. If multiple Proxmox cluster API endpoints are provided, they will be selected to receive API requests at random<sup>1</sup>, and we cache the responses for up to 29 seconds, which should produce fresh metrics if scraped in 30 second intervals, or respond with cache at least half of the time if scraped in 15 second intervals. If you run highly available Prometheus instances that each scrape this exporter, it should only need to make the same set of requests to Proxmox's API one time per scrape interval.

This exporter avoids exporting metrics which are redundant to metrics that may be collected by [node_exporter.](https://github.com/prometheus/node_exporter) Ideally, node_exporter should be ran in tandem with this, on your Proxmox nodes as well as in all of your guests. This exporter could be written to export many of the same metrics from node_exporter, but many of those metrics would be misleading. For example a guest may report back to Proxmox a high percentage usage of its allocated memory, but the majority of its memory usage is actually just cache. Other Proxmox exporters do serve these metrics, though (subjectively) they provide no value. If you have a metric that you would like to collect and export to Prometheus through the Proxmox API, but don't see it here, open an Issue and/or a Pull Request, and we can discuss it there.

<sup>1.</sup> Golang maps are abused for randomness, though they should be sufficient for the relatively light purposes of this exporter.

## Metrics

```
# HELP proxmox_cluster_cpus_allocated Total number of vCPU (cores/threads) allocated to guests for a cluster.
# TYPE proxmox_cluster_cpus_allocated gauge
proxmox_cluster_cpus_allocated 24
# HELP proxmox_cluster_cpus_total Total number of vCPU (cores/threads) for a cluster.
# TYPE proxmox_cluster_cpus_total gauge
proxmox_cluster_cpus_total 32
# HELP proxmox_guest_up Shows whether VMs and LXCs in a proxmox cluster are up. (0=down,1=up)
# TYPE proxmox_guest_up gauge
proxmox_guest_up{host="proxmox1",name="CT101",type="lxc",vmid="101"} 0
proxmox_guest_up{host="proxmox1",name="worker1",type="qemu",vmid="102"} 1
proxmox_guest_up{host="proxmox2",name="worker2",type="qemu",vmid="103"} 1
proxmox_guest_up{host="proxmox3",name="worker3",type="qemu",vmid="104"} 1
# HELP proxmox_node_cpus_allocated Total number of vCPU (cores/threads) allocated to guests for a node.
# TYPE proxmox_node_cpus_allocated gauge
proxmox_node_cpus_allocated{name="proxmox1"} 12
proxmox_node_cpus_allocated{name="proxmox2"} 6
proxmox_node_cpus_allocated{name="proxmox3"} 6
# HELP proxmox_node_cpus_total Total number of vCPU (cores/threads) for a node.
# TYPE proxmox_node_cpus_total gauge
proxmox_node_cpus_total{name="proxmox1"} 16
proxmox_node_cpus_total{name="proxmox2"} 8
proxmox_node_cpus_total{name="proxmox3"} 8
# HELP proxmox_node_up Shows whether host nodes in a proxmox cluster are up. (0=down,1=up)
# TYPE proxmox_node_up gauge
proxmox_node_up{name="proxmox1",type="node"} 1
proxmox_node_up{name="proxmox2",type="node"} 1
proxmox_node_up{name="proxmox3",type="node"} 1
```

## How to use

You will need to know some Proxmox API endpoints (`--proxmox-endpoints`), and have a Proxmox API token that's valid to each of those endpoints (should be true in a cluster.) Your Proxmox API token needs at least the PVEAuditor role. When you create an API token (`--proxmox-token`), it comes with a user identifying string (`--proxmox-token-id`) which is also needed. Lastly, if your API server's TLS cannot be verified, you will need to set `--proxmox-api-insecure=true`.

You can pass your configuration with the following CLI flags.

```bash
Usage:
  proxmox-exporter [flags]

Flags:
  -h, --help                       help for proxmox-exporter
      --log-level string           The log-level for the application, can be one of info, warn, error, debug. (default "info")
      --proxmox-api-insecure       Whether or not this client should accept insecure connections to Proxmox (default: false)
      --proxmox-endpoints string   The Proxmox API endpoint, you can pass in multiple endpoints separated by commas (ex: https://localhost:8006/)
      --proxmox-token string       Proxmox API token
      --proxmox-token-id string    Proxmox API token ID
      --server-port uint16         The port the metrics server binds to. (default 8080)
```

Or you can set the corresponding environment variables.

```bash
PROXMOX_EXPORTER_LOG_LEVEL="info"
PROXMOX_EXPORTER_PROXMOX_API_INSECURE=false
PROXMOX_EXPORTER_PROXMOX_ENDPOINTS="https://x:8006/,https://y:8006/,https://z:8006/"
PROXMOX_EXPORTER_PROXMOX_TOKEN="redacted-token"
PROXMOX_EXPORTER_PROXMOX_TOKEN_ID="redacted-token-id"
PROXMOX_EXPORTER_SERVER_PORT=8080
```

## Deploy

Documentation and deployment details in progress.

## TODO

- Add metrics to export
- Loop to retry API requests using a different client (handle proxmox node shutdowns for cluster maintenances gracefully)
- Add helm chart
- Finish documentation

## Planned metrics
Cluster memory used (gauge)
Cluster memory total (gauge)

Node mem allocated (gauge)
Node mem total (gauge)

Node certificate expiry countdown days (gauge) (labels for certificate name, and node name)

Storage usage per volume per node (gauge) (labels for storage type, storage name, and what node it's on)
Storage capacity per volume per node (gauge) (labels for storage type, storage name, and what node it's on)

Smart status by disk in the cluster