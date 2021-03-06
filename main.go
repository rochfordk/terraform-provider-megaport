package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/utilitywarehouse/terraform-provider-megaport/megaport"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: megaport.Provider})
}
