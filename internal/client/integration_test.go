package client

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestClientNetworkLifecycleIntegration(t *testing.T) {
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

	octet := 100 + time.Now().UTC().Nanosecond()%100
	ip := fmt.Sprintf("10.250.%d.0", octet)

	ctx := context.Background()

	created, err := apiClient.CreateNetwork(ctx, CreateNetworkInput{
		ClientName:  clientName,
		IP:          ip,
		Bitmask:     24,
		Description: "Terraform integration network",
		Site:        "Lon",
		Category:    "test",
		Comment:     "created by integration test",
		Sync:        false,
	})
	if err != nil {
		t.Fatalf("create network: %v", err)
	}

	t.Cleanup(func() {
		_ = apiClient.DeleteNetwork(context.Background(), DeleteNetworkInput{
			ID:         created.ID,
			ClientName: clientName,
			IP:         ip,
			Bitmask:    24,
			IPVersion:  created.IPVersion,
		})
	})

	if created.ID == "" {
		t.Fatal("expected created network to include an id")
	}

	read, err := apiClient.ReadNetwork(ctx, clientName, ip, 24)
	if err != nil {
		t.Fatalf("read network: %v", err)
	}

	if read.Description != "Terraform integration network" {
		t.Fatalf("expected description %q, got %q", "Terraform integration network", read.Description)
	}

	updated, err := apiClient.UpdateNetwork(ctx, UpdateNetworkInput{
		ID:          created.ID,
		ClientName:  clientName,
		IP:          ip,
		Bitmask:     24,
		Description: "Terraform integration updated",
		Site:        "NY",
		Category:    "test",
		Comment:     "updated by integration test",
		Sync:        false,
		IPVersion:   created.IPVersion,
	})
	if err != nil {
		t.Fatalf("update network: %v", err)
	}

	if updated.Description != "Terraform integration updated" {
		t.Fatalf("expected updated description %q, got %q", "Terraform integration updated", updated.Description)
	}

	if updated.Site != "NY" {
		t.Fatalf("expected updated site %q, got %q", "NY", updated.Site)
	}

	if err := apiClient.DeleteNetwork(ctx, DeleteNetworkInput{
		ID:         created.ID,
		ClientName: clientName,
		IP:         ip,
		Bitmask:    24,
		IPVersion:  updated.IPVersion,
	}); err != nil {
		t.Fatalf("delete network: %v", err)
	}
}
