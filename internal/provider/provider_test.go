package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	resp := &provider.MetadataResponse{}

	New().Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "gestioip" {
		t.Fatalf("expected provider type name gestioip, got %q", resp.TypeName)
	}
}
