package client

import (
	"context"
	"os"
	"testing"
)

func TestClientVLANLifecycleIntegration(t *testing.T) {
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

	ctx := context.Background()

	created, err := apiClient.CreateVLAN(ctx, CreateVLANInput{
		ClientName:  clientName,
		Number:      "321",
		Name:        "tf-vlan-create",
		Description: "Terraform VLAN",
		BGColor:     "blue",
		FontColor:   "white",
	})
	if err != nil {
		t.Fatalf("create vlan: %v", err)
	}

	t.Cleanup(func() {
		_ = apiClient.DeleteVLAN(context.Background(), DeleteVLANInput{
			ID:          created.ID,
			ClientName:  clientName,
			Number:      created.Number,
			Name:        created.Name,
			Description: created.Description,
			BGColor:     created.BGColor,
		})
	})

	if created.ID == "" {
		t.Fatal("expected created vlan to include an id")
	}

	read, err := apiClient.ReadVLAN(ctx, clientName, "321")
	if err != nil {
		t.Fatalf("read vlan: %v", err)
	}

	if read.Name != "tf-vlan-create" {
		t.Fatalf("expected name %q, got %q", "tf-vlan-create", read.Name)
	}

	updated, err := apiClient.UpdateVLAN(ctx, UpdateVLANInput{
		ID:          created.ID,
		ClientName:  clientName,
		Number:      "321",
		Name:        "tf-vlan-update",
		Description: "Terraform VLAN updated",
		BGColor:     "green",
		FontColor:   "black",
	})
	if err != nil {
		t.Fatalf("update vlan: %v", err)
	}

	if updated.Name != "tf-vlan-update" {
		t.Fatalf("expected updated name %q, got %q", "tf-vlan-update", updated.Name)
	}

	if updated.BGColor != "green" {
		t.Fatalf("expected updated bg_color %q, got %q", "green", updated.BGColor)
	}

	if err := apiClient.DeleteVLAN(ctx, DeleteVLANInput{
		ID:          updated.ID,
		ClientName:  clientName,
		Number:      updated.Number,
		Name:        updated.Name,
		Description: updated.Description,
		BGColor:     updated.BGColor,
	}); err != nil {
		t.Fatalf("delete vlan: %v", err)
	}
}
