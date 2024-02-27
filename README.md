# Proxmox Exporter

A Prometheus exporter for Proxmox.

This exports metrics for your Proxmox server/cluster, while avoiding exporting data that's redundant to metrics that may be collected by [node_exporter.](https://github.com/prometheus/node_exporter) Many of those metrics are misleading, for example a guest may report a high percentage usage of allocated memory, but the majority of its memory usage is actually just cache.

This exporter does client request dispersion, and response cacheing. If multiple Proxmox cluster API endpoints are provided, they will be selected to make requests to at random<sup>1</sup>, and we cache the responses for up to 29 seconds, which should produce fresh metrics if scraped in 30 second intervals, or often respond with cache if scraped in 15 second intervals.

<sup>1.</sup> Golang maps are abused for randomness, though they should be sufficient for the relatively light purposes of this exporter.

## Metrics

Documentation and deployment details in progress.

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
      --proxmox-endpoints string   The Proxmox API endpoint, you can pass in multiple endpoints separated by commas (ex: https://localhost:8006/api2/json)
      --proxmox-token string       Proxmox API token
      --proxmox-token-id string    Proxmox API token ID
      --server-port uint16         The port the metrics server binds to. (default 8080)
```

Or you can set the corresponding environment variables.

```bash
PROXMOX_EXPORTER_LOG_LEVEL="info"
PROXMOX_EXPORTER_PROXMOX_API_INSECURE=false
PROXMOX_EXPORTER_PROXMOX_ENDPOINTS="https://x:8006/api2/json,https://y:8006/api2/json,https://z:8006/api2/json"
PROXMOX_EXPORTER_PROXMOX_TOKEN="redacted-token"
PROXMOX_EXPORTER_PROXMOX_TOKEN_ID="redacted-token-id"
PROXMOX_EXPORTER_SERVER_PORT=8080
```

## Deploy

Documentation and deployment details in progress.

## TODO
- Additional labels option to add to metrics
- Add metrics to export
- Maybe all API requests should actually be in a loop that retries for each client available, in case of cluster maintenances
- Add helm chart
- Finish documentation
