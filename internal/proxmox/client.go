package proxmox

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/starttoaster/proxmox-exporter/internal/logger"

	"github.com/luthermonson/go-proxmox"
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

		// Add client to map
		log.Logger.Debug("Creating Proxmox client", "endpoint", endpoint, "hostname", hostname)
		clients[hostname] = proxmox.NewClient(endpoint,
			proxmox.WithHTTPClient(&httpClient),
			proxmox.WithAPIToken(tokenID, token),
		)
	}

	// Test API client with a request to the server
	log.Logger.Debug("Testing Proxmox API client with request to server version")
	version, err := anyClient().Version(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving Proxmox server version: \n%v", err)
	}

	// Check and log version
	if version == nil {
		return errors.New("the Proxmox server returned no error from version request, but version is nil")
	}
	log.Logger.Debug("Proxmox server", "version", version.Release)

	// init cache -- at longest, cache will live for 29 seconds
	// which should ensure metrics are updated if scraping in 30 second intervals
	// TODO should cache lifespan and cache expiration intervals be user configurable?
	cash = cache.New(24*time.Second, 5*time.Second)

	return nil
}

// anyClient ranges over the list of Proxmox clients available and returns the first found.
// In Proxmox clusters, this gets us pseudo-random clients used for each request.
// This leaves nilness checking to the consumer of this function,
// but it may be presumable that an actual client is returned here
// because we check for zero length clients in the init function
func anyClient() *proxmox.Client {
	for _, v := range clients {
		return v
	}
	return nil
}

/*
Unclear yet if this will end up being useful, maybe remove if we start feeling like the exporter is nearly complete and this still isn't used

// theClient accepts the string key for a specific client, useful if the metric comes from a specific host in a cluster
// Returns an error if the named client isn't found
func theClient(k string) (*proxmox.Client, error) {
	c := clients[k]
	if c == nil {
		return nil, fmt.Errorf("client specified by key \"%s\" not found", k)
	}
	return c, nil
}
*/
