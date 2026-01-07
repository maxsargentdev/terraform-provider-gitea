package main

import (
	"context"
	"log"

	"github.com/maxsargendev/terraform-provider-icegitea/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version string = "dev"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "hashicorp.com/maxsargentdev/icegitea",
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
