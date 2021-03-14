package main

import (
	"github.com/hanneshayashi/terraform-provider-gdrive/provider"

	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.Provider,
	})
}
