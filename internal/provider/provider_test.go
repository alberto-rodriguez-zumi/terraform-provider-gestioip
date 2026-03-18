package provider

import (
	"context"
	"testing"

	"github.com/alberto-rodriguez-zumi/terraform-provider-gestioip/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	resp := &provider.MetadataResponse{}

	New().Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "gestioip" {
		t.Fatalf("expected provider type name gestioip, got %q", resp.TypeName)
	}
}

func TestProviderSchemaIncludesClientName(t *testing.T) {
	t.Parallel()

	resp := &provider.SchemaResponse{}

	New().Schema(context.Background(), provider.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["client_name"]
	if !ok {
		t.Fatal("expected client_name attribute to be present in provider schema")
	}

	stringAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected client_name attribute to be a string attribute, got %T", attr)
	}

	if !stringAttr.Optional {
		t.Fatal("expected client_name attribute to be optional")
	}
}

func TestNetworkResourceResolveClientName(t *testing.T) {
	t.Parallel()

	resource := &networkResource{
		client: mustClient(t, "DEFAULT"),
	}

	clientName, ok := resource.resolveClientName(types.StringNull())
	if !ok {
		t.Fatal("expected provider client_name fallback to be used")
	}

	if clientName != "DEFAULT" {
		t.Fatalf("expected DEFAULT, got %q", clientName)
	}
}

func TestNetworkResourceResolveClientNameOverride(t *testing.T) {
	t.Parallel()

	resource := &networkResource{
		client: mustClient(t, "DEFAULT"),
	}

	clientName, ok := resource.resolveClientName(types.StringValue("OTHER"))
	if !ok {
		t.Fatal("expected resource-level client_name to be used")
	}

	if clientName != "OTHER" {
		t.Fatalf("expected OTHER, got %q", clientName)
	}
}

func mustClient(t *testing.T, clientName string) *client.Client {
	t.Helper()

	apiClient, err := client.New(client.Config{
		BaseURL:    "https://gestioip.example.com",
		ClientName: clientName,
		Username:   "admin",
		Password:   "secret",
	})
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}

	return apiClient
}
