package proxmox

import (
	"testing"
	"time"

	log "github.com/starttoaster/proxmox-exporter/internal/logger"
)

func init() {
	log.Init("error")
}

func TestInit_NoEndpoints(t *testing.T) {
	err := Init([]string{}, "tokenid", "token", false)
	if err == nil {
		t.Fatal("expected error when no endpoints supplied")
	}
}

func TestInit_EmptySlice(t *testing.T) {
	err := Init(nil, "tokenid", "token", false)
	if err == nil {
		t.Fatal("expected error when nil endpoints supplied")
	}
}

func TestBanClient(t *testing.T) {
	clients = map[string]wrappedClient{
		"host1": {banned: false},
	}

	c := clients["host1"]
	banClient("host1", c)

	updated := clients["host1"]
	if !updated.banned {
		t.Error("client should be banned after banClient call")
	}
	if updated.bannedUntil.Before(time.Now()) {
		t.Error("bannedUntil should be in the future")
	}
	if updated.bannedUntil.After(time.Now().Add(2 * banDuration)) {
		t.Error("bannedUntil should not be too far in the future")
	}
}

func TestBanClient_MultipleBans(t *testing.T) {
	clients = map[string]wrappedClient{
		"host1": {banned: false},
		"host2": {banned: false},
	}

	banClient("host1", clients["host1"])
	if !clients["host1"].banned {
		t.Error("host1 should be banned")
	}
	if clients["host2"].banned {
		t.Error("host2 should not be banned")
	}

	banClient("host2", clients["host2"])
	if !clients["host1"].banned {
		t.Error("host1 should still be banned")
	}
	if !clients["host2"].banned {
		t.Error("host2 should now be banned")
	}
}

func TestBanClient_PreservesClient(t *testing.T) {
	clients = map[string]wrappedClient{
		"host1": {banned: false},
	}

	c := clients["host1"]
	origClient := c.client

	banClient("host1", c)

	updated := clients["host1"]
	if updated.client != origClient {
		t.Error("banClient should preserve the original proxmox client reference")
	}
}

func TestBanDuration(t *testing.T) {
	if banDuration != 1*time.Minute {
		t.Errorf("expected ban duration of 1 minute, got %v", banDuration)
	}
}

func TestWrappedClient_DefaultValues(t *testing.T) {
	c := wrappedClient{}
	if c.banned {
		t.Error("default wrapped client should not be banned")
	}
	if !c.bannedUntil.IsZero() {
		t.Error("default wrapped client bannedUntil should be zero")
	}
	if c.client != nil {
		t.Error("default wrapped client should have nil client")
	}
}
