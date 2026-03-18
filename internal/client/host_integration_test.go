package client

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestClientHostLifecycleIntegration(t *testing.T) {
	baseURL := os.Getenv("GESTIOIP_BASE_URL")
	username := os.Getenv("GESTIOIP_USERNAME")
	password := os.Getenv("GESTIOIP_PASSWORD")
	clientName := os.Getenv("GESTIOIP_CLIENT_NAME")

	if baseURL == "" || username == "" || password == "" || clientName == "" {
		t.Skip("set GESTIOIP_BASE_URL, GESTIOIP_USERNAME, GESTIOIP_PASSWORD and GESTIOIP_CLIENT_NAME to run integration tests")
	}

	apiClient, err := New(Config{
		BaseURL:    baseURL,
		ClientName: clientName,
		Username:   username,
		Password:   password,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	octet := 50 + time.Now().UTC().Nanosecond()%150
	networkIP := fmt.Sprintf("10.252.%d.0", octet)
	hostIP := fmt.Sprintf("10.252.%d.10", octet)

	ctx := context.Background()

	network, err := apiClient.CreateNetwork(ctx, CreateNetworkInput{
		ClientName:  clientName,
		IP:          networkIP,
		Bitmask:     24,
		Description: "Terraform host integration network",
		Site:        "Lon",
		Category:    "test",
		Comment:     "created by host integration test",
		Sync:        false,
	})
	if err != nil {
		t.Fatalf("create network: %v", err)
	}

	t.Cleanup(func() {
		_ = apiClient.DeleteNetwork(context.Background(), DeleteNetworkInput{
			ID:         network.ID,
			ClientName: clientName,
			IP:         networkIP,
			Bitmask:    24,
			IPVersion:  network.IPVersion,
		})
	})

	host, err := apiClient.CreateHost(ctx, CreateHostInput{
		ClientName:  clientName,
		IP:          hostIP,
		Hostname:    "tf-host-create",
		Description: "Terraform host",
		Site:        "Lon",
		Category:    "server",
		Comment:     "created by host integration test",
	})
	if err != nil {
		t.Fatalf("create host: %v", err)
	}

	if host.ID == "" {
		t.Fatal("expected created host to include an id")
	}

	if host.NetworkID != network.ID {
		t.Fatalf("expected host network_id %q, got %q", network.ID, host.NetworkID)
	}

	read, err := apiClient.ReadHost(ctx, clientName, hostIP)
	if err != nil {
		t.Fatalf("read host: %v", err)
	}

	if read.Hostname != "tf-host-create" {
		t.Fatalf("expected hostname %q, got %q", "tf-host-create", read.Hostname)
	}

	updated, err := apiClient.UpdateHost(ctx, UpdateHostInput{
		ID:          host.ID,
		IPInt:       host.IPInt,
		NetworkID:   host.NetworkID,
		ClientName:  clientName,
		IP:          hostIP,
		Hostname:    "tf-host-update",
		Description: "Terraform host updated",
		Site:        "NY",
		Category:    "printer",
		Comment:     "updated by host integration test",
		IPVersion:   host.IPVersion,
	})
	if err != nil {
		t.Fatalf("update host: %v", err)
	}

	if updated.Hostname != "tf-host-update" {
		t.Fatalf("expected updated hostname %q, got %q", "tf-host-update", updated.Hostname)
	}

	if updated.Site != "NY" {
		t.Fatalf("expected updated site %q, got %q", "NY", updated.Site)
	}

	if err := apiClient.DeleteHost(ctx, DeleteHostInput{
		IPInt:      updated.IPInt,
		NetworkID:  updated.NetworkID,
		ClientName: clientName,
		IP:         hostIP,
		IPVersion:  updated.IPVersion,
	}); err != nil {
		t.Fatalf("delete host: %v", err)
	}
}
