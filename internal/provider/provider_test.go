package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
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
