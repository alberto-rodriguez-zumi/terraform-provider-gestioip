package main

import (
	"context"
	"log"

	"github.com/alberto-rodriguez-zumi/terraform-provider-gestioip/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/alberto-rodriguez-zumi/gestioip",
	})
	if err != nil {
		log.Fatal(err)
	}
}
