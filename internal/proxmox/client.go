package proxmox

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/starttoaster/proxmox-exporter/internal/logger"

	proxmox "github.com/starttoaster/go-proxmox"
)

var (
	// ClusterName gets populated with the proxmox cluster's cluster name on clustered PVE instances
	ClusterName string

	clients     map[string]wrappedClient
	banDuration = time.Duration(1 * time.Minute)
	cash        *cache.Cache
)

type wrappedClient struct {
	client      *proxmox.Client
	banned      bool
	bannedUntil time.Time
}

// Init constructs a proxmox API client for this package taking in a token
func Init(endpoints []string, tokenID, token string, tlsVerify bool) error {
	// Fail early if endpoints slice is 0 length
	if len(endpoints) == 0 {
		return fmt.Errorf("no Proxmox API endpoints supplied")
	}

	// Define http client, for optional insecure API endpoints
	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: tlsVerify,
			},
		},
	}

	// Make and init proxmox client map
	clients = make(map[string]wrappedClient)
	for _, endpoint := range endpoints {
		// Parse URL for hostname
		parsedURL, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("error parsing URL: \"%s\"", err)
		}
		hostname := parsedURL.Hostname()

		// Create API client
		log.Logger.Debug("Creating Proxmox client", "endpoint", endpoint, "hostname", hostname)
		c, err := proxmox.NewClient(tokenID, token,
			proxmox.WithBaseURL(endpoint),
			proxmox.WithHTTPClient(&httpClient),
		)
		if err != nil {
			return fmt.Errorf("error creating API client for exporter: %w", err)
		}

		// Add client to map
		clients[hostname] = wrappedClient{
			client: c,
		}
	}

	// init cache -- at longest, cache will live for 29 seconds
	// which should ensure metrics are updated if scraping in 30 second intervals
	// TODO should cache lifespan and cache expiration intervals be user configurable?
	cash = cache.New(24*time.Second, 5*time.Second)

	// Maintain client bans
	go refreshClientBans()

	retrieveClusterName()

	return nil
}

func retrieveClusterName() {
	// Retrieve cluster status -- if clustered
	clusterStatus, err := GetClusterStatus()
	if err != nil {
		return
	}

	// Exit if no data returned
	if len(clusterStatus.Data) == 0 {
		return
	}

	// Parse out cluster name
	for _, cluster := range clusterStatus.Data {
		if strings.EqualFold(cluster.Type, "cluster") {
			ClusterName = cluster.Name
			break
		}
	}
	if ClusterName != "" {
		log.Logger.Info("discovered PVE cluster", "cluster", ClusterName)
	}
}

// refreshClientBans iterates over the configured clients and checks if their ban is still valid over time
func refreshClientBans() {
	for {
		// Loop through clients from client map
		for name, c := range clients {
			// Check if the client is banned -- if banned we need to check if the client's banUntil time has expired
			if c.banned && time.Now().After(c.bannedUntil) {
				// If the ban expired, make a request, see if it succeeds, and unban it if successful. Increase the ban timer if not
				_, _, err := c.client.Nodes.GetNodes()
				if err == nil {
					// Unban client - request successful
					log.Logger.Debug("unbanning client, test request successful", "name", name)
					clients[name] = wrappedClient{
						client: c.client,
					}
					continue
				} else {
					// Re-up ban timer - request failed
					log.Logger.Debug("re-upping ban on client, test request failed", "name", name, "error", err)
					banClient(name, c)
					continue
				}
			}
		}

		time.Sleep(5 * time.Second)
	}
}

// banClient bans a client for the defined duration
func banClient(name string, c wrappedClient) {
	log.Logger.Debug("banning client", "name", name, "duration", banDuration)
	clients[name] = wrappedClient{
		client:      c.client,
		banned:      true,
		bannedUntil: time.Now().Add(banDuration),
	}
}
