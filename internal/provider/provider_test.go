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

func TestNetworkDataSourceResolveClientName(t *testing.T) {
	t.Parallel()

	dataSource := &networkDataSource{
		client: mustClient(t, "DEFAULT"),
	}

	clientName, ok := dataSource.resolveClientName(types.StringNull())
	if !ok {
		t.Fatal("expected provider client_name fallback to be used")
	}

	if clientName != "DEFAULT" {
		t.Fatalf("expected DEFAULT, got %q", clientName)
	}
}

func TestNetworkDataSourceResolveClientNameOverride(t *testing.T) {
	t.Parallel()

	dataSource := &networkDataSource{
		client: mustClient(t, "DEFAULT"),
	}

	clientName, ok := dataSource.resolveClientName(types.StringValue("OTHER"))
	if !ok {
		t.Fatal("expected data source-level client_name to be used")
	}

	if clientName != "OTHER" {
		t.Fatalf("expected OTHER, got %q", clientName)
	}
}

func TestFindNetwork(t *testing.T) {
	t.Parallel()

	networks := []client.Network{
		{IP: "192.168.1.0", Bitmask: 24},
		{IP: "10.0.0.0", Bitmask: 8},
	}

	network, found := findNetwork(networks, "10.0.0.0", 8)
	if !found {
		t.Fatal("expected network to be found")
	}

	if network.IP != "10.0.0.0" || network.Bitmask != 8 {
		t.Fatalf("unexpected network returned: %+v", network)
	}
}

func TestParseHostImportID(t *testing.T) {
	t.Parallel()

	clientName, ip, err := parseHostImportID("DEFAULT|10.0.0.10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clientName != "DEFAULT" {
		t.Fatalf("expected DEFAULT, got %q", clientName)
	}

	if ip != "10.0.0.10" {
		t.Fatalf("expected 10.0.0.10, got %q", ip)
	}
}

func TestParseHostImportIDWithoutClientName(t *testing.T) {
	t.Parallel()

	clientName, ip, err := parseHostImportID("10.0.0.10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clientName != "" {
		t.Fatalf("expected empty client name, got %q", clientName)
	}

	if ip != "10.0.0.10" {
		t.Fatalf("expected 10.0.0.10, got %q", ip)
	}
}

func TestParseNetworkImportID(t *testing.T) {
	t.Parallel()

	clientName, ip, bitmask, err := parseNetworkImportID("DEFAULT|10.0.0.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clientName != "DEFAULT" {
		t.Fatalf("expected DEFAULT, got %q", clientName)
	}

	if ip != "10.0.0.0" {
		t.Fatalf("expected 10.0.0.0, got %q", ip)
	}

	if bitmask != 24 {
		t.Fatalf("expected bitmask 24, got %d", bitmask)
	}
}

func TestParseNetworkImportIDSupportsIPv6(t *testing.T) {
	t.Parallel()

	clientName, ip, bitmask, err := parseNetworkImportID("DEFAULT|2001:db8::/64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clientName != "DEFAULT" {
		t.Fatalf("expected DEFAULT, got %q", clientName)
	}

	if ip != "2001:db8::" {
		t.Fatalf("expected 2001:db8::, got %q", ip)
	}

	if bitmask != 64 {
		t.Fatalf("expected bitmask 64, got %d", bitmask)
	}
}

func TestParseVLANImportID(t *testing.T) {
	t.Parallel()

	clientName, number, err := parseVLANImportID("DEFAULT|260")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clientName != "DEFAULT" {
		t.Fatalf("expected DEFAULT, got %q", clientName)
	}

	if number != "260" {
		t.Fatalf("expected 260, got %q", number)
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
