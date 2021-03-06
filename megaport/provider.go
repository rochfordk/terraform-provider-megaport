package megaport

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/utilitywarehouse/terraform-provider-megaport/megaport/api"
)

var (
	megaportMutexKV = mutexkv.NewMutexKV()
)

type Config struct {
	Client *api.Client
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"MEGAPORT_TOKEN",
				}, nil),
			},
			"api_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"MEGAPORT_API_ENDPOINT",
				}, api.EndpointProduction),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"megaport_port":        resourceMegaportPort(),
			"megaport_aws_vxc":     resourceMegaportAwsVxc(),
			"megaport_private_vxc": resourceMegaportPrivateVxc(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"megaport_location":     dataSourceMegaportLocation(),
			"megaport_partner_port": dataSourceMegaportPartnerPort(),
			"megaport_port":         dataSourceMegaportPort(),
		},

		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			client := api.NewClient(d.Get("api_endpoint").(string))
			fmt.Printf("initialised megaport api client at %s\n", client.BaseURL)
			if v, ok := d.GetOk("token"); ok { // TODO: is it an error if not found?
				client.Token = v.(string)
			}
			return &Config{
				Client: client,
			}, nil
		},
	}
}
