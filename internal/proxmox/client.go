package proxmox

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/starttoaster/proxmox-exporter/internal/logger"

	"github.com/starttoaster/proxmox-exporter/pkg/proxmox"
)

var (
	clients map[string]*proxmox.Client
	cash    *cache.Cache
)

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
	clients = make(map[string]*proxmox.Client)
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
		clients[hostname] = c
	}

	// init cache -- at longest, cache will live for 29 seconds
	// which should ensure metrics are updated if scraping in 30 second intervals
	// TODO should cache lifespan and cache expiration intervals be user configurable?
	cash = cache.New(24*time.Second, 5*time.Second)

	return nil
}
